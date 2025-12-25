"use client";

import type { View } from "@/app/page";
import { useDashboard, useStats, useAlerts } from "@/lib/hooks";
import { Loader2, TrendingUp, TrendingDown, Clock, AlertTriangle, CheckCircle, Activity } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Badge } from "@/components/ui/badge";

interface DashboardProps {
  onNavigate: (view: View) => void;
}

export function Dashboard({ onNavigate }: DashboardProps) {
  const { dashboard, isLoading } = useDashboard();
  const { stats } = useStats();
  const { alerts: rawAlerts } = useAlerts({ acknowledged: false });
  const alerts = rawAlerts || [];

  if (isLoading || !dashboard) {
    return (
      <div className="flex items-center justify-center py-24">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  // Handle null/undefined values from API with safe defaults
  const slo = dashboard.slo || { target: 99.9, current: 100, error_budget: 0.1, compliant: true };
  const percentiles = dashboard.percentiles || { p50: 0, p95: 0, p99: 0 };
  const monitors = dashboard.monitors || [];
  const incidents = dashboard.incidents || [];
  const latency_history = dashboard.latency_history || [];

  // Calculate stats
  const upCount = monitors.filter(m => m.status === "up").length;
  const downCount = monitors.filter(m => m.status === "down").length;
  const degradedCount = monitors.filter(m => m.status === "degraded").length;

  // Format time
  const formatTimeAgo = (dateStr: string) => {
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMins / 60);
    if (diffMins < 1) return "just now";
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    return date.toLocaleDateString();
  };

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-medium tracking-tight">Overview</h1>
          <p className="text-sm text-muted-foreground mt-1">
            SRE Troubleshooting Hub
            {stats?.scheduler_running && (
              <span className="ml-2 text-emerald-500/70">• Live</span>
            )}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Badge variant={slo.compliant ? "default" : "destructive"} className="text-xs">
            SLO: {slo.current.toFixed(2)}%
          </Badge>
        </div>
      </div>

      {/* Key Metrics */}
      <div className="grid grid-cols-5 gap-4">
        <Card className="glass-card border-border">
          <CardContent className="pt-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-3xl font-semibold tracking-tight text-emerald-500">{upCount}</p>
                <p className="text-xs text-muted-foreground mt-1">Healthy</p>
              </div>
              <CheckCircle className="h-5 w-5 text-emerald-500/50" />
            </div>
          </CardContent>
        </Card>

        <Card className="glass-card border-border">
          <CardContent className="pt-4">
            <div className="flex items-center justify-between">
              <div>
                <p className={`text-3xl font-semibold tracking-tight ${downCount > 0 ? "text-red-500" : ""}`}>
                  {downCount}
                </p>
                <p className="text-xs text-muted-foreground mt-1">Down</p>
              </div>
              <AlertTriangle className={`h-5 w-5 ${downCount > 0 ? "text-red-500/50" : "text-muted-foreground/30"}`} />
            </div>
          </CardContent>
        </Card>

        <Card className="glass-card border-border">
          <CardContent className="pt-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-3xl font-semibold tracking-tight">{Math.round(percentiles.p50)}</p>
                <p className="text-xs text-muted-foreground mt-1">p50 Latency</p>
              </div>
              <Activity className="h-5 w-5 text-muted-foreground/30" />
            </div>
          </CardContent>
        </Card>

        <Card className="glass-card border-border">
          <CardContent className="pt-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-3xl font-semibold tracking-tight">{Math.round(percentiles.p95)}</p>
                <p className="text-xs text-muted-foreground mt-1">p95 Latency</p>
              </div>
              <TrendingUp className="h-5 w-5 text-muted-foreground/30" />
            </div>
          </CardContent>
        </Card>

        <Card className="glass-card border-border">
          <CardContent className="pt-4">
            <div className="flex items-center justify-between">
              <div>
                <p className={`text-3xl font-semibold tracking-tight ${alerts.length > 0 ? "text-amber-500" : ""}`}>
                  {alerts.length}
                </p>
                <p className="text-xs text-muted-foreground mt-1">Active Alerts</p>
              </div>
              <Clock className="h-5 w-5 text-muted-foreground/30" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* SLO and Uptime */}
      <div className="grid grid-cols-2 gap-6">
        {/* SLO Card */}
        <Card className="glass-card border-border">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">SLO Compliance (24h)</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-4xl font-semibold tracking-tight">
                {slo.current.toFixed(2)}%
              </span>
              <div className="text-right">
                <p className="text-xs text-muted-foreground">Target: {slo.target}%</p>
                <p className={`text-sm font-medium ${slo.compliant ? "text-emerald-500" : "text-red-500"}`}>
                  {slo.compliant ? "Compliant" : "Below Target"}
                </p>
              </div>
            </div>
            <Progress
              value={Math.min(100, (slo.current / slo.target) * 100)}
              className="h-2"
            />
            <div className="flex justify-between text-xs text-muted-foreground">
              <span>Error Budget: {slo.error_budget.toFixed(3)}%</span>
              <span>
                {slo.compliant ? (
                  <span className="text-emerald-500">Within budget</span>
                ) : (
                  <span className="text-red-500">Budget exceeded</span>
                )}
              </span>
            </div>
          </CardContent>
        </Card>

        {/* Latency Chart */}
        <Card className="glass-card border-border">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Response Time (24h)</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-32 flex items-end gap-1">
              {latency_history.length > 0 ? (
                latency_history.slice(-24).map((bucket, i) => {
                  const maxLatency = Math.max(...latency_history.map(b => b.avg_latency), 1);
                  const height = (bucket.avg_latency / maxLatency) * 100;
                  const hour = new Date(bucket.hour).getHours();
                  return (
                    <div key={i} className="flex-1 flex flex-col items-center gap-1">
                      <div
                        className="w-full bg-emerald-500/60 rounded-t hover:bg-emerald-500 transition-colors"
                        style={{ height: `${Math.max(height, 4)}%` }}
                        title={`${hour}:00 - ${Math.round(bucket.avg_latency)}ms (${bucket.check_count} checks)`}
                      />
                      {i % 4 === 0 && (
                        <span className="text-[10px] text-muted-foreground">{hour}h</span>
                      )}
                    </div>
                  );
                })
              ) : (
                <div className="flex-1 flex items-center justify-center text-sm text-muted-foreground">
                  Collecting data...
                </div>
              )}
            </div>
            <div className="flex justify-between mt-2 text-xs text-muted-foreground">
              <span>Avg: {Math.round(percentiles.p50)}ms</span>
              <span>p95: {Math.round(percentiles.p95)}ms</span>
              <span>Max: {Math.round(percentiles.p99)}ms</span>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Monitor Status Grid */}
      <Card className="glass-card border-border">
        <CardHeader className="pb-2 flex flex-row items-center justify-between">
          <CardTitle className="text-sm font-medium">Monitor Status</CardTitle>
          <button
            onClick={() => onNavigate("network")}
            className="text-xs text-muted-foreground hover:text-foreground transition-colors"
          >
            View all →
          </button>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-6 gap-2">
            {monitors.map((monitor) => (
              <div
                key={monitor.ID}
                className="p-3 rounded-lg bg-background/50 border border-border hover:border-foreground/10 transition-colors cursor-pointer"
                title={`${monitor.name}\n${monitor.target}\nLatency: ${monitor.last_latency}ms`}
              >
                <div className="flex items-center gap-2 mb-1">
                  <div className={`h-2 w-2 rounded-full ${
                    monitor.status === "up" ? "bg-emerald-500" :
                    monitor.status === "degraded" ? "bg-amber-500" :
                    monitor.status === "down" ? "bg-red-500" : "bg-muted-foreground/50"
                  }`} />
                  <span className="text-xs font-medium truncate">{monitor.name}</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-[10px] text-muted-foreground uppercase">{monitor.type}</span>
                  <span className="text-xs font-mono">
                    {monitor.status === "up" ? `${monitor.last_latency}ms` : "—"}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Incidents Timeline */}
      <div className="grid grid-cols-2 gap-6">
        <Card className="glass-card border-border">
          <CardHeader className="pb-2 flex flex-row items-center justify-between">
            <CardTitle className="text-sm font-medium">Recent Incidents</CardTitle>
            <button
              onClick={() => onNavigate("alerts")}
              className="text-xs text-muted-foreground hover:text-foreground transition-colors"
            >
              View all →
            </button>
          </CardHeader>
          <CardContent>
            {incidents.length > 0 ? (
              <div className="space-y-3">
                {incidents.slice(0, 5).map((incident) => (
                  <div key={incident.id} className="flex items-start gap-3 py-2 border-b border-border last:border-0">
                    <div className={`h-1.5 w-1.5 rounded-full mt-1.5 ${
                      incident.status === "critical" ? "bg-red-500" : "bg-amber-500"
                    }`} />
                    <div className="flex-1 min-w-0">
                      <p className="text-sm truncate">{incident.message}</p>
                      <div className="flex gap-3 mt-0.5 text-xs text-muted-foreground">
                        <span className="font-mono">{incident.monitor_name}</span>
                        <span>{formatTimeAgo(incident.started_at)}</span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-8">
                <CheckCircle className="h-8 w-8 text-emerald-500/30 mx-auto mb-2" />
                <p className="text-sm text-muted-foreground">No incidents</p>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Quick Actions */}
        <Card className="glass-card border-border">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Quick Actions</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 gap-2">
              <button
                onClick={() => onNavigate("network")}
                className="p-4 rounded-lg bg-background/50 border border-border hover:border-emerald-500/50 hover:bg-emerald-500/5 transition-colors text-left"
              >
                <Activity className="h-5 w-5 text-emerald-500 mb-2" />
                <p className="text-sm font-medium">Monitors</p>
                <p className="text-xs text-muted-foreground">Manage endpoints</p>
              </button>
              <button
                onClick={() => onNavigate("alerts")}
                className="p-4 rounded-lg bg-background/50 border border-border hover:border-amber-500/50 hover:bg-amber-500/5 transition-colors text-left"
              >
                <AlertTriangle className="h-5 w-5 text-amber-500 mb-2" />
                <p className="text-sm font-medium">Alerts</p>
                <p className="text-xs text-muted-foreground">{alerts.length} active</p>
              </button>
              <button
                onClick={() => onNavigate("reports")}
                className="p-4 rounded-lg bg-background/50 border border-border hover:border-blue-500/50 hover:bg-blue-500/5 transition-colors text-left"
              >
                <TrendingUp className="h-5 w-5 text-blue-500 mb-2" />
                <p className="text-sm font-medium">Reports</p>
                <p className="text-xs text-muted-foreground">Generate reports</p>
              </button>
              <button
                onClick={() => onNavigate("dns")}
                className="p-4 rounded-lg bg-background/50 border border-border hover:border-purple-500/50 hover:bg-purple-500/5 transition-colors text-left"
              >
                <TrendingDown className="h-5 w-5 text-purple-500 mb-2" />
                <p className="text-sm font-medium">DNS Records</p>
                <p className="text-xs text-muted-foreground">Manage DNS</p>
              </button>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Footer Stats */}
      {stats && (
        <div className="pt-4 border-t border-border">
          <div className="flex items-center gap-6 text-xs text-muted-foreground">
            <span>Total Monitors: {stats.total_monitors}</span>
            <span>•</span>
            <span>Uptime: {stats.uptime_percent.toFixed(2)}%</span>
            <span>•</span>
            <span>Avg Latency: {Math.round(stats.avg_latency_ms)}ms</span>
            <span>•</span>
            <span className={stats.scheduler_running ? "text-emerald-500/70" : "text-amber-500"}>
              Scheduler: {stats.scheduler_running ? "Running" : "Stopped"}
            </span>
          </div>
        </div>
      )}
    </div>
  );
}
