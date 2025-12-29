"use client";

import { useState, useCallback, useEffect } from "react";
import { useWebSocket } from "./use-websocket";
import { WebSocketEvent, AgentEvent, SummaryEvent } from "@/lib/types";

interface MissionEventsState {
  agentEvents: AgentEvent[];
  summary: SummaryEvent | null;
  isLive: boolean;
}

export function useMissionEvents(missionId: string | null) {
  const [events, setEvents] = useState<MissionEventsState>({
    agentEvents: [],
    summary: null,
    isLive: false,
  });

  const handle_message = useCallback((event: WebSocketEvent) => {
    // Filter events for this mission
    if (missionId && event.data?.mission_id !== missionId) {
      return;
    }

    switch (event.type) {
      case "agent_status":
      case "action":
        setEvents((prev) => ({
          ...prev,
          agentEvents: [
            {
              ...(event.data as AgentEvent),
              timestamp: event.timestamp,
            },
            ...prev.agentEvents,
          ].slice(0, 500), // Keep last 500 events
          isLive: true,
        }));
        break;

      case "summary":
        setEvents((prev) => ({
          ...prev,
          summary: event.data as SummaryEvent,
          isLive: true,
        }));
        break;

      case "summary_tick":
        // Just indicates the connection is alive
        setEvents((prev) => ({ ...prev, isLive: true }));
        break;
    }
  }, [missionId]);

  const { isConnected, connect, disconnect } = useWebSocket({
    onMessage: handle_message,
    onConnect: () => {
      console.log("WebSocket connected");
    },
    onDisconnect: () => {
      setEvents((prev) => ({ ...prev, isLive: false }));
    },
  });

  // Auto-connect when missionId changes
  useEffect(() => {
    if (missionId) {
      connect();
    }
    return () => {
      disconnect();
    };
  }, [missionId, connect, disconnect]);

  return {
    ...events,
    isConnected,
    clearEvents: useCallback(() => {
      setEvents({ agentEvents: [], summary: null, isLive: false });
    }, []),
  };
}
