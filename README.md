# SwarmTest

SwarmTest is an AI-powered distributed "user swarm" testing tool. It spawns N concurrent agents that navigate web pages autonomously using Gemini AI for decision making.

## Features

- **Concurrent Agent Testing**: Run multiple agents in parallel to simulate real user traffic
- **AI-Powered Navigation**: Uses Google Gemini 2.0 Flash to decide next actions
- **HTML Parsing**: Extracts interactive elements (links, buttons, forms, inputs) without executing JavaScript
- **Rate Limiting**: Token bucket rate limiting to control request frequency
- **Real-time Events**: WebSocket streaming of agent actions and mission progress
- **REST API**: Simple API for creating and monitoring missions

## Architecture

```
┌─────────────┐     HTTP      ┌──────────────┐
│   Target    │◄──────────────┤   SwarmTest  │
│   Website   │               │   Backend    │
└─────────────┘               └──────┬───────┘
                                      │
                              ┌───────┴────────┐
                              ▼                ▼
                       ┌────────────┐  ┌─────────────┐
                       │  Gemini    │  │  WebSocket  │
                       │    AI      │  │   Events    │
                       └────────────┘  └─────────────┘
```

## Prerequisites

- Go 1.21 or higher
- Google API Key with access to Gemini API
- A target website to test

## Installation

### Backend (Go)

1. Clone the repository:
```bash
git clone <repository-url>
cd swarmtest
```

2. Set your Google API Key:
```bash
export GOOGLE_API_KEY="your-api-key-here"
```

3. Install dependencies:
```bash
go mod download
```

4. Run the server:
```bash
go run .
```

The server will start on `http://localhost:8080`

### Web Dashboard (Optional)

A modern web dashboard is available for visualizing missions and real-time agent activity.

1. Navigate to the dashboard directory:
```bash
cd dashboard
```

2. Install dependencies (requires [Bun](https://bun.sh)):
```bash
bun install
```

3. Start the development server:
```bash
bun dev
```

The dashboard will be available at `http://localhost:3000`

**Dashboard Features:**
- Create and monitor missions through a web UI
- Real-time agent activity feed via WebSocket
- Per-agent status and action history
- Mission metrics and progress visualization
- Built with Next.js 15, TypeScript, Tailwind CSS, and shadcn/ui

## API Endpoints

### Create Mission
```http
POST /api/missions
Content-Type: application/json

{
  "name": "Test Homepage Navigation",
  "target_url": "https://example.com",
  "num_agents": 5,
  "goal": "Navigate to the about page and find contact information",
  "max_duration_seconds": 300,
  "rate_limit_per_second": 2.0,
  "initial_system_prompt": "You are testing the website navigation. Focus on finding the about page."
}
```

Response:
```json
{
  "mission_id": "mission-abc12345"
}
```

### List Missions
```http
GET /api/missions
```

### Get Mission Status
```http
GET /api/missions/{mission_id}
```

### Health Check
```http
GET /api/health
```

### WebSocket Events
```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Event:', data.type, data.data);
};

// Event types:
// - "agent_status": Agent status changes
// - "action": Individual actions performed by agents
// - "summary": Periodic mission summary
// - "summary_tick": Periodic keepalive tick
```

## Configuration

### Mission Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | Mission name |
| `target_url` | string | Yes | Starting URL for agents |
| `num_agents` | int | Yes | Number of agents (1-1000) |
| `goal` | string | Yes | Mission goal for AI |
| `max_duration_seconds` | int | Yes | Maximum mission duration (10-3600) |
| `rate_limit_per_second` | float | Yes | Request rate limit (0-1000) |
| `initial_system_prompt` | string | No | Custom system prompt for AI |

## Agent Actions

Agents can perform the following actions:

- **click**: Click on buttons or links
- **type**: Fill input fields and submit forms
- **wait**: Pause and observe the page
- **go_back**: Navigate to the previous page

## Example Usage

### Using cURL

```bash
# Create a mission
curl -X POST http://localhost:8080/api/missions \
  -H "Content-Type: application/json" \
  -d @example_mission.json

# Get mission status
curl http://localhost:8080/api/missions/mission-abc12345

# List all missions
curl http://localhost:8080/api/missions
```

### Using WebSocket (Node.js)

```javascript
const WebSocket = require('ws');

const ws = new WebSocket('ws://localhost:8080/ws');

ws.on('open', () => {
  console.log('Connected to SwarmTest');
});

ws.on('message', (data) => {
  const event = JSON.parse(data);

  switch(event.type) {
    case 'action':
      console.log(`[${event.data.agent_id}] ${event.data.action_log.action}`);
      break;
    case 'summary':
      console.log(`Summary: ${event.data.active_agents}/${event.data.total_agents} active`);
      break;
  }
});
```

## Development

### Running Tests

```bash
# Run all tests
go test -v

# Run with coverage
go test -cover

# Run benchmark tests
go test -bench=.

# Run specific test
go test -v -run TestHTMLParser
```

### Project Structure

```
swarmtest/
├── main.go                 # Entry point and server setup
├── config.go              # Core data structures
├── agent.go               # Agent implementation
├── html_parser.go         # HTML parsing logic
├── gemini_client.go       # Gemini AI integration
├── http_client.go         # HTTP client and action execution
├── websocket_server.go    # WebSocket server
├── rest_api.go            # REST API handlers
├── rate_limiter.go        # Rate limiting
├── main_test.go           # Tests
├── go.mod                 # Go module definition
├── dashboard/             # Web dashboard (Next.js)
│   ├── src/
│   │   ├── app/          # Next.js App Router
│   │   ├── components/   # React components
│   │   ├── hooks/        # Custom hooks (WebSocket, etc.)
│   │   └── lib/          # Types and API client
│   └── package.json
└── README.md              # This file
```

## Limitations

- **No JavaScript Execution**: Agents parse static HTML only; JavaScript-heavy sites may not work correctly
- **HTTP-Only**: No browser automation; all actions are performed via HTTP requests
- **Form Submission**: Basic form support; complex forms with JavaScript validation may not work
- **Authentication**: No built-in authentication support

## Troubleshooting

### High Error Rates

- Reduce `rate_limit_per_second` to avoid overwhelming the target server
- Check if the target site requires JavaScript
- Verify the target URL is accessible

### Agents Not Progressing

- Check agent logs via WebSocket for specific errors
- Review the mission goal - ensure it's achievable
- Consider increasing `max_duration_seconds`

### Gemini API Errors

- Verify `GOOGLE_API_KEY` is set correctly
- Check API quota and billing
- Monitor for rate limiting on the Gemini API

## License

MIT License - see LICENSE file for details
