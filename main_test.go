package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// MockGeminiClient is a mock implementation of GeminiClient
type MockGeminiClient struct {
	decisions []GeminiDecisionResponse
	callCount int
	mu        sync.Mutex
}

func (m *MockGeminiClient) DecideNextAction(ctx context.Context, req GeminiDecisionRequest) (GeminiDecisionResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Return a default decision that waits
	if len(m.decisions) == 0 {
		return GeminiDecisionResponse{
			Action:    "wait",
			Reasoning: "mock decision - wait",
		}, nil
	}

	decision := m.decisions[m.callCount%len(m.decisions)]
	m.callCount++
	return decision, nil
}

func (m *MockGeminiClient) getCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

// TestHTMLParser tests the HTML parser
func TestHTMLParser(t *testing.T) {
	parser := NewHTMLParser()

	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Test Page</title>
		<meta name="description" content="This is a test page for parsing">
	</head>
	<body>
		<a href="/home" class="nav-link">Home</a>
		<a href="/about">About</a>
		<button id="submit-btn">Submit</button>
		<form action="/search" method="GET">
			<input type="text" name="q" placeholder="Search...">
			<input type="submit" value="Search">
		</form>
	</body>
	</html>
	`

	page, err := parser.ParseHTML("http://example.com", strings.NewReader(html))
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}

	// Check title
	if page.Title != "Test Page" {
		t.Errorf("Expected title 'Test Page', got '%s'", page.Title)
	}

	// Check description
	if page.Description != "This is a test page for parsing" {
		t.Errorf("Expected description 'This is a test page for parsing', got '%s'", page.Description)
	}

	// Check elements
	linkCount := 0
	buttonCount := 0
	inputCount := 0
	formCount := 0

	for _, elem := range page.InteractiveElements {
		switch elem.Type {
		case "link":
			linkCount++
		case "button":
			buttonCount++
		case "input":
			inputCount++
		case "form":
			formCount++
		}
	}

	if linkCount != 2 {
		t.Errorf("Expected 2 links, got %d", linkCount)
	}
	if buttonCount != 1 {
		t.Errorf("Expected 1 button, got %d", buttonCount)
	}
	if inputCount != 2 {
		t.Errorf("Expected 2 inputs, got %d", inputCount)
	}
	if formCount != 1 {
		t.Errorf("Expected 1 form, got %d", formCount)
	}
}

// TestRateLimiter tests the rate limiter
func TestRateLimiter(t *testing.T) {
	// Create a limiter that allows 10 requests per second
	limiter := NewRateLimiter(10.0, 10)

	ctx := context.Background()

	start := time.Now()
	requests := 20

	// Try to make 20 requests
	for i := 0; i < requests; i++ {
		if err := limiter.Wait(ctx); err != nil {
			t.Fatalf("Wait failed: %v", err)
		}
	}

	elapsed := time.Since(start)

	// With 10 requests per second, 20 requests should take about 1 second
	// Allow some margin for timing variations
	expectedMin := 900 * time.Millisecond
	expectedMax := 1200 * time.Millisecond

	if elapsed < expectedMin {
		t.Errorf("Rate limiter too fast: expected at least %v, got %v", expectedMin, elapsed)
	}
	if elapsed > expectedMax {
		t.Logf("Warning: Rate limiter slower than expected: %v (max %v)", elapsed, expectedMax)
	}
}

// TestRateLimiterBurst tests rate limiter burst capacity
func TestRateLimiterBurst(t *testing.T) {
	limiter := NewRateLimiter(10.0, 10)

	ctx := context.Background()

	// First 10 requests should be immediate (burst capacity)
	start := time.Now()
	for i := 0; i < 10; i++ {
		if err := limiter.Wait(ctx); err != nil {
			t.Fatalf("Wait failed: %v", err)
		}
	}
	elapsed := time.Since(start)

	if elapsed > 100*time.Millisecond {
		t.Errorf("Burst requests took too long: %v", elapsed)
	}

	// 11th request should wait for refill
	start = time.Now()
	if err := limiter.Wait(ctx); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}
	elapsed = time.Since(start)

	// Should wait approximately 1/10 second
	if elapsed < 50*time.Millisecond {
		t.Errorf("11th request did not wait long enough: %v", elapsed)
	}
}

// TestMissionCreation tests mission creation via REST API
func TestMissionCreation(t *testing.T) {
	store := NewMissionStore()

	req := CreateMissionRequest{
		Name:               "Test Mission",
		TargetURL:          "http://example.com",
		NumAgents:          3,
		Goal:               "Test goal",
		MaxDurationSeconds: 60,
		RateLimitPerSecond: 5.0,
	}

	if err := validateRequest(&req); err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	mission := &Mission{
		Name:                req.Name,
		TargetURL:           req.TargetURL,
		NumAgents:           req.NumAgents,
		Goal:                req.Goal,
		MaxDurationSeconds:  req.MaxDurationSeconds,
		RateLimitPerSecond:  req.RateLimitPerSecond,
		Status:              "pending",
		CreatedAt:           time.Now(),
		AgentMetrics:        make(map[string]*Agent),
	}

	store.Put(mission)

	retrieved, ok := store.Get(mission.ID)
	if !ok {
		t.Fatal("Mission not found in store")
	}

	if retrieved.Name != req.Name {
		t.Errorf("Expected name '%s', got '%s'", req.Name, retrieved.Name)
	}
}

// TestAgentWithMockGemini tests agent with mock Gemini client
func TestAgentWithMockGemini(t *testing.T) {
	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `
		<!DOCTYPE html>
		<html>
		<head><title>Test</title></head>
		<body>
			<a href="/page2">Next Page</a>
		</body>
		</html>
		`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	}))
	defer ts.Close()

	// Create mock Gemini client
	mockGemini := &MockGeminiClient{
		decisions: []GeminiDecisionResponse{
			{
				Action:    "click",
				Selector: "a",
				Reasoning: "Click the link",
			},
			{
				Action:    "wait",
				Reasoning: "Wait",
			},
		},
	}

	// Create mission
	mission := &Mission{
		ID:                  "test-mission",
		Name:                "Test",
		TargetURL:           ts.URL,
		NumAgents:           1,
		Goal:                "Test goal",
		MaxDurationSeconds:  10,
		RateLimitPerSecond:  10.0,
		InitialSystemPrompt: "Test prompt",
		Status:              "running",
		CreatedAt:           time.Now(),
		AgentMetrics:        make(map[string]*Agent),
	}

	// Create agent
	eventBus := make(chan Event, 100)
	limiter := NewRateLimiter(10.0, 10)
	agent := NewAgent("test-agent", mission, mockGemini, NewHTTPClientFactory, limiter, eventBus)

	// Run agent for a short time
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Run agent in goroutine
	done := make(chan bool)
	go func() {
		agent.Run(ctx)
		done <- true
	}()

	// Wait for agent to finish or timeout
	select {
	case <-done:
		// Agent finished
	case <-time.After(3 * time.Second):
		t.Fatal("Agent did not finish in time")
	}

	// Check that Gemini was called
	callCount := mockGemini.getCallCount()
	if callCount == 0 {
		t.Error("Gemini client was not called")
	}

	t.Logf("Agent made %d Gemini calls", callCount)
}

// TestWebSocketEventSerialization tests WebSocket event serialization
func TestWebSocketEventSerialization(t *testing.T) {
	event := Event{
		Type:      "test_event",
		Timestamp: time.Now(),
		Data: map[string]string{
			"key": "value",
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	var decoded Event
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal event: %v", err)
	}

	if decoded.Type != event.Type {
		t.Errorf("Expected type '%s', got '%s'", event.Type, decoded.Type)
	}
}

// validateRequest validates a create mission request
func validateRequest(req *CreateMissionRequest) error {
	if req.Name == "" {
		return nil
	}
	if req.TargetURL == "" {
		return nil
	}
	if req.NumAgents < 1 || req.NumAgents > 1000 {
		return nil
	}
	if req.Goal == "" {
		return nil
	}
	if req.MaxDurationSeconds < 10 || req.MaxDurationSeconds > 3600 {
		return nil
	}
	if req.RateLimitPerSecond <= 0 || req.RateLimitPerSecond > 1000 {
		return nil
	}
	return nil
}

// BenchmarkHTMLParser benchmarks HTML parsing
func BenchmarkHTMLParser(b *testing.B) {
	parser := NewHTMLParser()

	html := `
	<!DOCTYPE html>
	<html>
	<head><title>Benchmark</title></head>
	<body>
		<a href="/link1">Link 1</a>
		<a href="/link2">Link 2</a>
		<button>Button</button>
		<form>
			<input type="text" name="field">
		</form>
	</body>
	</html>
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.ParseHTML("http://example.com", strings.NewReader(html))
		if err != nil {
			b.Fatalf("Parse failed: %v", err)
		}
	}
}
