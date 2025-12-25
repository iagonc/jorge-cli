"use client";

import useSWR, { mutate } from "swr";
import {
  api,
  Resource,
  Monitor,
  CheckResult,
  Alert,
  Stats,
  DiagnosticResult,
  MonitorType,
  DashboardData,
  Report,
} from "./api";

// Fetcher functions
const resourcesFetcher = () => api.listResources();
const resourceFetcher = (id: number) => api.getResource(id);
const healthFetcher = () => api.healthCheck();

// Resources hooks
export function useResources() {
  const { data, error, isLoading, isValidating } = useSWR<Resource[]>(
    "resources",
    resourcesFetcher,
    {
      refreshInterval: 5000, // Refresh every 5 seconds
      revalidateOnFocus: true,
      dedupingInterval: 2000,
    }
  );

  return {
    resources: data || [],
    isLoading,
    isRefreshing: isValidating && !isLoading,
    error: error?.message,
  };
}

export function useResource(id: number | null) {
  const { data, error, isLoading } = useSWR<Resource>(
    id ? ["resource", id] : null,
    () => resourceFetcher(id!),
    {
      revalidateOnFocus: true,
    }
  );

  return {
    resource: data,
    isLoading,
    error: error?.message,
  };
}

// Health check hook
export function useHealthCheck() {
  const { data, error } = useSWR("health", healthFetcher, {
    refreshInterval: 10000,
    revalidateOnFocus: true,
  });

  return {
    isHealthy: data === true,
    isChecking: data === undefined && !error,
    error: error?.message,
  };
}

// Mutation helpers
export async function createResource(data: { name: string; dns: string }) {
  const resource = await api.createResource(data);
  await mutate("resources");
  return resource;
}

export async function updateResource(
  id: number,
  data: { name: string; dns: string }
) {
  const resource = await api.updateResource(id, data);
  await mutate("resources");
  await mutate(["resource", id]);
  return resource;
}

export async function deleteResource(id: number) {
  await api.deleteResource(id);
  await mutate("resources");
}

// Manual refresh
export function refreshResources() {
  return mutate("resources");
}

// ============================================
// MONITORS
// ============================================

const monitorsFetcher = () => api.listMonitors();
const monitorFetcher = (id: number) => api.getMonitor(id);
const monitorResultsFetcher = (id: number, limit: number) =>
  api.getMonitorResults(id, limit);

export function useMonitors() {
  const { data, error, isLoading, isValidating } = useSWR<Monitor[]>(
    "monitors",
    monitorsFetcher,
    {
      refreshInterval: 5000,
      revalidateOnFocus: true,
      dedupingInterval: 2000,
    }
  );

  return {
    monitors: data || [],
    isLoading,
    isRefreshing: isValidating && !isLoading,
    error: error?.message,
  };
}

export function useMonitor(id: number | null) {
  const { data, error, isLoading } = useSWR<Monitor>(
    id ? ["monitor", id] : null,
    () => monitorFetcher(id!),
    {
      refreshInterval: 5000,
      revalidateOnFocus: true,
    }
  );

  return {
    monitor: data,
    isLoading,
    error: error?.message,
  };
}

export function useMonitorResults(id: number | null, limit: number = 50) {
  const { data, error, isLoading } = useSWR<CheckResult[]>(
    id ? ["monitor-results", id, limit] : null,
    () => monitorResultsFetcher(id!, limit),
    {
      refreshInterval: 10000,
      revalidateOnFocus: true,
    }
  );

  return {
    results: data || [],
    isLoading,
    error: error?.message,
  };
}

// Monitor mutations
export async function createMonitor(data: {
  name: string;
  target: string;
  type: MonitorType;
  port?: number;
  interval?: number;
  timeout?: number;
  expected_status?: number;
  method?: string;
  enabled?: boolean;
}) {
  const monitor = await api.createMonitor(data);
  await mutate("monitors");
  return monitor;
}

export async function updateMonitor(id: number, data: Partial<Monitor>) {
  const monitor = await api.updateMonitor(id, data);
  await mutate("monitors");
  await mutate(["monitor", id]);
  return monitor;
}

export async function deleteMonitor(id: number) {
  await api.deleteMonitor(id);
  await mutate("monitors");
}

export async function toggleMonitor(id: number) {
  const monitor = await api.toggleMonitor(id);
  await mutate("monitors");
  await mutate(["monitor", id]);
  return monitor;
}

export function refreshMonitors() {
  return mutate("monitors");
}

// ============================================
// ALERTS
// ============================================

const alertsFetcher = (filters?: { acknowledged?: boolean }) =>
  api.listAlerts(filters);

export function useAlerts(filters?: { acknowledged?: boolean }) {
  const key = filters ? ["alerts", filters] : "alerts";
  const { data, error, isLoading, isValidating } = useSWR<Alert[]>(
    key,
    () => alertsFetcher(filters),
    {
      refreshInterval: 10000,
      revalidateOnFocus: true,
    }
  );

  return {
    alerts: data || [],
    isLoading,
    isRefreshing: isValidating && !isLoading,
    error: error?.message,
  };
}

export async function acknowledgeAlert(id: number) {
  const alert = await api.acknowledgeAlert(id);
  await mutate("alerts");
  await mutate(["alerts", { acknowledged: false }]);
  return alert;
}

export function refreshAlerts() {
  return mutate("alerts");
}

// ============================================
// STATS
// ============================================

const statsFetcher = () => api.getStats();

export function useStats() {
  const { data, error, isLoading } = useSWR<Stats>("stats", statsFetcher, {
    refreshInterval: 5000,
    revalidateOnFocus: true,
  });

  return {
    stats: data,
    isLoading,
    error: error?.message,
  };
}

export function refreshStats() {
  return mutate("stats");
}

// ============================================
// DIAGNOSTICS
// ============================================

export async function runDiagnostic(data: {
  type: MonitorType;
  target: string;
  port?: number;
  timeout?: number;
}): Promise<DiagnosticResult> {
  return api.runDiagnostic(data);
}

// ============================================
// SCHEDULER
// ============================================

export async function startScheduler() {
  await api.startScheduler();
  await mutate("stats");
}

export async function stopScheduler() {
  await api.stopScheduler();
  await mutate("stats");
}

// ============================================
// DASHBOARD
// ============================================

const dashboardFetcher = () => api.getDashboardData();

export function useDashboard() {
  const { data, error, isLoading } = useSWR<DashboardData>(
    "dashboard",
    dashboardFetcher,
    {
      refreshInterval: 10000,
      revalidateOnFocus: true,
    }
  );

  return {
    dashboard: data,
    isLoading,
    error: error?.message,
  };
}

export function refreshDashboard() {
  return mutate("dashboard");
}

// ============================================
// REPORTS
// ============================================

export async function generateReport(
  type: "uptime" | "performance" | "incident",
  period: "24h" | "7d" | "30d" = "24h"
): Promise<Report> {
  return api.generateReport(type, period);
}
