"use client";

import { useState } from "react";
import { Check, Loader2, RotateCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useAlerts, acknowledgeAlert, refreshAlerts } from "@/lib/hooks";

export function AlertsPanel() {
  const [filter, setFilter] = useState<"all" | "active">("all");
  const { alerts, isLoading, isRefreshing } = useAlerts(
    filter === "active" ? { acknowledged: false } : undefined
  );

  const activeCount = alerts.filter(a => !a.acknowledged).length;

  const handleAcknowledge = async (id: number) => {
    try {
      await acknowledgeAlert(id);
    } catch (error) {
      console.error("Failed to acknowledge alert:", error);
    }
  };

  // Format time difference
  const formatTimeAgo = (dateStr: string) => {
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMins / 60);

    if (diffMins < 1) return "just now";
    if (diffMins < 60) return `${diffMins} min ago`;
    if (diffHours < 24) return `${diffHours} hour${diffHours > 1 ? "s" : ""} ago`;
    return date.toLocaleDateString();
  };

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-medium tracking-tight">Alerts</h1>
          <p className="text-sm text-muted-foreground mt-1">
            {activeCount} active
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            className="h-8"
            onClick={() => refreshAlerts()}
            disabled={isRefreshing}
          >
            <RotateCw className={`h-4 w-4 ${isRefreshing ? "animate-spin" : ""}`} />
          </Button>
          <div className="flex gap-1">
            <Button
              variant={filter === "all" ? "secondary" : "ghost"}
              size="sm"
              className="h-8"
              onClick={() => setFilter("all")}
            >
              All
            </Button>
            <Button
              variant={filter === "active" ? "secondary" : "ghost"}
              size="sm"
              className="h-8"
              onClick={() => setFilter("active")}
            >
              Active
            </Button>
          </div>
        </div>
      </div>

      {isLoading ? (
        <div className="flex items-center justify-center py-16">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      ) : (
        <div className="space-y-2">
          {alerts.map((alert) => (
            <div
              key={alert.ID}
              className={`flex items-start gap-4 p-4 glass-card rounded-lg transition-all duration-200 ${
                alert.acknowledged ? "opacity-40" : "hover-lift"
              }`}
            >
              <div
                className={`h-1.5 w-1.5 rounded-full mt-2 ${
                  alert.severity === "critical" ? "bg-red-500" :
                  alert.severity === "warning" ? "bg-amber-500" : "bg-muted-foreground"
                }`}
              />
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium">{alert.title}</p>
                <p className="text-xs text-muted-foreground mt-0.5">{alert.message}</p>
                <div className="flex gap-4 mt-2 text-xs text-muted-foreground">
                  {alert.Monitor && (
                    <span className="font-mono">{alert.Monitor.name}</span>
                  )}
                  <span>{formatTimeAgo(alert.CreatedAt)}</span>
                  <span className={`uppercase ${
                    alert.severity === "critical" ? "text-red-500/70" :
                    alert.severity === "warning" ? "text-amber-500/70" : ""
                  }`}>
                    {alert.severity}
                  </span>
                </div>
              </div>
              <div className="flex gap-1">
                {!alert.acknowledged && (
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-7 w-7 p-0"
                    onClick={() => handleAcknowledge(alert.ID)}
                    title="Acknowledge"
                  >
                    <Check className="h-4 w-4" />
                  </Button>
                )}
                {alert.acknowledged && alert.acked_at && (
                  <span className="text-xs text-muted-foreground">
                    Acked {formatTimeAgo(alert.acked_at)}
                  </span>
                )}
              </div>
            </div>
          ))}

          {alerts.length === 0 && (
            <div className="text-center py-16">
              <p className="text-muted-foreground">No alerts</p>
              <p className="text-sm text-muted-foreground/70 mt-1">
                {filter === "active"
                  ? "All alerts have been acknowledged"
                  : "Your monitors haven't generated any alerts yet"}
              </p>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
