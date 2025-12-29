# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

SwarmTest is an AI-powered distributed "user swarm" testing tool. It spawns N concurrent agents that navigate web pages autonomously using Google Gemini 2.0 Flash for decision making.

**Key characteristics:**
- HTTP-only (no browser automation, no JavaScript execution)
- Token bucket rate limiting per mission
- Real-time WebSocket event streaming
- Static HTML parsing for element extraction
- AI-driven action selection (click, type, wait, go_back)

## Development Commands

### Backend (Go)

```bash
# Run the server
go run .

# Set required environment variable first
export GOOGLE_API_KEY="your-api-key-here"

# Run all tests
go test -v

# Run tests with coverage
go test -cover

# Run specific test
go test -v -run TestHTMLParser

# Run benchmark tests
go test -bench=.
```

### Dashboard (Next.js)

The dashboard is located in the `dashboard/` directory.

```bash
cd dashboard

# Install dependencies (requires Bun)
bun install

# Run development server
bun dev

# Build for production
bun run build
bun start
```

The dashboard connects to the backend at `http://localhost:8080` by default.

## Architecture Overview

The system follows an event-driven architecture with these core components:

### Data Flow
```
REST API Request → Create Mission → Spawn N Agents → Each Agent Loop:
  1. Rate limiter.Wait()
  2. Fetch page (HTTP GET)
  3. Parse HTML → StrippedPage
  4. Ask Gemini for next action
  5. Execute action (click/type/wait/go_back)
  6. Emit event to WebSocket hub
```

### Key Components

**config.go**: Core data structures
- `Mission`: Test mission configuration and runtime metrics
- `Agent`: Single agent state and metrics
- `StrippedPage`: Simplified page representation with interactive elements
- `GeminiDecisionRequest/Response`: AI decision contract

**agent.go**: Main agent execution loop
- Runs up to 30 steps or until max consecutive errors (3)
- Uses rate limiter before each page fetch
- Calls Gemini for action decision
- Executes actions via `ActionExecutor`
- Emits events to the event bus

**html_parser.go**: Static HTML parsing
- Extracts interactive elements (links, buttons, inputs, forms)
- Generates CSS selectors for elements
- No JavaScript execution

**http_client.go**: HTTP client and action execution
- `ActionExecutor`: Executes click/type actions via HTTP
- Handles form submissions (GET/POST)
- Cookie jar support for session persistence
- Retry logic for failed requests

**gemini_client.go**: AI decision making
- Uses Gemini 2.0 Flash model
- Constructs system prompt with action requirements
- Validates responses (action type, selector requirements)
- Retry logic (up to 3 attempts) for parse/validation failures

**rate_limiter.go**: Token bucket rate limiting
- Per-mission rate limiters
- Burst capacity = rate + 1 tokens
- All agents in a mission share one limiter

**websocket_server.go**: Event broadcasting
- `WebSocketHub`: Manages connections and broadcasts
- Event types: `agent_status`, `action`, `summary`, `summary_tick`
- `MissionSummaryBroadcaster`: Periodic mission summaries every 5s

**rest_api.go**: REST API handlers
- POST /api/missions: Create mission (starts immediately)
- GET /api/missions: List all missions
- GET /api/missions/{id}: Get mission status with agent states
- Uses goroutines to run missions asynchronously

## Important Constraints

**No JavaScript Execution**: Agents parse static HTML only. This means:
- Cannot interact with dynamic content loaded via JS
- Cannot handle SPAs that require client-side routing
- Click actions only work for `<a href="">` links and form submissions

**HTTP-Only Limitations**:
- Cannot execute JavaScript event handlers (onclick, etc.)
- Cannot interact with elements that require JS to become visible
- Form submission is best-effort (basic forms only)

**Agent Lifecycle**:
- Maximum 30 steps per agent (hardcoded)
- Max 3 consecutive errors before agent fails
- Human-like delays (500ms-3000ms) between actions

## Gemini Prompt Engineering

The system prompt (in `gemini_client.go`) enforces:
- JSON-only responses (no markdown formatting)
- Specific action types: click, type, wait, go_back
- Selector requirement for click/type actions
- TextInput requirement for type action

The user prompt includes:
- Mission goal
- Current page title, URL, description
- All interactive elements with selectors
- Recent action history (last 10 actions)

## Testing Patterns

