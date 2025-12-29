package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// RESTAPI handles REST API endpoints
type RESTAPI struct {
	store              *MissionStore
	hub                *WebSocketHub
	geminiService      *GeminiService
	rateLimitRegistry  *RateLimiterRegistry
	agentFactory       func(id string, mission *Mission, gemini GeminiClient, httpFactory HTTPClientFactory, limiter *RateLimiter) *Agent
	httpFactory        HTTPClientFactory
}

// NewRESTAPI creates a new REST API handler
func NewRESTAPI(
	store *MissionStore,
	hub *WebSocketHub,
	geminiService *GeminiService,
	rateLimitRegistry *RateLimiterRegistry,
	agentFactory func(id string, mission *Mission, gemini GeminiClient, httpFactory HTTPClientFactory, limiter *RateLimiter) *Agent,
	httpFactory HTTPClientFactory,
) *RESTAPI {
	return &RESTAPI{
		store:             store,
		hub:               hub,
		geminiService:     geminiService,
		rateLimitRegistry: rateLimitRegistry,
		agentFactory:      agentFactory,
		httpFactory:       httpFactory,
	}
}

// CreateMission handles POST /api/missions
func (api *RESTAPI) CreateMission(w http.ResponseWriter, r *http.Request) {
	var req CreateMissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := api.validateCreateMissionRequest(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create mission
	mission := &Mission{
		ID:                  generateMissionID(),
		Name:                req.Name,
		TargetURL:           req.TargetURL,
		NumAgents:           req.NumAgents,
		Goal:                req.Goal,
		MaxDurationSeconds:  req.MaxDurationSeconds,
		RateLimitPerSecond:  req.RateLimitPerSecond,
		InitialSystemPrompt: req.InitialSystemPrompt,
		Status:              "pending",
		CreatedAt:           time.Now(),
		AgentMetrics:        make(map[string]*Agent),
	}

	if mission.InitialSystemPrompt == "" {
		mission.InitialSystemPrompt = defaultSystemPrompt
	}

	// Store mission
	api.store.Put(mission)

	// Start mission immediately
	go api.startMission(mission)

	// Respond
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CreateMissionResponse{MissionID: mission.ID})

	log.Printf("[RESTAPI] Created mission %s: %s (%d agents)", mission.ID, mission.Name, mission.NumAgents)
}

// ListMissions handles GET /api/missions
func (api *RESTAPI) ListMissions(w http.ResponseWriter, r *http.Request) {
	missions := api.store.List()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]*Mission{
		"missions": missions,
	})
}

