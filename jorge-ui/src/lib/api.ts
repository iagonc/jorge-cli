const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

export interface Resource {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  DeletedAt: string | null;
  name: string;
  dns: string;
}

export type MonitorType = "http" | "tcp" | "ping" | "dns" | "ssl" | "trace";
export type MonitorStatus = "up" | "down" | "degraded" | "unknown";
export type AlertSeverity = "critical" | "warning" | "info";

export interface Monitor {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  name: string;
  target: string;
  type: MonitorType;
  port: number;
  interval: number;
  timeout: number;
  enabled: boolean;
  status: MonitorStatus;
  last_check_at: string | null;
  last_latency: number;
  uptime: number;
  expected_status: number;
  method: string;
}

export interface CheckResult {
  ID: number;
  CreatedAt: string;
  monitor_id: number;
  status: MonitorStatus;
  latency: number;
  status_code: number;
  output: string;
  error_message: string;
  checked_at: string;
}

export interface Alert {
  ID: number;
  CreatedAt: string;
  monitor_id: number;
  severity: AlertSeverity;
  title: string;
  message: string;
  acknowledged: boolean;
  acked_at: string | null;
  Monitor?: Monitor;
}

export interface Stats {
  total_monitors: number;
  up_monitors: number;
  down_monitors: number;
  degraded_monitors: number;
  total_alerts: number;
  unresolved_alerts: number;
  avg_latency_ms: number;
  uptime_percent: number;
  scheduler_running: boolean;
}

export interface DiagnosticResult {
  success: boolean;
  type: string;
  target: string;
  latency_ms: number;
  output: string;
  error?: string;
  details?: Record<string, unknown>;
}

export interface LatencyBucket {
  hour: string;
  avg_latency: number;
  p95_latency: number;
  check_count: number;
}

export interface UptimeSlot {
  monitor_id: number;
  monitor_name: string;
  hour: string;
  up_count: number;
  total_count: number;
}

export interface Incident {
  id: number;
  monitor_id: number;
  monitor_name: string;
  started_at: string;
  status: string;
  message: string;
}

export interface SLOData {
  target: number;
  current: number;
  error_budget: number;
  compliant: boolean;
}

export interface Percentiles {
  p50: number;
  p95: number;
  p99: number;
}

export interface DashboardData {
  latency_history: LatencyBucket[];
  uptime_slots: UptimeSlot[];
  monitors: Monitor[];
  incidents: Incident[];
  slo: SLOData;
  percentiles: Percentiles;
}

export interface MonitorUptime {
  id: number;
  name: string;
  target: string;
  type: string;
  total_checks: number;
  success_count: number;
  fail_count: number;
  uptime_percent: number;
}

export interface MonitorPerformance {
  id: number;
  name: string;
  target: string;
  type: string;
  avg_latency: number;
  min_latency: number;
  max_latency: number;
  check_count: number;
  success_rate: number;
}

export interface UptimeReport {
  type: "uptime";
  period: string;
  generated: string;
  overall_uptime: number;
  total_checks: number;
  monitors: MonitorUptime[];
}

export interface PerformanceReport {
  type: "performance";
  period: string;
  generated: string;
  monitors: MonitorPerformance[];
}

export interface IncidentReport {
  type: "incident";
  period: string;
  generated: string;
  total: number;
  critical: number;
  warning: number;
  resolved: number;
  incidents: Alert[];
}

export type Report = UptimeReport | PerformanceReport | IncidentReport;

export interface ApiResponse<T> {
  message: string;
  data: T;
}

export interface ApiError {
  error: string;
}