**MockGeminiClient** (in main_test.go): Use for testing agent logic without real AI calls
```go
mockGemini := &MockGeminiClient{
    decisions: []GeminiDecisionResponse{{
        Action: "click",
        Selector: "a",
        Reasoning: "Click the link",
    }},
}
```

**httptest.NewServer**: Use for testing HTTP actions against a mock server
```go
ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Return test HTML
}))
defer ts.Close()
```

## Common Tasks

**Add a new action type:**
1. Add action to `AvailableActions` in agent.go
2. Add case in `ActionExecutor.ExecuteAction()` in http_client.go
3. Add validation in `validateResponse()` in gemini_client.go
4. Update system prompt to include new action

**Modify rate limiting behavior:**
- Edit `NewRateLimiter()` to change capacity calculation
- Current: capacity = int(rate) + 1 (burst of 1 second)

**Customize element extraction:**
- Edit `extractElements()` in html_parser.go
- Add new element types to CSS selectors

**Change agent retry/failure logic:**
- Modify constants in agent.go: `maxSteps`, `maxConsecutiveErrors`
- Add new failure conditions in the agent Run() loop

## Dashboard Architecture

The dashboard is a Next.js 15 application with TypeScript and shadcn/ui components.

### Tech Stack
- **Framework**: Next.js 15 with App Router (Server Components)
- **Styling**: Tailwind CSS with shadcn/ui components
- **State Management**: TanStack Query for server state, React hooks for local state
- **Forms**: React Hook Form + Zod validation
- **Real-time**: Custom WebSocket hook for live event streaming

### Directory Structure
```
dashboard/
├── src/
│   ├── app/                    # Next.js App Router
│   │   ├── layout.tsx          # Root layout with Toaster
│   │   ├── page.tsx            # Main dashboard page
│   │   └── globals.css         # Global styles with CSS variables
│   ├── components/
│   │   ├── ui/                 # shadcn/ui base components
│   │   ├── mission/            # Mission-related components
│   │   │   ├── mission-create-form.tsx
│   │   │   ├── mission-list.tsx
│   │   │   └── mission-details.tsx
│   │   └── agent/              # Agent-related components
│   │       ├── agent-list.tsx
│   │       ├── agent-details.tsx
│   │       └── events-display.tsx
│   ├── hooks/
│   │   ├── use-websocket.ts    # Generic WebSocket connection manager
│   │   └── use-mission-events.ts # Mission-specific event handling
│   └── lib/
│       ├── types.ts            # TypeScript types (matches backend)
│       ├── api-client.ts       # REST API client
│       └── utils.ts            # Utility functions (cn helper)
└── package.json
```

### Key Components

**MissionCreateForm**: Creates new missions with React Hook Form + Zod validation
- Validates all mission parameters
- Shows loading state during submission
- Calls `apiClient.createMission()`

**MissionList**: Displays all missions with selection
- Shows mission status badges
- Displays agent count and progress
- Auto-selects first mission on load

**MissionDetails**: Shows selected mission details
- Metrics grid (active, completed, failed agents)
- Progress bar
- Real-time summary from WebSocket

**AgentList**: Lists agents for selected mission
- Shows agent status and current URL
- Displays success/error counts
- Click to view agent details

**AgentDetails**: Shows detailed agent information
- Action history with timestamps
- URL navigation history
- Success/error metrics

**EventsDisplay**: Live event feed from WebSocket
- Shows real-time agent actions
- Displays latency and error messages
- Auto-scrolls to latest events

### WebSocket Integration

The `useWebSocket` hook manages the WebSocket connection:
- Auto-reconnect on disconnect (3s interval)
- Message parsing and error handling
- Connection status tracking

The `useMissionEvents` hook filters events for a specific mission:
- Filters events by `mission_id`
- Maintains last 500 events
- Updates summary statistics

### Data Fetching

Uses TanStack Query for server state:
- `useQuery` for missions list (5s poll interval)
- `useQuery` for mission status (3s poll interval)
- Automatic refetch on mutation success
- Optimistic updates where appropriate

### Dashboard Development Tasks

**Add a new mission field:**
1. Update `CreateMissionRequest` type in `lib/types.ts`
2. Add form field in `MissionCreateForm` with Zod validation
3. Update backend Go struct to accept the field

**Add a new event type:**
1. Update `WebSocketEvent` type in `lib/types.ts`
2. Add handler in `use-mission-events.ts`
3. Create display component for the event

**Customize polling intervals:**
- Edit `refetchInterval` in `useQuery` calls (page.tsx)
- Default: 5s for missions, 3s for mission status
