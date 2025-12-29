"use client";

import { Agent } from "@/lib/types";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { formatDistanceToNow } from "date-fns";
import { CheckCircle, XCircle, Clock } from "lucide-react";

interface AgentDetailsProps {
  agent: Agent;
}

export function AgentDetails({ agent }: AgentDetailsProps) {
  return (
    <div className="space-y-4">
      {/* Header */}
      <Card>
        <CardHeader>
          <div className="flex items-start justify-between">
            <div className="space-y-1">
              <CardTitle>{agent.id}</CardTitle>
              <p className="text-sm text-muted-foreground break-all">
                {agent.current_url}
              </p>
            </div>
            <Badge variant={agent.status === "running" ? "default" : "secondary"}>
              {agent.status}
            </Badge>
          </div>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-3 gap-4">
            <div className="flex items-center gap-2">
              <CheckCircle className="h-4 w-4 text-green-600" />
              <div>
                <p className="text-2xl font-bold">{agent.success_count}</p>
                <p className="text-xs text-muted-foreground">Success</p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <XCircle className="h-4 w-4 text-red-600" />
              <div>
                <p className="text-2xl font-bold">{agent.error_count}</p>
                <p className="text-xs text-muted-foreground">Errors</p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <Clock className="h-4 w-4 text-blue-600" />
              <div>
                <p className="text-2xl font-bold">{agent.total_latency_ms}ms</p>
                <p className="text-xs text-muted-foreground">Avg Latency</p>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Action History */}
      <Card>
        <CardHeader>
          <CardTitle className="text-sm">Action History</CardTitle>
        </CardHeader>
        <CardContent>
          <ScrollArea className="h-[400px]">
            <div className="space-y-2">
              {agent.action_history.length === 0 ? (
                <p className="text-sm text-muted-foreground">No actions yet</p>
              ) : (
                agent.action_history.map((action, index) => (
                  <div
                    key={index}
                    className="rounded-lg border p-2 text-sm"
                  >
                    <span className="font-medium">{index + 1}.</span> {action}
                  </div>
                ))
              )}
            </div>
          </ScrollArea>
        </CardContent>
      </Card>

      {/* URL History */}
      {agent.url_history.length > 1 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-sm">URL History</CardTitle>
          </CardHeader>
          <CardContent>
            <ScrollArea className="h-[200px]">
              <div className="space-y-1">
                {agent.url_history.map((url, index) => (
                  <div key={index} className="flex gap-2 text-sm">
                    <span className="text-muted-foreground">{index + 1}.</span>
                    <span className="break-all">{url}</span>
                  </div>
                ))}
              </div>
            </ScrollArea>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
