import { CreateMissionRequest, CreateMissionResponse, Mission, MissionStatusResponse } from "./types";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string = API_BASE_URL) {
    this.baseUrl = baseUrl;
  }

  private async request<T>(
    endpoint: string,
    options?: RequestInit
  ): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;
    const response = await fetch(url, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...options?.headers,
      },
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(`API Error: ${response.status} - ${error}`);
    }

    return response.json();
  }

  async createMission(
    request: CreateMissionRequest
  ): Promise<CreateMissionResponse> {
    return this.request<CreateMissionResponse>("/api/missions", {
      method: "POST",
      body: JSON.stringify(request),
    });
  }

  async listMissions(): Promise<{ missions: Mission[] }> {
    return this.request<{ missions: Mission[] }>("/api/missions");
  }

  async getMission(missionId: string): Promise<MissionStatusResponse> {
    return this.request<MissionStatusResponse>(`/api/missions/${missionId}`);
  }

  async healthCheck(): Promise<{ status: string; version: string; build_time: string }> {
    return this.request("/api/health");
  }
}

export const apiClient = new ApiClient();
