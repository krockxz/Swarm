"use client";

import { useState, useEffect, useCallback } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Mission, Agent } from "@/lib/types";
import { apiClient } from "@/lib/api-client";
import { MissionCreateForm } from "@/components/mission/mission-create-form";
import { MissionList } from "@/components/mission/mission-list";
import { MissionDetails } from "@/components/mission/mission-details";
import { AgentList } from "@/components/agent/agent-list";
import { AgentDetails } from "@/components/agent/agent-details";
import { EventsDisplay } from "@/components/agent/events-display";
import { useMissionEvents } from "@/hooks/use-mission-events";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Loader2, Wifi, WifiOff, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";

export default function HomePage() {
  const queryClient = useQueryClient();
  const [selectedMissionId, setSelectedMissionId] = useState<string | null>(null);
  const [selectedAgentId, setSelectedAgentId] = useState<string | null>(null);

  // Fetch missions
  const {
    data: missionsData,
    isLoading: isLoadingMissions,
    error: missionsError,
    refetch: refetchMissions,
  } = useQuery({
    queryKey: ["missions"],
    queryFn: () => apiClient.listMissions(),
    refetchInterval: 5000, // Poll every 5 seconds
  });

  // Fetch selected mission details
  const { data: missionStatus } = useQuery({
    queryKey: ["mission", selectedMissionId],
    queryFn: () => apiClient.getMission(selectedMissionId!),
    enabled: !!selectedMissionId,
    refetchInterval: 3000, // Poll every 3 seconds for active missions
  });

  // WebSocket events for selected mission
  const { agentEvents, summary, isConnected, clearEvents } = useMissionEvents(
    selectedMissionId
  );

  // Auto-select first mission
  useEffect(() => {
    if (missionsData?.missions && missionsData.missions.length > 0 && !selectedMissionId) {
      setSelectedMissionId(missionsData.missions[0].id);
    }
  }, [missionsData, selectedMissionId]);

  // Handle mission selection
  const handleSelectMission = useCallback((missionId: string) => {
    setSelectedMissionId(missionId);
    setSelectedAgentId(null);
    clearEvents();
  }, [clearEvents]);

  // Handle agent selection
  const handleSelectAgent = useCallback((agentId: string) => {
    setSelectedAgentId(agentId);
  }, []);

  // Get selected agent
  const selectedAgent = missionStatus?.agent_states?.find(
    (agent) => agent.id === selectedAgentId
  );

  // Get selected mission
  const selectedMission = missionsData?.missions?.find(
    (mission) => mission.id === selectedMissionId
  );

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b">
        <div className="container flex h-16 items-center justify-between px-4">
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold">SwarmTest Dashboard</h1>
            {isConnected ? (
              <Wifi className="h-4 w-4 text-green-600" />
            ) : (
              <WifiOff className="h-4 w-4 text-muted-foreground" />
            )}
          </div>
          <div className="flex items-center gap-4">
            <Button
              variant="outline"
              size="sm"
              onClick={() => refetchMissions()}
              disabled={isLoadingMissions}
            >
              <RefreshCw className={`mr-2 h-4 w-4 ${isLoadingMissions ? "animate-spin" : ""}`} />
              Refresh
            </Button>
            <MissionCreateForm
              onSuccess={(missionId) => {
                queryClient.invalidateQueries({ queryKey: ["missions"] });
                setSelectedMissionId(missionId);
              }}
            />
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="container h-[calc(100vh-4rem)] p-4">
        {isLoadingMissions ? (
          <div className="flex h-full items-center justify-center">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : missionsError ? (
          <div className="flex h-full items-center justify-center">
            <div className="text-center">
              <p className="text-lg font-medium text-destructive">Failed to load missions</p>
              <p className="text-sm text-muted-foreground">
                Make sure the SwarmTest backend is running on port 8080
              </p>
            </div>
          </div>
        ) : !missionsData?.missions || missionsData.missions.length === 0 ? (
          <div className="flex h-full items-center justify-center">
            <div className="text-center">
              <p className="text-lg font-medium">No missions yet</p>
              <p className="text-sm text-muted-foreground">
                Create your first mission to get started
              </p>
            </div>
          </div>
        ) : (
          <div className="grid h-full grid-cols-12 gap-4">
            {/* Mission List */}
            <div className="col-span-3">
              <MissionList
                missions={missionsData.missions}
                selectedMissionId={selectedMissionId}
                onSelectMission={handleSelectMission}
              />
            </div>

            {/* Mission Details / Agents */}
            <div className="col-span-3">
              {selectedMission ? (
                <Tabs defaultValue="mission" className="h-full">
                  <TabsList className="w-full">
                    <TabsTrigger value="mission" className="flex-1">
                      Mission
                    </TabsTrigger>
                    <TabsTrigger value="agents" className="flex-1">
                      Agents
                    </TabsTrigger>
                  </TabsList>
                  <TabsContent value="mission" className="h-[calc(100vh-350px)] overflow-y-auto">
                    {missionStatus ? (
                      <MissionDetails
                        mission={missionStatus.mission}
                        summary={summary}
                      />
                    ) : (
                      <Loader2 className="h-8 w-8 animate-spin" />
                    )}
                  </TabsContent>
                  <TabsContent value="agents" className="h-[calc(100vh-350px)] overflow-hidden">
                    <AgentList
                      agents={missionStatus?.agent_states ?? []}
                      selectedAgentId={selectedAgentId}
                      onSelectAgent={handleSelectAgent}
                    />
                  </TabsContent>
                </Tabs>
              ) : (
                <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
                  Select a mission to view details
                </div>
              )}
            </div>

            {/* Agent Details / Events */}
            <div className="col-span-6">
              {selectedAgent ? (
                <div className="h-[calc(100vh-150px)] overflow-y-auto">
                  <AgentDetails agent={selectedAgent} />
                </div>
              ) : (
                <EventsDisplay events={agentEvents} />
              )}
            </div>
          </div>
        )}
      </main>
    </div>
  );
}