class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;

    const response = await fetch(url, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...options.headers,
      },
    });

    if (!response.ok) {
      const error: ApiError = await response.json().catch(() => ({
        error: `HTTP ${response.status}: ${response.statusText}`,
      }));
      throw new Error(error.error);
    }

    return response.json();
  }

  // Resources (DNS Records)
  async listResources(): Promise<Resource[]> {
    const response = await this.request<ApiResponse<Resource[]>>("/resources");
    return response.data || [];
  }

  async getResource(id: number): Promise<Resource> {
    const response = await this.request<ApiResponse<Resource>>(
      `/resource?id=${id}`
    );
    return response.data;
  }

  async searchResources(name: string): Promise<Resource[]> {
    const response = await this.request<ApiResponse<Resource[]>>(
      `/resources/name?name=${encodeURIComponent(name)}`
    );
    return response.data || [];
  }

  async createResource(data: { name: string; dns: string }): Promise<Resource> {
    const response = await this.request<ApiResponse<Resource>>("/resource", {
      method: "POST",
      body: JSON.stringify(data),
    });
    return response.data;
  }

  async updateResource(
    id: number,
    data: { name: string; dns: string }
  ): Promise<Resource> {
    const response = await this.request<ApiResponse<Resource>>(
      `/resource?id=${id}`,
      {
        method: "PUT",
        body: JSON.stringify(data),
      }
    );
    return response.data;
  }

  async deleteResource(id: number): Promise<void> {
    await this.request(`/resource?id=${id}`, {
      method: "DELETE",
    });
  }

  // Health check
  async healthCheck(): Promise<boolean> {
    try {
      await this.request("/resources");
      return true;
    } catch {
      return false;
    }
  }

  // Monitors
  async listMonitors(): Promise<Monitor[]> {
    const response = await this.request<ApiResponse<Monitor[]>>("/monitors");
    return response.data || [];
  }

  async getMonitor(id: number): Promise<Monitor> {
    const response = await this.request<ApiResponse<Monitor>>(`/monitor?id=${id}`);
    return response.data;
  }

  async createMonitor(data: {
    name: string;
    target: string;
    type: MonitorType;
    port?: number;
    interval?: number;
    timeout?: number;
    expected_status?: number;
    method?: string;
    enabled?: boolean;
  }): Promise<Monitor> {
    const response = await this.request<ApiResponse<Monitor>>("/monitor", {
      method: "POST",
      body: JSON.stringify(data),
    });
    return response.data;
  }

  async updateMonitor(id: number, data: Partial<Monitor>): Promise<Monitor> {
    const response = await this.request<ApiResponse<Monitor>>(`/monitor?id=${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
    return response.data;
  }

  async deleteMonitor(id: number): Promise<void> {
    await this.request(`/monitor?id=${id}`, { method: "DELETE" });
  }

  async toggleMonitor(id: number): Promise<Monitor> {
    const response = await this.request<ApiResponse<Monitor>>(`/monitor/toggle?id=${id}`, {
      method: "POST",
    });
    return response.data;
  }

  async getMonitorResults(id: number, limit: number = 50): Promise<CheckResult[]> {
    const response = await this.request<ApiResponse<CheckResult[]>>(
      `/monitor/results?id=${id}&limit=${limit}`
    );
    return response.data || [];
  }

  // Diagnostics
  async runDiagnostic(data: {
    type: MonitorType;
    target: string;
    port?: number;
    timeout?: number;
  }): Promise<DiagnosticResult> {
    const response = await this.request<ApiResponse<DiagnosticResult>>("/diagnostic", {
      method: "POST",
      body: JSON.stringify(data),
    });
    return response.data;
  }

  // Alerts
  async listAlerts(filters?: {
    acknowledged?: boolean;
    severity?: AlertSeverity;
  }): Promise<Alert[]> {
    let endpoint = "/alerts";
    const params: string[] = [];
    if (filters?.acknowledged !== undefined) {
      params.push(`acknowledged=${filters.acknowledged}`);
    }
    if (filters?.severity) {
      params.push(`severity=${filters.severity}`);
    }
    if (params.length > 0) {
      endpoint += `?${params.join("&")}`;
    }
    const response = await this.request<ApiResponse<Alert[]>>(endpoint);
    return response.data || [];
  }

  async acknowledgeAlert(id: number): Promise<Alert> {
    const response = await this.request<ApiResponse<Alert>>(`/alert/acknowledge?id=${id}`, {
      method: "POST",
    });
    return response.data;
  }

  // Stats
  async getStats(): Promise<Stats> {
    const response = await this.request<ApiResponse<Stats>>("/stats");
    return response.data;
  }

  // Scheduler
  async startScheduler(): Promise<void> {
    await this.request("/scheduler/start", { method: "POST" });
  }

  async stopScheduler(): Promise<void> {
    await this.request("/scheduler/stop", { method: "POST" });
  }

  // Dashboard
  async getDashboardData(): Promise<DashboardData> {
    const response = await this.request<ApiResponse<DashboardData>>("/dashboard");
    return response.data;
  }

  // Reports
  async generateReport(type: "uptime" | "performance" | "incident", period: "24h" | "7d" | "30d" = "24h"): Promise<Report> {
    const response = await this.request<ApiResponse<Report>>(`/report?type=${type}&period=${period}`);
    return response.data;
  }
}

export const api = new ApiClient(API_BASE_URL);

// React Query-style hooks helpers
export function getQueryKey(type: string, ...params: unknown[]) {
  return [type, ...params];
}
