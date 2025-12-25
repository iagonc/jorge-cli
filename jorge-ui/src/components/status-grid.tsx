"use client";

import { CheckCircle, AlertTriangle, XCircle, Clock } from "lucide-react";

const endpoints = [
  { name: "api.example.com", status: "online", latency: "45ms", uptime: "99.99%" },
  { name: "cdn.example.com", status: "online", latency: "12ms", uptime: "100%" },
  { name: "db.example.com", status: "online", latency: "8ms", uptime: "99.95%" },
  { name: "auth.example.com", status: "warning", latency: "250ms", uptime: "98.50%" },
  { name: "mail.example.com", status: "online", latency: "89ms", uptime: "99.90%" },
  { name: "backup.example.com", status: "offline", latency: "-", uptime: "95.20%" },
];

const StatusIcon = ({ status }: { status: string }) => {
  switch (status) {
    case "online":
      return <CheckCircle className="h-4 w-4 text-green-500" />;
    case "warning":
      return <AlertTriangle className="h-4 w-4 text-yellow-500" />;
    case "offline":
      return <XCircle className="h-4 w-4 text-red-500" />;
    default:
      return null;
  }
};

export function StatusGrid() {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
      {endpoints.map((endpoint) => (
        <div
          key={endpoint.name}
          className={`p-4 rounded-xl border transition-all duration-200 hover:glow-emerald cursor-pointer ${
            endpoint.status === "online"
              ? "border-green-500/20 bg-green-500/5"
              : endpoint.status === "warning"
              ? "border-yellow-500/20 bg-yellow-500/5"
              : "border-red-500/20 bg-red-500/5"
          }`}
        >
          <div className="flex items-center justify-between mb-2">
            <span className="font-medium text-sm truncate">{endpoint.name}</span>
            <StatusIcon status={endpoint.status} />
          </div>
          <div className="flex items-center gap-4 text-xs text-muted-foreground">
            <div className="flex items-center gap-1">
              <Clock className="h-3 w-3" />
              <span>{endpoint.latency}</span>
            </div>
            <div className="flex items-center gap-1">
              <CheckCircle className="h-3 w-3" />
              <span>{endpoint.uptime}</span>
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}
