"use client";

import { useState } from "react";
import { Plus, Pause, Play, Trash2, RotateCw, Loader2, Zap } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  useMonitors,
  useStats,
  createMonitor,
  deleteMonitor,
  toggleMonitor,
  refreshMonitors,
  startScheduler,
  stopScheduler,
  runDiagnostic,
} from "@/lib/hooks";
import type { MonitorType, DiagnosticResult } from "@/lib/api";

export function NetworkMonitor() {
  const { monitors, isLoading, isRefreshing } = useMonitors();
  const { stats } = useStats();
  const [isOpen, setIsOpen] = useState(false);
  const [isDiagOpen, setIsDiagOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [diagResult, setDiagResult] = useState<DiagnosticResult | null>(null);
  const [diagLoading, setDiagLoading] = useState(false);

  // Form state
  const [formData, setFormData] = useState({
    name: "",
    target: "",
    type: "http" as MonitorType,
    port: "",
    interval: "30",
  });

  const [diagData, setDiagData] = useState({
    type: "ping" as MonitorType,
    target: "",
    port: "",
  });

  const handleSubmit = async () => {
    if (!formData.name || !formData.target) return;

    setIsSubmitting(true);
    try {
      await createMonitor({
        name: formData.name,
        target: formData.target,
        type: formData.type,
        port: formData.port ? parseInt(formData.port) : undefined,
        interval: parseInt(formData.interval),
        enabled: true,
      });
      setIsOpen(false);
      setFormData({ name: "", target: "", type: "http", port: "", interval: "30" });
    } catch (error) {
      console.error("Failed to create monitor:", error);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDelete = async (id: number) => {
    try {
      await deleteMonitor(id);
    } catch (error) {
      console.error("Failed to delete monitor:", error);
    }
  };

  const handleToggle = async (id: number) => {
    try {
      await toggleMonitor(id);
    } catch (error) {
      console.error("Failed to toggle monitor:", error);
    }
  };

  const handleDiagnostic = async () => {
    if (!diagData.target) return;

    setDiagLoading(true);
    setDiagResult(null);
    try {
      const result = await runDiagnostic({
        type: diagData.type,
        target: diagData.target,
        port: diagData.port ? parseInt(diagData.port) : undefined,
        timeout: 10,
      });
      setDiagResult(result);
    } catch (error) {
      console.error("Failed to run diagnostic:", error);
    } finally {
      setDiagLoading(false);
    }
  };

  const handleSchedulerToggle = async () => {
    try {
      if (stats?.scheduler_running) {
        await stopScheduler();
      } else {
        await startScheduler();
      }
    } catch (error) {
      console.error("Failed to toggle scheduler:", error);
    }
  };

  const healthyCount = monitors.filter(m => m.status === "up" && m.enabled).length;
  const enabledCount = monitors.filter(m => m.enabled).length;

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-medium tracking-tight">Monitors</h1>
          <p className="text-sm text-muted-foreground mt-1">
            {healthyCount} of {enabledCount} healthy
            {stats?.scheduler_running && (
              <span className="ml-2 text-emerald-500/70">• Live</span>
            )}
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            className="h-8"
            onClick={() => refreshMonitors()}
            disabled={isRefreshing}
          >
            <RotateCw className={`h-4 w-4 ${isRefreshing ? "animate-spin" : ""}`} />
          </Button>
          <Button
            variant="outline"
            size="sm"
            className="h-8"
            onClick={handleSchedulerToggle}
          >
            {stats?.scheduler_running ? (
              <Pause className="h-4 w-4" />
            ) : (
              <Play className="h-4 w-4" />
            )}
          </Button>

          {/* Diagnostic Dialog */}
          <Dialog open={isDiagOpen} onOpenChange={setIsDiagOpen}>
            <DialogTrigger asChild>
              <Button variant="outline" size="sm" className="h-8">
                <Zap className="h-4 w-4 mr-1" />
                Quick test
              </Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-md">
              <DialogHeader>
                <DialogTitle>Run Diagnostic</DialogTitle>
              </DialogHeader>
              <div className="space-y-4 pt-4">
                <div className="space-y-2">
                  <Label className="text-xs text-muted-foreground">Target</Label>
                  <Input
                    placeholder="google.com or 8.8.8.8"
                    value={diagData.target}
                    onChange={(e) => setDiagData({ ...diagData, target: e.target.value })}
                  />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label className="text-xs text-muted-foreground">Type</Label>
                    <Select
                      value={diagData.type}
                      onValueChange={(v) => setDiagData({ ...diagData, type: v as MonitorType })}
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="ping">Ping</SelectItem>
                        <SelectItem value="dns">DNS</SelectItem>
                        <SelectItem value="tcp">TCP</SelectItem>
                        <SelectItem value="http">HTTP</SelectItem>
                        <SelectItem value="ssl">SSL</SelectItem>
                        <SelectItem value="trace">Traceroute</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  {(diagData.type === "tcp" || diagData.type === "ssl") && (
                    <div className="space-y-2">
                      <Label className="text-xs text-muted-foreground">Port</Label>
                      <Input
                        placeholder="443"
                        value={diagData.port}
                        onChange={(e) => setDiagData({ ...diagData, port: e.target.value })}
                      />
                    </div>
                  )}
                </div>

                {diagResult && (
                  <div className={`p-3 rounded-lg text-sm ${diagResult.success ? "bg-emerald-500/10 border border-emerald-500/20" : "bg-red-500/10 border border-red-500/20"}`}>
                    <div className="flex items-center justify-between mb-2">
                      <span className={diagResult.success ? "text-emerald-500" : "text-red-500"}>
                        {diagResult.success ? "Success" : "Failed"}
                      </span>
                      {diagResult.latency_ms > 0 && (
                        <span className="font-mono text-xs">{diagResult.latency_ms}ms</span>
                      )}
                    </div>
                    <pre className="text-xs text-muted-foreground whitespace-pre-wrap max-h-40 overflow-auto font-mono">
                      {diagResult.output || diagResult.error}
                    </pre>
                  </div>
                )}

                <div className="flex justify-end gap-2 pt-2">
                  <Button variant="ghost" size="sm" onClick={() => setIsDiagOpen(false)}>
                    Close
                  </Button>
                  <Button size="sm" onClick={handleDiagnostic} disabled={diagLoading || !diagData.target}>
                    {diagLoading ? <Loader2 className="h-4 w-4 animate-spin" /> : "Run"}
                  </Button>
                </div>
              </div>
            </DialogContent>
          </Dialog>

          {/* Add Monitor Dialog */}
          <Dialog open={isOpen} onOpenChange={setIsOpen}>
            <DialogTrigger asChild>
              <Button size="sm" className="h-8">
                <Plus className="h-4 w-4 mr-1" />
                Add monitor
              </Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-md">
              <DialogHeader>
                <DialogTitle>Add Monitor</DialogTitle>
              </DialogHeader>
              <div className="space-y-4 pt-4">
                <div className="space-y-2">
                  <Label className="text-xs text-muted-foreground">Name</Label>
                  <Input
                    placeholder="Production API"
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  />
                </div>
                <div className="space-y-2">
                  <Label className="text-xs text-muted-foreground">Target</Label>
                  <Input
                    placeholder="api.example.com or https://api.example.com"
                    value={formData.target}
                    onChange={(e) => setFormData({ ...formData, target: e.target.value })}
                  />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label className="text-xs text-muted-foreground">Type</Label>
                    <Select
                      value={formData.type}
                      onValueChange={(v) => setFormData({ ...formData, type: v as MonitorType })}
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="http">HTTP</SelectItem>
                        <SelectItem value="tcp">TCP</SelectItem>
                        <SelectItem value="dns">DNS</SelectItem>
                        <SelectItem value="ping">Ping</SelectItem>
                        <SelectItem value="ssl">SSL</SelectItem>
                        <SelectItem value="trace">Traceroute</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="space-y-2">
                    <Label className="text-xs text-muted-foreground">Interval</Label>
                    <Select
                      value={formData.interval}
                      onValueChange={(v) => setFormData({ ...formData, interval: v })}
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="10">10s</SelectItem>
                        <SelectItem value="30">30s</SelectItem>
                        <SelectItem value="60">1m</SelectItem>
                        <SelectItem value="300">5m</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                {(formData.type === "tcp" || formData.type === "ssl") && (
                  <div className="space-y-2">
                    <Label className="text-xs text-muted-foreground">Port</Label>
                    <Input
                      placeholder="443"
                      value={formData.port}
                      onChange={(e) => setFormData({ ...formData, port: e.target.value })}
                    />
                  </div>
                )}
                <div className="flex justify-end gap-2 pt-4">
                  <Button variant="ghost" size="sm" onClick={() => setIsOpen(false)}>
                    Cancel
                  </Button>
                  <Button size="sm" onClick={handleSubmit} disabled={isSubmitting}>
                    {isSubmitting ? <Loader2 className="h-4 w-4 animate-spin" /> : "Add"}
                  </Button>
                </div>
              </div>
            </DialogContent>
          </Dialog>
        </div>
      </div>

      {isLoading ? (
        <div className="flex items-center justify-center py-16">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      ) : monitors.length === 0 ? (
        <div className="text-center py-16">
          <p className="text-muted-foreground">No monitors configured</p>
          <p className="text-sm text-muted-foreground/70 mt-1">
            Add a monitor to start tracking your endpoints
          </p>
        </div>
      ) : (
        <div className="grid gap-3">
          {monitors.map((monitor) => (
            <div
              key={monitor.ID}
              className={`flex items-center justify-between p-4 glass-card rounded-lg hover-lift ${!monitor.enabled ? "opacity-50" : ""}`}
            >
              <div className="flex items-center gap-4">
                <div
                  className={`h-2 w-2 rounded-full ${
                    !monitor.enabled ? "bg-muted-foreground/30" :
                    monitor.status === "up" ? "bg-emerald-500" :
                    monitor.status === "degraded" ? "bg-amber-500" :
                    monitor.status === "down" ? "bg-red-500" : "bg-muted-foreground/50"
                  }`}
                />
                <div>
                  <p className="text-sm font-medium">{monitor.name}</p>
                  <p className="text-xs text-muted-foreground font-mono">{monitor.target}</p>
                </div>
              </div>
              <div className="flex items-center gap-6 text-sm">
                <div className="text-right">
                  <p className="font-mono">
                    {monitor.status === "up" && monitor.last_latency > 0
                      ? `${monitor.last_latency}ms`
                      : "—"}
                  </p>
                  <p className="text-xs text-muted-foreground">latency</p>
                </div>
                <div className="text-right">
                  <p className="font-mono">{monitor.uptime.toFixed(1)}%</p>
                  <p className="text-xs text-muted-foreground">uptime</p>
                </div>
                <span className="text-xs text-muted-foreground uppercase w-10 text-right">
                  {monitor.type}
                </span>
                <div className="flex gap-1">
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-7 w-7 p-0"
                    onClick={() => handleToggle(monitor.ID)}
                  >
                    {monitor.enabled ? (
                      <Pause className="h-3.5 w-3.5" />
                    ) : (
                      <Play className="h-3.5 w-3.5" />
                    )}
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-7 w-7 p-0 text-red-500/70 hover:text-red-500"
                    onClick={() => handleDelete(monitor.ID)}
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                  </Button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
