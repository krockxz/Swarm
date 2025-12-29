// main.go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"google.golang.org/genai"
)

var (
	version   = "1.0.0"
	buildTime = "unknown"
)

func main() {
	// Validate environment variables
	if os.Getenv("GEMINI_API_KEY") == "" {
		log.Fatal("GEMINI_API_KEY environment variable is required")
	}

	// Initialize context
	ctx := context.Background()

	// Initialize Gemini client
	genaiClient, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  os.Getenv("GEMINI_API_KEY"),
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatalf("Failed to create Gemini client: %v", err)
	}

	log.Println("Gemini client initialized successfully")

	// Initialize shared state
	eventBus := make(chan Event, 1000)
	missionStore := NewMissionStore()
	wsHub := NewWebSocketHub(eventBus)
	rateLimiterRegistry := NewRateLimiterRegistry()

	// Start WebSocket hub
	go wsHub.Run(ctx)

	// Initialize components
	geminiService := NewGeminiService(genaiClient)
	httpClientFactory := NewHTTPClientFactory
	agentFactory := func(id string, mission *Mission, gemini GeminiClient, httpFactory HTTPClientFactory, limiter *RateLimiter) *RuntimeAgent {
		return NewAgent(id, mission, gemini, httpFactory, limiter, eventBus)
	}

	// Setup REST API handlers
	restAPI := NewRESTAPI(missionStore, wsHub, geminiService, rateLimiterRegistry, agentFactory, httpClientFactory)

	// Setup HTTP server
	mux := http.NewServeMux()

	// REST API routes
	mux.HandleFunc("/api/missions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			restAPI.CreateMission(w, r)
		case http.MethodGet:
			restAPI.ListMissions(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/missions/", restAPI.GetMission)

	// Health check
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","version":"%s","build_time":"%s"}`, version, buildTime)
	})

	// WebSocket endpoint
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWebSocket(wsHub, w, r)
	})

	// Create server with timeouts
	server := &http.Server{
		Addr:         ":8080",
		Handler:      enableCORS(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	log.Printf("SwarmTest server starting on :8080 (version %s)", version)
	log.Printf("Endpoints:")
	log.Printf("  POST   /api/missions        - Create new mission")
	log.Printf("  GET    /api/missions        - List all missions")
	log.Printf("  GET    /api/missions/{id}   - Get mission status")
	log.Printf("  GET    /api/health          - Health check")
	log.Printf("  GET    /ws                  - WebSocket events")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}

// MissionStore stores missions in memory
type MissionStore struct {
	mu       sync.RWMutex
	missions map[string]*Mission
}

func NewMissionStore() *MissionStore {
	return &MissionStore{
		missions: make(map[string]*Mission),
	}
}

func (s *MissionStore) Put(mission *Mission) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.missions[mission.ID] = mission
}

func (s *MissionStore) Get(id string) (*Mission, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	mission, ok := s.missions[id]
	return mission, ok
}

func (s *MissionStore) List() []*Mission {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Mission, 0, len(s.missions))
	for _, m := range s.missions {
		result = append(result, m)
	}
	return result
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
