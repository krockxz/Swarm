package main

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketUpgrader upgrades HTTP to WebSocket
var WebSocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for MVP
	},
}

// WebSocketHub manages WebSocket connections and broadcasts events
type WebSocketHub struct {
	connections map[*websocket.Conn]bool
	mu          sync.RWMutex
	eventBus    <-chan Event
	register    chan *websocket.Conn
	unregister  chan *websocket.Conn
}

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub(eventBus <-chan Event) *WebSocketHub {
	return &WebSocketHub{
		connections: make(map[*websocket.Conn]bool),
		eventBus:    eventBus,
		register:    make(chan *websocket.Conn),
		unregister:  make(chan *websocket.Conn),
	}
}

// Run starts the hub's event loop
func (h *WebSocketHub) Run(ctx context.Context) {
	// Start periodic summary broadcaster
	go h.runSummaryBroadcaster(ctx)

	// Process events
	for {
		select {
		case <-ctx.Done():
			log.Println("[WebSocketHub] Context cancelled, shutting down")
			return

		case conn := <-h.register:
			h.mu.Lock()
			h.connections[conn] = true
			h.mu.Unlock()
			log.Printf("[WebSocketHub] Client connected (total: %d)", len(h.connections))

		case conn := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.connections[conn]; ok {
				delete(h.connections, conn)
				conn.Close()
			}
			h.mu.Unlock()
			log.Printf("[WebSocketHub] Client disconnected (total: %d)", len(h.connections))

		case event := <-h.eventBus:
			h.broadcast(event)
		}
	}
}

// broadcast sends an event to all connected clients
func (h *WebSocketHub) broadcast(event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.connections) == 0 {
		return
	}

	// Marshal event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("[WebSocketHub] Failed to marshal event: %v", err)
		return
	}

	// Send to all connections
	for conn := range h.connections {
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("[WebSocketHub] Failed to send to client: %v", err)
			go h.unregister <- conn
		}
	}
}

// runSummaryBroadcaster sends periodic summary events
func (h *WebSocketHub) runSummaryBroadcaster(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.broadcast(Event{
				Type:      "summary_tick",
				Timestamp: time.Now(),
				Data:      map[string]string{"message": "periodic_tick"},
			})
		}
	}
}

// ServeWebSocket handles a new WebSocket connection
func ServeWebSocket(hub *WebSocketHub, w http.ResponseWriter, r *http.Request) {
	conn, err := WebSocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WebSocket] Failed to upgrade connection: %v", err)
		return
	}

	// Register connection
	hub.register <- conn

	// Start read pump to handle disconnects
	go func() {
		defer func() {
			hub.unregister <- conn
		}()

		// Keep reading to detect disconnects
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("[WebSocket] Unexpected close: %v", err)
				}
				break
			}
		}
	}()
}

// MissionSummaryBroadcaster broadcasts mission-specific summaries
type MissionSummaryBroadcaster struct {
	hub      *WebSocketHub
	mission  *Mission
	store    *MissionStore
	stopChan chan struct{}
}

// NewMissionSummaryBroadcaster creates a new mission summary broadcaster
func NewMissionSummaryBroadcaster(hub *WebSocketHub, mission *Mission, store *MissionStore) *MissionSummaryBroadcaster {
	return &MissionSummaryBroadcaster{
		hub:      hub,
		mission:  mission,
		store:    store,
		stopChan: make(chan struct{}),
	}
}

// Start starts broadcasting summaries for this mission
func (b *MissionSummaryBroadcaster) Start() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-b.stopChan:
			return
		case <-ticker.C:
			b.broadcastSummary()
		}
	}
}

// Stop stops broadcasting
func (b *MissionSummaryBroadcaster) Stop() {
	close(b.stopChan)
}

// broadcastSummary broadcasts a summary event
func (b *MissionSummaryBroadcaster) broadcastSummary() {
	mission, ok := b.store.Get(b.mission.ID)
	if !ok {
		return
	}

	// Calculate metrics
	activeAgents := 0
	for _, agent := range mission.AgentMetrics {
		if agent.Status == "running" {
			activeAgents++
		}
	}

	summary := SummaryEvent{
		MissionID:        mission.ID,
		TotalAgents:      mission.NumAgents,
		ActiveAgents:     activeAgents,
		CompletedAgents:  mission.CompletedAgents,
		FailedAgents:     mission.FailedAgents,
		TotalActions:     mission.TotalActions,
		AverageLatencyMS: mission.AverageLatencyMS,
		ErrorRatePercent: calculateErrorRate(mission),
	}

	b.hub.broadcast(Event{
		Type:      "summary",
		Timestamp: time.Now(),
		Data:      summary,
	})
}

// calculateErrorRate calculates the error rate percentage
func calculateErrorRate(mission *Mission) float64 {
	total := mission.TotalActions + mission.TotalErrors
	if total == 0 {
		return 0
	}
	return float64(mission.TotalErrors) / float64(total) * 100
}
