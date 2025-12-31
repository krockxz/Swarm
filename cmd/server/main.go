// cmd/server/main.go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"google.golang.org/genai"

	"swarmtest/internal/api"
	"swarmtest/internal/gemini"
	"swarmtest/internal/models"
	"swarmtest/internal/services"
	"swarmtest/internal/store"
	"swarmtest/internal/utils"
)

const (
	serverPort      = ":8080"
	eventBusBuffer  = 1000
	readTimeout     = 15 * time.Second
	writeTimeout    = 15 * time.Second
	idleTimeout     = 60 * time.Second
	shutdownTimeout = 10 * time.Second
	
	queryParamKey   = "default_query_exec_mode"
	queryParamValue = "simple_protocol"
)

var (
	version   = "1.0.0"
	buildTime = "unknown"
)

func main() {
	ctx := context.Background()

	// Initialize dependencies
	genaiClient := initGeminiClient(ctx)
	db := initDatabase()
	defer db.Close()

	eventBus := make(chan models.Event, eventBusBuffer)
	browserPool := initBrowserPool()
	if browserPool != nil {
		defer browserPool.Close()
	}

	// Initialize services
	missionStore := store.NewSupabaseStore(db)
	wsHub := api.NewWebSocketHub(eventBus)
	geminiService := gemini.NewGeminiService(genaiClient)
	restAPI := api.NewRESTAPI(missionStore, geminiService, eventBus)

	// Start background services
	go wsHub.Run(ctx)
	go services.NewEventLogger(missionStore, eventBus).Run(ctx)
	log.Println("EventLogger service started")

	// Setup and start HTTP server
	server := setupServer(restAPI, wsHub)
	startServer(server)
}

// initGeminiClient initializes the Gemini AI client
func initGeminiClient(ctx context.Context) *genai.Client {
	apiKey := requireEnv("GEMINI_API_KEY")
	
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatalf("Failed to create Gemini client: %v", err)
	}

	log.Println("Gemini client initialized successfully")
	return client
}

// initDatabase initializes the database connection
func initDatabase() *sql.DB {
	dbURL := requireEnv("SUPABASE_DB_URL")
	dbURL = addQueryParam(dbURL, queryParamKey, queryParamValue)

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Connected to Supabase database")
	return db
}

// initBrowserPool initializes the headless browser pool
func initBrowserPool() *utils.BrowserPool {
	pool, err := utils.NewBrowserPool(true)
	if err != nil {
		log.Printf("Warning: Failed to initialize browser pool: %v. Browser execution mode will be unavailable.", err)
		return nil
	}

	utils.SharedBrowserPool = pool
	log.Println("Browser pool initialized successfully (headless chrome)")
	return pool
}

// setupServer creates and configures the HTTP server
func setupServer(restAPI *api.RESTAPI, wsHub *api.WebSocketHub) *http.Server {
	mux := http.NewServeMux()

	restAPI.RegisterRoutes(mux)
	mux.HandleFunc("/api/health", handleHealth)
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		api.ServeWebSocket(wsHub, w, r)
	})

	return &http.Server{
		Addr:         serverPort,
		Handler:      enableCORS(mux),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}
}

// handleHealth handles the health check endpoint
func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	browserMode := "disabled"
	if utils.SharedBrowserPool != nil {
		browserMode = "enabled"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","version":"%s","build_time":"%s","browser_mode":"%s"}`, 
		version, buildTime, browserMode)
}

// startServer starts the HTTP server with graceful shutdown
func startServer(server *http.Server) {
	go handleShutdown(server)

	logServerInfo()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}

	log.Println("Server stopped")
}

// handleShutdown handles graceful server shutdown
func handleShutdown(server *http.Server) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
}

// logServerInfo logs server startup information
func logServerInfo() {
	browserMode := "disabled"
	if utils.SharedBrowserPool != nil {
		browserMode = "enabled"
	}

	log.Printf("SwarmTest server starting on %s (version %s)", serverPort, version)
	log.Printf("Endpoints:")
	log.Printf("  POST   /api/missions        - Create new mission")
	log.Printf("  GET    /api/missions        - List all missions")
	log.Printf("  GET    /api/missions/{id}   - Get mission status")
	log.Printf("  GET    /api/health          - Health check")
	log.Printf("  GET    /ws                  - WebSocket events")
	log.Printf("  Browser Mode:              %s", browserMode)
}

// enableCORS adds CORS headers to responses
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

// requireEnv gets an environment variable or exits if not set
func requireEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("%s environment variable is required", key)
	}
	return value
}

// addQueryParam adds a query parameter to a URL if not already present
func addQueryParam(url, key, value string) string {
	if strings.Contains(url, key) {
		return url
	}

	separator := "?"
	if strings.Contains(url, "?") {
		separator = "&"
	}

	return url + separator + key + "=" + value
}
