package agent

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"swarmtest/internal/models"
	"swarmtest/internal/gemini"
	"swarmtest/internal/utils"
)

// RuntimeAgent represents a running agent
type RuntimeAgent struct {
	id          string
	mission     *models.Mission
	gemini      gemini.GeminiClient
	httpFactory utils.HTTPClientFactory
	limiter     *utils.RateLimiter
	eventBus    chan<- models.Event

	// Browser mode support
	browserExecutor *utils.BrowserExecutor
	isBrowserMode   bool

	// State
	status        string
	currentURL    string
	actionHistory []string
	urlHistory    []string
	errorCount    int
	successCount  int
	totalLatency  time.Duration
	consecutiveErrors int
	lastActionAt  time.Time
}

// NewAgent creates a new agent
func NewAgent(
	id string,
	mission *models.Mission,
	gemini gemini.GeminiClient,
	httpFactory utils.HTTPClientFactory,
	limiter *utils.RateLimiter,
	eventBus chan<- models.Event,
	browserExecutor *utils.BrowserExecutor,
) *RuntimeAgent {
	isBrowserMode := mission.ExecutionMode == models.ExecutionModeBrowser

	return &RuntimeAgent{
		id:               id,
		mission:          mission,
		gemini:           gemini,
		httpFactory:      httpFactory,
		limiter:          limiter,
		eventBus:         eventBus,
		browserExecutor:  browserExecutor,
		isBrowserMode:    isBrowserMode,
		status:           "initialized",
		currentURL:       mission.TargetURL,
		actionHistory:    make([]string, 0),
		urlHistory:       make([]string, 0),
	}
}

// Run starts the agent loop
func (a *RuntimeAgent) Run(ctx context.Context) {
	log.Printf("[Agent %s] Starting mission: %s (mode: %s)", a.id, a.mission.Goal, a.mission.ExecutionMode)
	a.status = "running"
	a.urlHistory = append(a.urlHistory, a.currentURL)

	// Create HTTP client (always needed for fallback or mixed mode potentially)
	client := a.httpFactory()
	
	// Create executor (only for HTTP mode)
	var httpExecutor *utils.ActionExecutor
	var err error
	
	if !a.isBrowserMode {
		httpExecutor, err = utils.NewActionExecutor(client, a.currentURL)
		if err != nil {
			a.handleError(err, "init_executor")
			a.status = "failed"
			return
		}
	} else {
		// Browser mode: ensure we have an executor
		if a.browserExecutor == nil {
			a.handleError(fmt.Errorf("browser executor is nil"), "init_browser")
			a.status = "failed"
			return
		}
		
		// Initial navigation
		result := a.browserExecutor.ExecuteAction(ctx, models.GeminiDecisionResponse{Action: "visit"}, a.currentURL)
		if result.Error != nil {
			a.handleError(result.Error, "initial_visit")
			// Try to continue?
		} else {
			log.Printf("[Agent %s] Initial visit successful", a.id)
		}
	}

	// Clean up browser tab on exit (AFTER the loop)
	if a.browserExecutor != nil {
		defer a.browserExecutor.Close()
	}

	for {
		select {
		case <-ctx.Done():
			log.Printf("[Agent %s] Context done, stopping", a.id)
			a.status = "stopped"
			return
		default:
			// 1. Rate Limiting
			if err := a.limiter.Wait(ctx); err != nil {
				a.status = "stopped"
				return
			}

			startTime := time.Now()
			
			var page *models.StrippedPage
			
			if a.isBrowserMode {
				// Get current state from browser
				htmlContent, urlStr, err := a.browserExecutor.CaptureDOM(ctx)
				if err != nil {
					a.handleError(err, "fetch_page_browser")
					continue
				}
				
				// Update current URL if changed
				if urlStr != "" && urlStr != a.currentURL {
					a.currentURL = urlStr
					a.urlHistory = append(a.urlHistory, a.currentURL)
				}
				
				parser := utils.NewHTMLParser()
				page, err = parser.ParseHTMLString(a.currentURL, htmlContent)
				if err != nil {
					a.handleError(err, "parse_page")
					continue
				}
			} else {
				// HTTP Mode
				req, _ := http.NewRequestWithContext(ctx, "GET", a.currentURL, nil)
				req.Header.Set("User-Agent", "SwarmTest/1.0")
				resp, err := client.Do(req)
				if err != nil {
					a.handleError(err, "fetch_page")
					continue
				}
				defer resp.Body.Close()
				
				parser := utils.NewHTMLParser()
				page, err = parser.ParseHTML(a.currentURL, resp.Body)
				if err != nil {
					a.handleError(err, "parse_page")
					continue
				}
			}
			
			// 3. Ask Gemini
			decision, err := a.gemini.DecideNextAction(ctx, a.mission, a.GetSnapshot(), page)
			if err != nil {
				a.handleError(err, "gemini_decision")
				continue
			}

			// Handle terminal actions immediately
			if decision.Action == "completed" {
				a.recordAction(*decision, 0, "") 
				a.status = "completed"
				return
			}
			if decision.Action == "failed" {
				a.recordAction(*decision, 0, "") 
				a.status = "failed"
				return
			}
			
			// 4. Execute Action
			var result utils.ExecuteActionResult
			
			if a.isBrowserMode {
				result = a.browserExecutor.ExecuteAction(ctx, *decision, a.currentURL)
			} else {
				result = httpExecutor.ExecuteAction(ctx, *decision, a.currentURL)
			}
			
			latency := time.Since(startTime)
			a.totalLatency += latency
			a.lastActionAt = time.Now()

			if result.Error != nil {
				a.handleError(result.Error, decision.Action)
			} else {
				a.recordAction(*decision, latency.Milliseconds(), result.NewURL)
				
				// Update URL if changed
				if result.NewURL != "" && result.NewURL != a.currentURL {
					a.currentURL = result.NewURL
					a.urlHistory = append(a.urlHistory, a.currentURL)
					
					// Update executor base URL by recreating it (HTTP only)
					if !a.isBrowserMode {
						httpExecutor, _ = utils.NewActionExecutor(client, a.currentURL)
					}
				}
				
				if decision.Action == "completed" {
					a.status = "completed"
					return
				}
			}
		}
	}
}

