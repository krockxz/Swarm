"use client";

import { Agent } from "@/lib/types";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { formatDistanceToNow } from "date-fns";

interface AgentListProps {
  agents: Agent[];
  selectedAgentId: string | null;
  onSelectAgent: (agentId: string) => void;
}

const statusVariants: Record<
  Agent["status"],
  "default" | "secondary" | "destructive" | "outline"
> = {
  starting: "secondary",
  running: "default",
  completed: "outline",
  failed: "destructive",
  cancelled: "destructive",
  rate_limited: "secondary",
};

export function AgentList({
  agents,
  selectedAgentId,
  onSelectAgent,
}: AgentListProps) {
  return (
    <Card className="h-full">
      <CardHeader>
        <CardTitle>Agents</CardTitle>
        <p className="text-sm text-muted-foreground">
          {agents.length} {agents.length === 1 ? "agent" : "agents"}
        </p>
      </CardHeader>
      <CardContent className="p-0">
        <ScrollArea className="h-[calc(100vh-250px)]">
          <div className="space-y-1 p-4">
            {agents.map((agent) => (
              <button
                key={agent.id}
                onClick={() => onSelectAgent(agent.id)}
                className={`w-full rounded-lg border p-3 text-left transition-colors hover:bg-accent ${
                  selectedAgentId === agent.id ? "border-primary bg-accent" : ""
                }`}
              >
                <div className="flex items-start justify-between gap-2">
                  <div className="flex-1 space-y-1">
                    <p className="truncate text-sm font-medium">{agent.id}</p>
                    <p className="truncate text-xs text-muted-foreground">
                      {agent.current_url}
                    </p>
                  </div>
                  <Badge variant={statusVariants[agent.status]} className="shrink-0">
                    {agent.status}
                  </Badge>
                </div>

                <div className="mt-2 flex items-center gap-3 text-xs text-muted-foreground">
                  <span>{agent.success_count} actions</span>
                  <span>{agent.error_count} errors</span>
                  {agent.last_action_at && (
                    <span>{formatDistanceToNow(new Date(agent.last_action_at), { addSuffix: true })}</span>
                  )}
                </div>
              </button>
            ))}
          </div>
        </ScrollArea>
      </CardContent>
    </Card>
  );
}
