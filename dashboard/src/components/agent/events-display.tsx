"use client";

import { AgentEvent, ActionLog } from "@/lib/types";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { formatDistanceToNow } from "date-fns";
import { CheckCircle, XCircle } from "lucide-react";

interface EventsDisplayProps {
  events: AgentEvent[];
  maxEvents?: number;
}

export function EventsDisplay({ events, maxEvents = 100 }: EventsDisplayProps) {
  const displayEvents = events.slice(0, maxEvents);

  return (
    <Card className="h-full">
      <CardHeader>
        <CardTitle>Live Events</CardTitle>
        <p className="text-sm text-muted-foreground">
          {displayEvents.length} {displayEvents.length === 1 ? "event" : "events"}
        </p>
      </CardHeader>
      <CardContent className="p-0">
        <ScrollArea className="h-[calc(100vh-250px)]">
          <div className="space-y-1 p-4">
            {displayEvents.length === 0 ? (
              <div className="flex h-32 items-center justify-center text-sm text-muted-foreground">
                No events yet. Start a mission to see real-time events.
              </div>
            ) : (
              displayEvents.map((event, index) => (
                <div
                  key={`${event.agent_id}-${index}`}
                  className="rounded-lg border bg-card p-3"
                >
                  <div className="flex items-start justify-between gap-2">
                    <div className="flex-1 space-y-1">
                      <p className="text-sm font-medium">{event.agent_id}</p>
                      {event.action_log ? (
                        <>
                          <p className="text-xs text-muted-foreground">
                            {event.action_log.action}
                            {event.action_log.selector && (
                              <span className="ml-1 text-muted-foreground">
                                → {event.action_log.selector}
                              </span>
                            )}
                          </p>
                          {event.action_log.error_message && (
                            <p className="text-xs text-destructive">
                              {event.action_log.error_message}
                            </p>
                          )}
                        </>
                      ) : (
                        <Badge variant="outline">{event.status}</Badge>
                      )}
                    </div>
                    <div className="flex flex-col items-end gap-1">
                      {event.action_log && (
                        <div className="flex items-center gap-1">
                          {event.action_log.result === "success" ? (
                            <CheckCircle className="h-3 w-3 text-green-600" />
                          ) : (
                            <XCircle className="h-3 w-3 text-red-600" />
                          )}
                          <span className="text-xs text-muted-foreground">
                            {event.action_log.latency_ms}ms
                          </span>
                        </div>
                      )}
                      <span className="text-xs text-muted-foreground">
                        {formatDistanceToNow(new Date(event.timestamp || Date.now()), {
                          addSuffix: true,
                        })}
                      </span>
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>
        </ScrollArea>
      </CardContent>
    </Card>
  );
}