// executeHTTPAction executes an action using HTTP client
func (a *RuntimeAgent) executeHTTPAction(decision *models.GeminiDecisionResponse, executor *utils.ActionExecutor, client *http.Client, startTime time.Time) {
	result := executor.ExecuteAction(context.Background(), *decision, a.currentURL)

	latency := time.Since(startTime)

	if result.Error != nil {
		a.handleError(result.Error, decision.Action)
	} else {
		a.recordAction(*decision, latency.Milliseconds(), result.NewURL)

		// Update URL if changed
		if result.NewURL != "" && result.NewURL != a.currentURL {
			a.currentURL = result.NewURL
			a.urlHistory = append(a.urlHistory, a.currentURL)

			// Update executor base URL by recreating it
			executor, _ = utils.NewActionExecutor(client, a.currentURL)
		}

		if decision.Action == "completed" {
			a.status = "completed"
		}
	}
}

// executeBrowserAction executes an action using the browser
func (a *RuntimeAgent) executeBrowserAction(decision *models.GeminiDecisionResponse, startTime time.Time) {
	result := a.browserExecutor.ExecuteAction(context.Background(), *decision, a.currentURL)

	latency := time.Since(startTime)

	if result.Error != nil {
		a.handleError(result.Error, decision.Action)
	} else {
		a.recordAction(*decision, latency.Milliseconds(), result.NewURL)

		// Update URL if changed
		if result.NewURL != "" && result.NewURL != a.currentURL {
			a.currentURL = result.NewURL
			a.urlHistory = append(a.urlHistory, a.currentURL)
		}

		if decision.Action == "completed" {
			a.status = "completed"
		}
	}
}

// handleError handles an error
func (a *RuntimeAgent) handleError(err error, action string) {
	a.errorCount++
	a.consecutiveErrors++
	log.Printf("[Agent %s] Error during %s: %v", a.id, action, err)

	a.emitEvent(models.ActionLog{
		Timestamp:    time.Now(),
		AgentID:      a.id,
		Action:       action,
		Result:       "failed",
		ErrorMessage: err.Error(),
		LatencyMS:    0,
	})
	
	// Backoff
	time.Sleep(time.Duration(a.consecutiveErrors) * time.Second)
	
	if a.consecutiveErrors > 10 {
		a.status = "failed"
	}
}

// recordAction records a successful action
func (a *RuntimeAgent) recordAction(decision models.GeminiDecisionResponse, latencyMS int64, newURL string) {
	a.successCount++
	a.consecutiveErrors = 0
	
	actionDesc := decision.Action
	if decision.Selector != "" {
		actionDesc += fmt.Sprintf(" %s", decision.Selector)
	}
	a.actionHistory = append(a.actionHistory, actionDesc)
	
	a.emitEvent(models.ActionLog{
		Timestamp: time.Now(),
		AgentID:   a.id,
		Action:    decision.Action,
		Selector:  decision.Selector,
		Result:    "success",
		LatencyMS: latencyMS,
		NewURL:    newURL,
	})
}

// emitEvent sends an event to the bus
func (a *RuntimeAgent) emitEvent(logEntry models.ActionLog) {
	select {
	case a.eventBus <- models.Event{
		Type:      "agent_action",
		Timestamp: time.Now(),
		Data:      logEntry,
	}:
	default:
		// Drop event if bus is full
	}
}

// GetMetrics returns current metrics
func (a *RuntimeAgent) GetMetrics() models.Agent {
	return models.Agent{
		ID:                a.id,
		MissionID:         a.mission.ID,
		Status:            a.status,
		CurrentURL:        a.currentURL,
		ActionHistory:     a.actionHistory,
		ErrorCount:        a.errorCount,
		SuccessCount:      a.successCount,
		TotalLatencyMS:    a.totalLatency.Milliseconds(),
		ConsecutiveErrors: a.consecutiveErrors,
		URLHistory:        a.urlHistory,
		LastActionAt:      &a.lastActionAt,
	}
}

// GetSnapshot returns a model snapshot of the runtime agent
func (a *RuntimeAgent) GetSnapshot() *models.Agent {
    // Map RuntimeAgent state to models.Agent DTO
    metrics := a.GetMetrics()
    return &metrics
}
