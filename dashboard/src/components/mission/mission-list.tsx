"use client";

import { Mission } from "@/lib/types";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { formatDistanceToNow } from "date-fns";

interface MissionListProps {
  missions: Mission[];
  selectedMissionId: string | null;
  onSelectMission: (missionId: string) => void;
}

const statusVariants: Record<Mission["status"], "default" | "secondary" | "destructive" | "outline"> = {
  pending: "secondary",
  running: "default",
  completed: "outline",
  cancelled: "destructive",
};

export function MissionList({
  missions,
  selectedMissionId,
  onSelectMission,
}: MissionListProps) {
  return (
    <Card className="h-full">
      <CardHeader>
        <CardTitle>Missions</CardTitle>
        <CardDescription>
          {missions.length} {missions.length === 1 ? "mission" : "missions"} total
        </CardDescription>
      </CardHeader>
      <CardContent className="p-0">
        <ScrollArea className="h-[calc(100vh-250px)]">
          <div className="space-y-1 p-4">
            {missions.length === 0 ? (
              <div className="flex h-32 items-center justify-center text-sm text-muted-foreground">
                No missions yet. Create one to get started.
              </div>
            ) : (
              missions.map((mission) => (
                <button
                  key={mission.id}
                  onClick={() => onSelectMission(mission.id)}
                  className={`w-full rounded-lg border p-3 text-left transition-colors hover:bg-accent ${selectedMissionId === mission.id ? "border-primary bg-accent" : ""
                    }`}
                >
                  <div className="flex items-start justify-between gap-3">
                    <div className="flex-1 space-y-1 min-w-0">
                      <p className="truncate text-sm font-medium">{mission.name}</p>
                      <p className="truncate text-xs text-muted-foreground">
                        {mission.target_url}
                      </p>
                    </div>
                    <Badge variant={statusVariants[mission.status]} className="shrink-0">
                      {mission.status}
                    </Badge>
                  </div>

                  <div className="mt-2 flex items-center gap-4 text-xs text-muted-foreground">
                    <span>{mission.num_agents} agents</span>
                    <span>{mission.completed_agents + mission.failed_agents} done</span>
                    <span>{formatDistanceToNow(new Date(mission.created_at), { addSuffix: true })}</span>
                  </div>
                </button>
              ))
            )}
          </div>
        </ScrollArea>
      </CardContent>
    </Card>
  );
}
