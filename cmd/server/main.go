// cmd/server/main.go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"google.golang.org/genai"
	
	"swarmtest/internal/api"
	"swarmtest/internal/gemini"
	"swarmtest/internal/store"
	"swarmtest/internal/models"
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
	eventBus := make(chan models.Event, 1000)
	
	// Database connection
	dbURL := os.Getenv("SUPABASE_DB_URL")
	if dbURL == "" {
		log.Fatal("SUPABASE_DB_URL environment variable is required")
	}

	// Fix for Supavisor prepared statement issues
	if !strings.Contains(dbURL, "default_query_exec_mode") {
		if strings.Contains(dbURL, "?") {
			dbURL += "&default_query_exec_mode=simple_protocol"
		} else {
			dbURL += "?default_query_exec_mode=simple_protocol"
		}
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to Supabase database")

	missionStore := store.NewSupabaseStore(db)
	wsHub := api.NewWebSocketHub(eventBus)
	
	// Start WebSocket hub
	go wsHub.Run(ctx)

	// Initialize Gemini Service
	geminiService := gemini.NewGeminiService(genaiClient)

	// Setup REST API handlers
	restAPI := api.NewRESTAPI(missionStore, geminiService, eventBus)

	// Setup HTTP server
	mux := http.NewServeMux()

	// Register routes
	restAPI.RegisterRoutes(mux)

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
		api.ServeWebSocket(wsHub, w, r)
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
