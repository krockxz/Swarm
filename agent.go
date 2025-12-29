package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
)

// Agent represents a single testing agent
type Agent struct {
	id             string
	mission        *Mission
	gemini         GeminiClient
	httpFactory    HTTPClientFactory
	limiter        *RateLimiter
	eventBus       chan<- Event
	parser         *HTMLParser

	// State
	status           string
	currentURL       string
	actionHistory    []string
	urlHistory       []string
	errorCount       int
	successCount     int
	totalLatencyMS   int64
	consecutiveErrors int
	lastActionAt     *time.Time
}

// NewAgent creates a new agent
func NewAgent(id string, mission *Mission, gemini GeminiClient, httpFactory HTTPClientFactory, limiter *RateLimiter, eventBus chan<- Event) *Agent {
	return &Agent{
		id:          id,
		mission:     mission,
		gemini:      gemini,
		httpFactory: httpFactory,
		limiter:     limiter,
		eventBus:    eventBus,
		parser:      NewHTMLParser(),
		status:      "starting",
		currentURL:  mission.TargetURL,
		actionHistory: []string{},
		urlHistory:     []string{mission.TargetURL},
	}
}

// Run starts the agent's main loop
func (a *Agent) Run(ctx context.Context) {
	log.Printf("[Agent %s] Starting mission %s", a.id, a.mission.ID)
	a.status = "running"
	a.emitAgentEvent()

	const maxSteps = 30
	const maxConsecutiveErrors = 3
	const humanLikeDelayMin = 500 * time.Millisecond
	const humanLikeDelayMax = 3000 * time.Millisecond

	client := a.httpFactory()

	for step := 0; step < maxSteps; step++ {
		// Check context
		if ctx.Err() != nil {
			log.Printf("[Agent %s] Context cancelled", a.id)
			a.status = "cancelled"
			a.emitAgentEvent()
			return
		}

		// Check consecutive errors
		if a.consecutiveErrors >= maxConsecutiveErrors {
			log.Printf("[Agent %s] Too many consecutive errors (%d), marking as failed", a.id, a.consecutiveErrors)
			a.status = "failed"
			a.emitAgentEvent()
			return
		}

		// Rate limit
		if err := a.limiter.Wait(ctx); err != nil {
			log.Printf("[Agent %s] Rate limit error: %v", a.id, err)
			a.status = "rate_limited"
			a.emitAgentEvent()
			return
		}

		// Fetch current page
		startTime := time.Now()
		page, err := a.fetchPage(ctx, client, a.currentURL)
		latency := time.Since(startTime)

		if err != nil {
			a.handleError(fmt.Errorf("fetch page: %w", err), latency, "fetch_page")
			continue
		}

		log.Printf("[Agent %s] Step %d: On page %s (%d elements)",
			a.id, step, page.Title, len(page.InteractiveElements))

		// Ask Gemini for next action
		geminiReq := GeminiDecisionRequest{
			SystemPrompt:     a.mission.InitialSystemPrompt,
			MissionGoal:      a.mission.Goal,
			CurrentPage:      *page,
			ActionHistory:    a.actionHistory,
			AvailableActions: []string{"click", "type", "wait", "go_back"},
		}

		decision, err := a.gemini.DecideNextAction(ctx, geminiReq)
		if err != nil {
			a.handleError(fmt.Errorf("gemini decision: %w", err), 0, "gemini_decision")
			continue
		}

		log.Printf("[Agent %s] Decision: %s %s (reasoning: %s)",
			a.id, decision.Action, decision.Selector, decision.Reasoning)

		// Handle go_back action
		if decision.Action == "go_back" {
			if len(a.urlHistory) > 1 {
				a.urlHistory = a.urlHistory[:len(a.urlHistory)-1]
				a.currentURL = a.urlHistory[len(a.urlHistory)-1]
				a.recordAction(fmt.Sprintf("go_back to %s", a.currentURL), decision.Action, "", 0, nil)
			} else {
				log.Printf("[Agent %s] Cannot go back - no history", a.id)
			}
			continue
		}

		// Execute action
		executor, err := NewActionExecutor(client, a.currentURL)
		if err != nil {
			a.handleError(fmt.Errorf("create executor: %w", err), 0, "create_executor")
			continue
		}

		actionStart := time.Now()
		result := executor.ExecuteAction(ctx, decision, a.currentURL)
		actionLatency := time.Since(actionStart)

		if result.Error != nil {
			// Check if it's just a wait action
			if strings.HasPrefix(result.Error.Error(), "wait:") {
				a.recordAction("wait", decision.Action, decision.Selector, actionLatency.Milliseconds(), nil)
			} else {
				a.handleError(fmt.Errorf("execute action: %w", result.Error), actionLatency, decision.Action)
			}
			continue
		}

		// Update state
		a.currentURL = result.NewURL
		a.urlHistory = append(a.urlHistory, a.currentURL)

		actionDesc := fmt.Sprintf("%s %s", decision.Action, decision.Selector)
		if decision.Action == "type" {
			actionDesc = fmt.Sprintf("type '%s' into %s", decision.TextInput, decision.Selector)
		}
		a.recordAction(actionDesc, decision.Action, decision.Selector, actionLatency.Milliseconds(), nil)

		// Human-like delay
		delay := humanLikeDelayMin + time.Duration(rand.Int63n(int64(humanLikeDelayMax-humanLikeDelayMin)))
		log.Printf("[Agent %s] Waiting %v before next action", a.id, delay)
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			a.status = "cancelled"
			a.emitAgentEvent()
			return
		}
	}

	// Completed max steps
	log.Printf("[Agent %s] Completed %d steps, marking as completed", a.id, maxSteps)
	a.status = "completed"
	a.emitAgentEvent()
}

