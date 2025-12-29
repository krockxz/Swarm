"use client";

import { Mission, SummaryEvent } from "@/lib/types";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ScrollArea } from "@/components/ui/scroll-area";
import { formatDistanceToNow } from "date-fns";
import { Activity, Clock, Users, AlertCircle, CheckCircle } from "lucide-react";

interface MissionDetailsProps {
  mission: Mission;
  summary: SummaryEvent | null;
}

export function MissionDetails({ mission, summary }: MissionDetailsProps) {
  const progress = mission.num_agents > 0
    ? ((mission.completed_agents + mission.failed_agents) / mission.num_agents) * 100
    : 0;

  return (
    <div className="space-y-4">
      {/* Header */}
      <Card>
        <CardHeader>
          <div className="flex items-start justify-between gap-4">
            <div className="space-y-1 min-w-0">
              <CardTitle className="truncate">{mission.name}</CardTitle>
              <CardDescription className="truncate">
                {mission.target_url}
              </CardDescription>
            </div>
            <Badge variant={mission.status === "running" ? "default" : "secondary"} className="shrink-0">
              {mission.status}
            </Badge>
          </div>
        </CardHeader>
        <CardContent>
          <div className="space-y-2 text-sm">
            <p>
              <span className="font-medium">Goal:</span> {mission.goal}
            </p>
            <div className="flex items-center gap-4 text-muted-foreground">
              <span className="flex items-center gap-1">
                <Users className="h-3 w-3" />
                {mission.num_agents} agents
              </span>
              <span className="flex items-center gap-1">
                <Clock className="h-3 w-3" />
                {formatDistanceToNow(new Date(mission.created_at), { addSuffix: true })}
              </span>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Metrics */}
      <div className="grid gap-4 md:grid-cols-4">
        <MetricCard
          title="Active"
          value={summary?.active_agents ?? 0}
          total={mission.num_agents}
          icon={<Activity className="h-4 w-4" />}
        />
        <MetricCard
          title="Completed"
          value={mission.completed_agents}
          total={mission.num_agents}
          icon={<CheckCircle className="h-4 w-4" />}
          variant="success"
        />
        <MetricCard
          title="Failed"
          value={mission.failed_agents}
          total={mission.num_agents}
          icon={<AlertCircle className="h-4 w-4" />}
          variant="destructive"
        />
        <MetricCard
          title="Total Actions"
          value={mission.total_actions}
          suffix={mission.average_latency_ms > 0 ? ` (${mission.average_latency_ms}ms avg)` : ""}
        />
      </div>

      {/* Progress */}
      <Card>
        <CardHeader>
          <CardTitle className="text-sm">Progress</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            <div className="flex justify-between text-sm">
              <span>{Math.round(progress)}% complete</span>
              <span className="text-muted-foreground">
                {mission.completed_agents + mission.failed_agents} / {mission.num_agents} agents
              </span>
            </div>
            <div className="h-2 w-full overflow-hidden rounded-full bg-secondary">
              <div
                className="h-full bg-primary transition-all duration-300"
                style={{ width: `${progress}%` }}
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Error Rate */}
      {summary && summary.error_rate_percent > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-sm">Error Rate</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-2">
              <div className="text-2xl font-bold">
                {summary.error_rate_percent.toFixed(1)}%
              </div>
              <Badge variant={summary.error_rate_percent > 10 ? "destructive" : "secondary"}>
                {summary.total_errors} / {summary.total_actions + summary.total_errors}
              </Badge>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

interface MetricCardProps {
  title: string;
  value: number;
  total?: number;
  suffix?: string;
  icon?: React.ReactNode;
  variant?: "default" | "success" | "destructive";
}

function MetricCard({ title, value, total, suffix, icon, variant = "default" }: MetricCardProps) {
  const variantColors = {
    default: "text-foreground",
    success: "text-green-600 dark:text-green-400",
    destructive: "text-red-600 dark:text-red-400",
  };

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-xs font-medium">{title}</CardTitle>
        {icon && <div className={variantColors[variant]}>{icon}</div>}
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">
          {value}
          {suffix && <span className="text-sm font-normal text-muted-foreground">{suffix}</span>}
        </div>
        {total !== undefined && (
          <p className="text-xs text-muted-foreground">out of {total}</p>
        )}
      </CardContent>
    </Card>
  );
}
