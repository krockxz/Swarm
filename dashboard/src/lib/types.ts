// Types matching the backend Go structs

export interface Mission {
  id: string;
  name: string;
  target_url: string;
  num_agents: number;
  goal: string;
  max_duration_seconds: number;
  rate_limit_per_second: number;
  initial_system_prompt: string;
  status: "pending" | "running" | "completed" | "cancelled";
  created_at: string;
  started_at?: string;
  completed_at?: string;
  total_actions: number;
  total_errors: number;
  average_latency_ms: number;
  completed_agents: number;
  failed_agents: number;
  recent_events: ActionLog[];
  agent_metrics: Record<string, Agent>;
}

export interface Agent {
  id: string;
  mission_id: string;
  status: "starting" | "running" | "completed" | "failed" | "cancelled" | "rate_limited";
  current_url: string;
  action_history: string[];
  error_count: number;
  success_count: number;
  total_latency_ms: number;
  consecutive_errors: number;
  url_history: string[];
  last_action_at?: string;
}

export interface ActionLog {
  timestamp: string;
  agent_id: string;
  action: string;
  selector?: string;
  result: string;
  latency_ms: number;
  error_message?: string;
  new_url?: string;
}

export interface StrippedPage {
  url: string;
  title: string;
  description: string;
  interactive_elements: Element[];
  timestamp: string;
}

export interface Element {
  id?: string;
  type: "button" | "link" | "input" | "form";
  text?: string;
  selector: string;
  href?: string;
  name?: string;
  placeholder?: string;
  input_type?: string;
}

export interface CreateMissionRequest {
  name: string;
  target_url: string;
  num_agents: number;
  goal: string;
  max_duration_seconds: number;
  rate_limit_per_second: number;
  initial_system_prompt: string;
}

export interface CreateMissionResponse {
  mission_id: string;
}

export interface MissionStatusResponse {
  mission: Mission;
  agent_states: Agent[];
  summary: SummaryEvent;
}

export interface WebSocketEvent {
  type: "agent_status" | "action" | "summary" | "summary_tick";
  timestamp: string;
  data: AgentEvent | SummaryEvent;
}

export interface AgentEvent {
  agent_id: string;
  mission_id: string;
  status: string;
  action_log?: ActionLog;
}

export interface SummaryEvent {
  mission_id: string;
  total_agents: number;
  active_agents: number;
  completed_agents: number;
  failed_agents: number;
  total_actions: number;
  total_errors: number;
  average_latency_ms: number;
  error_rate_percent: number;
}
