# SwarmTest Dashboard

Modern web-based dashboard for the SwarmTest AI-powered distributed testing tool.

## Tech Stack

- **Framework**: Next.js 15 with App Router
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **UI Components**: shadcn/ui (Radix UI primitives)
- **State Management**: TanStack Query for server state
- **Forms**: React Hook Form + Zod validation
- **Real-time**: WebSocket connection for live events

## Prerequisites

- SwarmTest backend running on `http://localhost:8080`
- Node.js 18+ and Bun

## Installation

```bash
cd dashboard
bun install
```

## Development

```bash
bun dev
```

The dashboard will be available at `http://localhost:3000`.

## Configuration

Environment variables (optional):

```bash
# Backend API URL
NEXT_PUBLIC_API_URL=http://localhost:8080

# WebSocket URL
NEXT_PUBLIC_WS_URL=ws://localhost:8080/ws
```

## Features

### Mission Management
- Create new testing missions with configurable parameters
- View all missions in a list with status indicators
- Monitor mission progress in real-time

### Real-time Events
- Live WebSocket connection to the backend
- Real-time agent status updates
- Action logs with latency information
- Error tracking and alerts

### Agent Monitoring
- View individual agent status
- Track agent action history
- Monitor URL navigation paths
- View success/error counts

## Project Structure

```
dashboard/
├── src/
│   ├── app/              # Next.js app directory
│   │   ├── globals.css   # Global styles
│   │   ├── layout.tsx    # Root layout
│   │   └── page.tsx      # Main dashboard page
│   ├── components/       # React components
│   │   ├── ui/          # shadcn/ui components
│   │   ├── mission/     # Mission-related components
│   │   └── agent/       # Agent-related components
│   ├── hooks/           # Custom React hooks
│   │   ├── use-websocket.ts
│   │   └── use-mission-events.ts
│   └── lib/             # Utilities
│       ├── types.ts     # TypeScript types
│       ├── api-client.ts
│       └── utils.ts
├── public/              # Static assets
└── package.json
```

## Building for Production

```bash
bun run build
bun start
```

## Component Architecture

### Mission Components
- `MissionCreateForm`: Form to create new missions with validation
- `MissionList`: List of all missions with selection
- `MissionDetails`: Detailed view of selected mission with metrics

### Agent Components
- `AgentList`: List of agents for selected mission
- `AgentDetails`: Detailed view of selected agent
- `EventsDisplay`: Live event feed from WebSocket

### Hooks
- `useWebSocket`: Generic WebSocket connection manager
- `useMissionEvents`: Mission-specific event handling with filtering