// GetMission handles GET /api/missions/{id}
func (api *RESTAPI) GetMission(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	id := extractMissionID(r.URL.Path)
	if id == "" {
		http.Error(w, "Invalid mission ID", http.StatusBadRequest)
		return
	}

	mission, ok := api.store.Get(id)
	if !ok {
		http.Error(w, "Mission not found", http.StatusNotFound)
		return
	}

	// Build response
	response := MissionStatusResponse{
		Mission:     mission,
		AgentStates: make([]Agent, 0, len(mission.AgentMetrics)),
	}

	for _, agent := range mission.AgentMetrics {
		response.AgentStates = append(response.AgentStates, *agent)
	}

	// Calculate summary
	activeAgents := 0
	for _, agent := range mission.AgentMetrics {
		if agent.Status == "running" {
			activeAgents++
		}
	}

	response.Summary = &SummaryEvent{
		MissionID:        mission.ID,
		TotalAgents:      mission.NumAgents,
		ActiveAgents:     activeAgents,
		CompletedAgents:  mission.CompletedAgents,
		FailedAgents:     mission.FailedAgents,
		TotalActions:     mission.TotalActions,
		AverageLatencyMS: mission.AverageLatencyMS,
		ErrorRatePercent: calculateErrorRate(mission),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// startMission starts a mission
func (api *RESTAPI) startMission(mission *Mission) {
	log.Printf("[Mission %s] Starting...", mission.ID)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(mission.MaxDurationSeconds)*time.Second)
	defer cancel()

	// Update status
	now := time.Now()
	mission.Status = "running"
	mission.StartedAt = &now

	// Create rate limiter
	limiter := api.rateLimitRegistry.Get(mission.ID, mission.RateLimitPerSecond)

	// Create mission summary broadcaster
	broadcaster := NewMissionSummaryBroadcaster(api.hub, mission, api.store)
	go broadcaster.Start()
	defer broadcaster.Stop()

	// Create agents
	var wg sync.WaitGroup
	agentMutex := sync.Mutex{}

	for i := 0; i < mission.NumAgents; i++ {
		wg.Add(1)
		agentID := fmt.Sprintf("%s-agent-%d", mission.ID, i+1)

		// Create agent
		agent := api.agentFactory(agentID, mission, api.geminiService, api.httpFactory, limiter)

		// Store agent reference
		agentMutex.Lock()
		mission.AgentMetrics[agentID] = agent
		agentMutex.Unlock()

		// Run agent in goroutine
		go func(a *Agent) {
			defer wg.Done()
			a.Run(ctx)

			// Update mission metrics
			agentMutex.Lock()
			metrics := a.GetMetrics()
			mission.AgentMetrics[a.id] = &metrics
			mission.TotalActions += metrics.SuccessCount
			mission.TotalErrors += metrics.ErrorCount

			if metrics.Status == "completed" {
				mission.CompletedAgents++
			} else if metrics.Status == "failed" {
				mission.FailedAgents++
			}

			// Update average latency
			if mission.TotalActions > 0 {
				mission.AverageLatencyMS = (mission.AverageLatencyMS*int64(mission.TotalActions-metrics.SuccessCount) + metrics.TotalLatencyMS) / int64(mission.TotalActions)
			} else {
				mission.AverageLatencyMS = metrics.TotalLatencyMS
			}

			// Update recent events
			if metrics.Status == "completed" || metrics.Status == "failed" {
				event := ActionLog{
					Timestamp: time.Now(),
					AgentID:   a.id,
					Action:    "mission_end",
					Result:    metrics.Status,
				}
				mission.RecentEvents = append(mission.RecentEvents, event)
				if len(mission.RecentEvents) > 10 {
					mission.RecentEvents = mission.RecentEvents[1:]
				}
			}
			agentMutex.Unlock()
		}(agent)
	}

	// Wait for all agents to complete
	wg.Wait()

	// Update mission status
	completedAt := time.Now()
	mission.Status = "completed"
	mission.CompletedAt = &completedAt

	log.Printf("[Mission %s] Completed. Total actions: %d, Errors: %d",
		mission.ID, mission.TotalActions, mission.TotalErrors)
}

// validateCreateMissionRequest validates the create mission request
func (api *RESTAPI) validateCreateMissionRequest(req *CreateMissionRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.TargetURL == "" {
		return fmt.Errorf("target_url is required")
	}
	if req.NumAgents < 1 || req.NumAgents > 1000 {
		return fmt.Errorf("num_agents must be between 1 and 1000")
	}
	if req.Goal == "" {
		return fmt.Errorf("goal is required")
	}
	if req.MaxDurationSeconds < 10 || req.MaxDurationSeconds > 3600 {
		return fmt.Errorf("max_duration_seconds must be between 10 and 3600")
	}
	if req.RateLimitPerSecond <= 0 || req.RateLimitPerSecond > 1000 {
		return fmt.Errorf("rate_limit_per_second must be between 0 and 1000")
	}
	return nil
}

// extractMissionID extracts mission ID from URL path
func extractMissionID(path string) string {
	// Path format: /api/missions/{id}
	parts := splitPath(path)
	if len(parts) >= 3 && parts[1] == "missions" {
		return parts[2]
	}
	return ""
}

// splitPath splits URL path into parts
func splitPath(path string) []string {
	if path == "" {
		return []string{}
	}

	parts := []string{}
	current := ""

	for _, ch := range path {
		if ch == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

// generateMissionID generates a unique mission ID
func generateMissionID() string {
	return "mission-" + uuid.New().String()[:8]
}

// defaultSystemPrompt is the default system prompt for agents
const defaultSystemPrompt = `You are an autonomous web testing agent. Navigate websites efficiently to achieve the given goal.

Available actions:
- click: Click on buttons, links, or interactive elements
- type: Fill in input fields and submit forms
- wait: Pause and observe the page
- go_back: Navigate to the previous page

Be strategic and deliberate. Each action should bring you closer to the goal.`
