package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"swarmtest/internal/agent"
	"swarmtest/internal/gemini"
	"swarmtest/internal/models"
	"swarmtest/internal/store"
	"swarmtest/internal/utils"
)

// RESTAPI handles REST endpoints
type RESTAPI struct {
	store      store.MissionStore
	gemini     gemini.GeminiClient
	eventBus   chan models.Event
	rateLimits *utils.RateLimiterRegistry
}

// NewRESTAPI creates a new REST API handler
func NewRESTAPI(store store.MissionStore, gemini gemini.GeminiClient, eventBus chan models.Event) *RESTAPI {
	return &RESTAPI{
		store:      store,
		gemini:     gemini,
		eventBus:   eventBus,
		rateLimits: utils.NewRateLimiterRegistry(),
	}
}

// RegisterRoutes registers routes
func (api *RESTAPI) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/missions", api.handleMissions)
	mux.HandleFunc("/api/missions/", api.handleMissionDetail)
}

func (api *RESTAPI) handleMissions(w http.ResponseWriter, r *http.Request) {
	// CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method == "POST" {
		api.createMission(w, r)
		return
	}

	if r.Method == "GET" {
		api.listMissions(w, r)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (api *RESTAPI) handleMissionDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	if r.Method == "GET" {
		id := extractMissionID(r.URL.Path)
		if id == "" {
			http.Error(w, "Invalid mission ID", http.StatusBadRequest)
			return
		}

		mission, exists := api.store.Get(id)
		if !exists {
			http.Error(w, "Mission not found", http.StatusNotFound)
			return
		}

		// Calculate summary
		activeAgents := 0
		for _, agent := range mission.AgentMetrics {
			if agent.Status == "running" {
				activeAgents++
			}
		}

		total := mission.TotalActions + mission.TotalErrors
		errorRate := 0.0
		if total > 0 {
			errorRate = float64(mission.TotalErrors) / float64(total) * 100
		}

		agentStates := make([]models.Agent, 0, len(mission.AgentMetrics))
		for _, a := range mission.AgentMetrics {
			agentStates = append(agentStates, *a)
		}

		resp := models.MissionStatusResponse{
			Mission: mission,
			AgentStates: agentStates,
			Summary: &models.SummaryEvent{
				MissionID:        mission.ID,
				TotalAgents:      mission.NumAgents,
				ActiveAgents:     activeAgents,
				CompletedAgents:  mission.CompletedAgents,
				FailedAgents:     mission.FailedAgents,
				TotalActions:     mission.TotalActions,
				TotalErrors:      mission.TotalErrors,
				AverageLatencyMS: mission.AverageLatencyMS,
				ErrorRatePercent: errorRate,
			},
		}

		json.NewEncoder(w).Encode(resp)
		return
	}
}

func (api *RESTAPI) createMission(w http.ResponseWriter, r *http.Request) {
	var req models.CreateMissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	missionID := generateMissionID()
	now := time.Now()

	mission := &models.Mission{
		ID:                 missionID,
		Name:               req.Name,
		TargetURL:          req.TargetURL,
		NumAgents:          req.NumAgents,
		Goal:               req.Goal,
		MaxDurationSeconds: req.MaxDurationSeconds,
		RateLimitPerSecond: req.RateLimitPerSecond,
		InitialSystemPrompt: req.InitialSystemPrompt,
		Status:             "created",
		CreatedAt:          now,
		AgentMetrics:       make(map[string]*models.Agent),
		RecentEvents:       []models.ActionLog{},
	}

	api.store.Put(mission)

	// Start mission asynchronously
	go api.startMission(mission)

	json.NewEncoder(w).Encode(models.CreateMissionResponse{
		MissionID: missionID,
	})
}

func (api *RESTAPI) listMissions(w http.ResponseWriter, r *http.Request) {
	missions := api.store.List()
	json.NewEncoder(w).Encode(map[string][]*models.Mission{
		"missions": missions,
	})
}

func (api *RESTAPI) startMission(mission *models.Mission) {
	log.Printf("Starting mission %s with %d agents", mission.ID, mission.NumAgents)
	
	mission.Status = "running"
	now := time.Now()
	mission.StartedAt = &now
	api.store.Put(mission)

	// Create rate limiter
	limiter := api.rateLimits.Get(mission.ID, mission.RateLimitPerSecond)
	
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(mission.MaxDurationSeconds)*time.Second)
	defer cancel()

	for i := 0; i < mission.NumAgents; i++ {
		agentID := fmt.Sprintf("%s-agent-%d", mission.ID, i)
		runtimeAgent := agent.NewAgent(
			agentID,
			mission,
			api.gemini,
			utils.NewHTTPClientFactory,
			limiter,
			api.eventBus, 
		)
		
		// Initialize agent metric in mission
		mission.AgentMetrics[agentID] = &models.Agent{
			ID:        agentID,
			MissionID: mission.ID,
			Status:    "initialized",
		}
		
		go func(a *agent.RuntimeAgent) {
			a.Run(ctx)
		}(runtimeAgent)
	}

	// Wait for context done (mission timeout) or completion logic?
	// For now, we just let agents run until timeout.
	
	<-ctx.Done()
	
	log.Printf("Mission %s finished (timeout or completed)", mission.ID)
	mission.Status = "completed"
	completedAt := time.Now()
	mission.CompletedAt = &completedAt
	
	// Final save
	api.store.Put(mission)
	
	// Clean up rate limiter
	api.rateLimits.Remove(mission.ID)
}

// extractMissionID extracts mission ID from URL path
func extractMissionID(path string) string {
	// Path format: /api/missions/{id}
	// Trim prefix /api/missions/
	// This is cleaner than splitting
	prefix := "/api/missions/"
	if len(path) > len(prefix) && path[:len(prefix)] == prefix {
		return path[len(prefix):]
	}
	return ""
}

// generateMissionID generates a unique mission ID
func generateMissionID() string {
	return "mission-" + uuid.New().String()[:8]
}
