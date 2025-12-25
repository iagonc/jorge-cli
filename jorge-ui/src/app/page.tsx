"use client";

import { useState, useEffect } from "react";
import { Sidebar } from "@/components/sidebar";
import { Dashboard } from "@/components/dashboard";
import { DNSManager } from "@/components/dns-manager";
import { NetworkMonitor } from "@/components/network-monitor";
import { AlertsPanel } from "@/components/alerts-panel";
import { ReportsPanel } from "@/components/reports-panel";
import { SettingsPanel } from "@/components/settings-panel";
import { cn } from "@/lib/utils";

export type View = "dashboard" | "dns" | "network" | "alerts" | "reports" | "settings";

export default function Home() {
  const [currentView, setCurrentView] = useState<View>("dashboard");
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [isTransitioning, setIsTransitioning] = useState(false);

  const handleViewChange = (view: View) => {
    if (view === currentView) return;
    setIsTransitioning(true);
    setTimeout(() => {
      setCurrentView(view);
      setIsTransitioning(false);
    }, 100);
  };

  const renderView = () => {
    switch (currentView) {
      case "dashboard":
        return <Dashboard onNavigate={handleViewChange} />;
      case "dns":
        return <DNSManager />;
      case "network":
        return <NetworkMonitor />;
      case "alerts":
        return <AlertsPanel />;
      case "reports":
        return <ReportsPanel />;
      case "settings":
        return <SettingsPanel />;
      default:
        return <Dashboard onNavigate={handleViewChange} />;
    }
  };

  return (
    <div className="min-h-screen">
      <Sidebar
        currentView={currentView}
        onViewChange={handleViewChange}
        collapsed={sidebarCollapsed}
        onCollapse={setSidebarCollapsed}
      />
      <main
        className={cn(
          "transition-all duration-200 ease-out",
          sidebarCollapsed ? "ml-16" : "ml-56"
        )}
      >
        <div className="max-w-6xl mx-auto px-8 py-8">
          <div
            className={cn(
              "transition-all duration-150",
              isTransitioning ? "opacity-0 translate-y-1" : "opacity-100 translate-y-0"
            )}
          >
            {renderView()}
          </div>
        </div>
      </main>
    </div>
  );
}