// fetchPage fetches and parses a page
func (a *Agent) fetchPage(ctx context.Context, client *http.Client, url string) (*StrippedPage, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "SwarmTest/1.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return a.parser.ParseHTML(resp.Request.URL.String(), resp.Body)
}

// handleError handles an error
func (a *Agent) handleError(err error, latency time.Duration, action string) {
	a.errorCount++
	a.consecutiveErrors++
	a.totalLatencyMS += latency.Milliseconds()

	log.Printf("[Agent %s] Error: %v", a.id, err)

	a.emitActionLog(ActionLog{
		Timestamp:    time.Now(),
		AgentID:      a.id,
		Action:       action,
		Result:       "error",
		LatencyMS:    latency.Milliseconds(),
		ErrorMessage: err.Error(),
	})
}

// recordAction records a successful action
func (a *Agent) recordAction(description string, action string, selector string, latency int64, err error) {
	a.successCount++
	a.consecutiveErrors = 0
	a.totalLatencyMS += latency
	a.actionHistory = append(a.actionHistory, description)

	now := time.Now()
	a.lastActionAt = &now

	a.emitActionLog(ActionLog{
		Timestamp:  now,
		AgentID:    a.id,
		Action:     action,
		Selector:   selector,
		Result:     "success",
		LatencyMS:  latency,
		NewURL:     a.currentURL,
	})
}

// emitAgentEvent emits an agent status event
func (a *Agent) emitAgentEvent() {
	a.eventBus <- Event{
		Type:      "agent_status",
		Timestamp: time.Now(),
		Data: AgentEvent{
			AgentID:   a.id,
			MissionID: a.mission.ID,
			Status:    a.status,
		},
	}
}

// emitActionLog emits an action log event
func (a *Agent) emitActionLog(log ActionLog) {
	a.eventBus <- Event{
		Type:      "action",
		Timestamp: log.Timestamp,
		Data: AgentEvent{
			AgentID:   a.id,
			MissionID: a.mission.ID,
			ActionLog: &log,
		},
	}
}

// GetMetrics returns current agent metrics
func (a *Agent) GetMetrics() Agent {
	return Agent{
		ID:                a.id,
		MissionID:         a.mission.ID,
		Status:            a.status,
		CurrentURL:        a.currentURL,
		ActionHistory:     append([]string{}, a.actionHistory...),
		ErrorCount:        a.errorCount,
		SuccessCount:      a.successCount,
		TotalLatencyMS:    a.totalLatencyMS,
		ConsecutiveErrors: a.consecutiveErrors,
		URLHistory:        append([]string{}, a.urlHistory...),
		LastActionAt:      a.lastActionAt,
	}
}
