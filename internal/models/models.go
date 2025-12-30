package models

import (
	"time"
)

// Mission represents a test mission configuration
type Mission struct {
	ID                   string    `json:"id"`
	Name                 string    `json:"name"`
	TargetURL            string    `json:"target_url"`
	NumAgents            int       `json:"num_agents"`
	Goal                 string    `json:"goal"`
	MaxDurationSeconds   int       `json:"max_duration_seconds"`
	RateLimitPerSecond   float64   `json:"rate_limit_per_second"`
	InitialSystemPrompt  string    `json:"initial_system_prompt"`
	Status               string    `json:"status"`
	CreatedAt            time.Time `json:"created_at"`
	StartedAt            *time.Time `json:"started_at,omitempty"`
	CompletedAt          *time.Time `json:"completed_at,omitempty"`

	// Runtime metrics
	TotalActions        int                `json:"total_actions"`
	TotalErrors         int                `json:"total_errors"`
	AverageLatencyMS    int64              `json:"average_latency_ms"`
	CompletedAgents     int                `json:"completed_agents"`
	FailedAgents        int                `json:"failed_agents"`
	RecentEvents        []ActionLog        `json:"recent_events"`
	AgentMetrics        map[string]*Agent  `json:"agent_metrics"`
}

// Agent represents a single testing agent
type Agent struct {
	ID              string         `json:"id"`
	MissionID       string         `json:"mission_id"`
	Status          string         `json:"status"`
	CurrentURL      string         `json:"current_url"`
	ActionHistory   []string       `json:"action_history"`
	ErrorCount      int            `json:"error_count"`
	SuccessCount    int            `json:"success_count"`
	TotalLatencyMS  int64          `json:"total_latency_ms"`
	ConsecutiveErrors int          `json:"consecutive_errors"`
	URLHistory      []string       `json:"url_history"`
	LastActionAt    *time.Time     `json:"last_action_at,omitempty"`
}

// ActionLog represents a single action performed by an agent
type ActionLog struct {
	Timestamp     time.Time `json:"timestamp"`
	AgentID       string    `json:"agent_id"`
	Action        string    `json:"action"`
	Selector      string    `json:"selector,omitempty"`
	Result        string    `json:"result"`
	LatencyMS     int64     `json:"latency_ms"`
	ErrorMessage  string    `json:"error_message,omitempty"`
	NewURL        string    `json:"new_url,omitempty"`
}

// StrippedPage represents a simplified view of a web page
type StrippedPage struct {
	URL                  string    `json:"url"`
	Title                string    `json:"title"`
	Description          string    `json:"description"`
	TextContent          string    `json:"text_content"`
	InteractiveElements  []Element `json:"interactive_elements"`
	Timestamp            time.Time `json:"timestamp"`
}

// Element represents an interactive element on a page
type Element struct {
	ID          string `json:"id,omitempty"`
	Type        string `json:"type"` // button, link, input, form
	Text        string `json:"text,omitempty"`
	Selector    string `json:"selector"`
	Href        string `json:"href,omitempty"`
	Name        string `json:"name,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
	InputType   string `json:"input_type,omitempty"`
}

// GeminiDecisionRequest is the request sent to Gemini for action decision
type GeminiDecisionRequest struct {
	SystemPrompt     string        `json:"system_prompt"`
	MissionGoal      string        `json:"mission_goal"`
	CurrentPage      StrippedPage  `json:"current_page"`
	ActionHistory    []string      `json:"action_history"`
	AvailableActions []string      `json:"available_actions"`
}

// GeminiDecisionResponse is the response from Gemini
type GeminiDecisionResponse struct {
	Reasoning          string `json:"reasoning"`
	Action             string `json:"action"` // click, type, wait, go_back
	Selector           string `json:"selector,omitempty"`
	TextInput          string `json:"text_input,omitempty"`
	ExpectedNextState  string `json:"expected_next_state,omitempty"`
}

// CreateMissionRequest is the request body for creating a mission
type CreateMissionRequest struct {
	Name                 string  `json:"name"`
	TargetURL            string  `json:"target_url"`
	NumAgents            int     `json:"num_agents"`
	Goal                 string  `json:"goal"`
	MaxDurationSeconds   int     `json:"max_duration_seconds"`
	RateLimitPerSecond   float64 `json:"rate_limit_per_second"`
	InitialSystemPrompt  string  `json:"initial_system_prompt"`
}

// CreateMissionResponse is the response when creating a mission
type CreateMissionResponse struct {
	MissionID string `json:"mission_id"`
}

// Event represents any event that can be broadcast via WebSocket
type Event struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Data      any       `json:"data"`
}

// AgentEvent is an event specific to an agent
type AgentEvent struct {
	AgentID  string     `json:"agent_id"`
	MissionID string    `json:"mission_id"`
	Status   string     `json:"status"`
	ActionLog *ActionLog `json:"action_log,omitempty"`
}

// SummaryEvent is a periodic summary of mission progress
type SummaryEvent struct {
	MissionID        string  `json:"mission_id"`
	TotalAgents      int     `json:"total_agents"`
	ActiveAgents     int     `json:"active_agents"`
	CompletedAgents  int     `json:"completed_agents"`
	FailedAgents     int     `json:"failed_agents"`
	TotalActions     int     `json:"total_actions"`
	TotalErrors      int     `json:"total_errors"`
	AverageLatencyMS int64   `json:"average_latency_ms"`
	ErrorRatePercent float64 `json:"error_rate_percent"`
}

// MissionStatusResponse is the response for mission status
type MissionStatusResponse struct {
	Mission      *Mission      `json:"mission"`
	AgentStates  []Agent       `json:"agent_states"`
	Summary      *SummaryEvent `json:"summary"`
}
