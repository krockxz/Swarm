package services

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"swarmtest/internal/models"
	"swarmtest/internal/store"
)

const (
	flushInterval   = 5 * time.Second
	agentIDSeparator = "-agent-"
)

// EventLogger consumes events from the event bus and persists them to the store.
// It implements the Single Responsibility Principle by handling only event logging.
type EventLogger struct {
	store          store.MissionStore
	eventBus       <-chan models.Event
	mu             sync.RWMutex
	missionMetrics map[string]*missionMetrics
}

type missionMetrics struct {
	totalActions int
	totalErrors  int
	totalLatency int64
	actionCount  int
}

// NewEventLogger creates a new event logger
func NewEventLogger(store store.MissionStore, eventBus <-chan models.Event) *EventLogger {
	return &EventLogger{
		store:          store,
		eventBus:       eventBus,
		missionMetrics: make(map[string]*missionMetrics),
	}
}

// Run starts the event logger's main loop
func (e *EventLogger) Run(ctx context.Context) {
	log.Println("[EventLogger] Starting event logger service")
	
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[EventLogger] Context cancelled, flushing final metrics")
			e.flushAllMetrics()
			return

		case event := <-e.eventBus:
			e.handleEvent(event)

		case <-ticker.C:
			e.flushAllMetrics()
		}
	}
}

// handleEvent processes a single event
func (e *EventLogger) handleEvent(event models.Event) {
	switch event.Type {
	case "action":
		e.handleActionEvent(event)
	case "mission_started":
		e.handleMissionLifecycleEvent(event, true)
	case "mission_completed":
		e.handleMissionLifecycleEvent(event, false)
	}
}

// handleActionEvent processes agent action events
func (e *EventLogger) handleActionEvent(event models.Event) {
	agentEvent, ok := event.Data.(models.AgentEvent)
	if !ok {
		log.Printf("[EventLogger] Invalid action event data: %T", event.Data)
		return
	}

	if agentEvent.ActionLog == nil {
		return
	}
	actionLog := *agentEvent.ActionLog

	missionID := agentEvent.MissionID
	if missionID == "" {
		missionID = extractMissionID(agentEvent.AgentID)
	}

	if missionID == "" {
		log.Printf("[EventLogger] Could not extract mission ID from agent ID: %s", agentEvent.AgentID)
		return
	}

	e.store.AddActionLog(actionLog, missionID)
	e.updateMetrics(missionID, actionLog)
}

// updateMetrics updates in-memory metrics for a mission
func (e *EventLogger) updateMetrics(missionID string, actionLog models.ActionLog) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.missionMetrics[missionID] == nil {
		e.missionMetrics[missionID] = &missionMetrics{}
	}
	
	metrics := e.missionMetrics[missionID]
	if actionLog.Result == "success" {
		metrics.totalActions++
		metrics.totalLatency += actionLog.LatencyMS
		metrics.actionCount++
	} else {
		metrics.totalErrors++
	}
}

// handleMissionLifecycleEvent handles mission start and completion events (DRY principle)
func (e *EventLogger) handleMissionLifecycleEvent(event models.Event, isStart bool) {
	data, ok := event.Data.(map[string]string)
	if !ok {
		return
	}

	missionID, exists := data["mission_id"]
	if !exists {
		return
	}

	if isStart {
		e.mu.Lock()
		e.missionMetrics[missionID] = &missionMetrics{}
		e.mu.Unlock()
		log.Printf("[EventLogger] Initialized metrics for mission %s", missionID)
	} else {
		e.flushMissionMetrics(missionID)
		e.mu.Lock()
		delete(e.missionMetrics, missionID)
		e.mu.Unlock()
		log.Printf("[EventLogger] Flushed final metrics for mission %s", missionID)
	}
}

// flushAllMetrics flushes all mission metrics to the database
func (e *EventLogger) flushAllMetrics() {
	missionIDs := e.getMissionIDs()
	for _, missionID := range missionIDs {
		e.flushMissionMetrics(missionID)
	}
}

// getMissionIDs returns a snapshot of all mission IDs (Law of Demeter)
func (e *EventLogger) getMissionIDs() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	missionIDs := make([]string, 0, len(e.missionMetrics))
	for missionID := range e.missionMetrics {
		missionIDs = append(missionIDs, missionID)
	}
	return missionIDs
}

// flushMissionMetrics flushes metrics for a specific mission
func (e *EventLogger) flushMissionMetrics(missionID string) {
	metrics := e.getMetricsCopy(missionID)
	if metrics == nil {
		return
	}

	mission, exists := e.store.Get(missionID)
	if !exists {
		return
	}

	e.updateMission(mission, metrics)
	e.store.Put(mission)
	e.resetMetrics(missionID)
}

// getMetricsCopy returns a copy of metrics for a mission
func (e *EventLogger) getMetricsCopy(missionID string) *missionMetrics {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	return e.missionMetrics[missionID]
}

// updateMission updates mission with accumulated metrics
func (e *EventLogger) updateMission(mission *models.Mission, metrics *missionMetrics) {
	mission.TotalActions += metrics.totalActions
	mission.TotalErrors += metrics.totalErrors

	if metrics.actionCount > 0 {
		avgLatency := metrics.totalLatency / int64(metrics.actionCount)
		if mission.AverageLatencyMS == 0 {
			mission.AverageLatencyMS = avgLatency
		} else {
			mission.AverageLatencyMS = (mission.AverageLatencyMS + avgLatency) / 2
		}
	}
}

// resetMetrics resets in-memory counters for a mission
func (e *EventLogger) resetMetrics(missionID string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.missionMetrics[missionID] != nil {
		e.missionMetrics[missionID] = &missionMetrics{}
	}
}

// extractMissionID extracts mission ID from agent ID
// Agent ID format: mission-xxx-agent-N
func extractMissionID(agentID string) string {
	idx := strings.LastIndex(agentID, agentIDSeparator)
	if idx == -1 {
		return ""
	}
	return agentID[:idx]
}
