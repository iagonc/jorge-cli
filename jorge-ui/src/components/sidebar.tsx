"use client";

import { useRef, useEffect, useState } from "react";
import { cn } from "@/lib/utils";
import type { View } from "@/app/page";
import { useHealthCheck } from "@/lib/hooks";
import {
  LayoutDashboard,
  Globe,
  Activity,
  Bell,
  BarChart3,
  Settings,
  ChevronLeft,
  ChevronRight,
} from "lucide-react";

interface SidebarProps {
  currentView: View;
  onViewChange: (view: View) => void;
  collapsed: boolean;
  onCollapse: (collapsed: boolean) => void;
}

const menuItems: { id: View; label: string; icon: typeof LayoutDashboard }[] = [
  { id: "dashboard", label: "Overview", icon: LayoutDashboard },
  { id: "dns", label: "DNS", icon: Globe },
  { id: "network", label: "Monitors", icon: Activity },
  { id: "alerts", label: "Alerts", icon: Bell },
  { id: "reports", label: "Reports", icon: BarChart3 },
  { id: "settings", label: "Settings", icon: Settings },
];

export function Sidebar({ currentView, onViewChange, collapsed, onCollapse }: SidebarProps) {
  const navRef = useRef<HTMLElement>(null);
  const [indicatorStyle, setIndicatorStyle] = useState({ top: 0, opacity: 0 });
  const { isHealthy, isChecking } = useHealthCheck();

  useEffect(() => {
    const activeIndex = menuItems.findIndex((item) => item.id === currentView);
    if (activeIndex !== -1) {
      const itemHeight = 40;
      const gap = 4;
      setIndicatorStyle({
        top: activeIndex * (itemHeight + gap),
        opacity: 1,
      });
    }
  }, [currentView]);

  return (
    <aside
      className={cn(
        "fixed left-0 top-0 z-40 h-screen glass border-r border-border",
        "transition-all duration-200 ease-out",
        collapsed ? "w-16" : "w-56"
      )}
    >
      <div className="flex h-full flex-col">
        {/* Header */}
        <div className={cn(
          "flex items-center h-14 border-b border-border transition-all duration-200",
          collapsed ? "px-4 justify-center" : "px-6"
        )}>
          <div className="flex items-center gap-2 overflow-hidden">
            <div className="h-6 w-6 rounded-md border border-border flex items-center justify-center flex-shrink-0">
              <span className="text-foreground text-xs font-medium">J</span>
            </div>
            <span className={cn(
              "text-sm font-medium tracking-tight transition-all duration-200",
              collapsed ? "opacity-0 w-0" : "opacity-100"
            )}>
              jorge
            </span>
          </div>
        </div>

        {/* Navigation */}
        <nav ref={navRef} className="flex-1 py-4 px-2 relative">
          {/* Liquid glass indicator */}
          <div
            className="nav-indicator mx-1"
            style={{
              transform: `translateY(${indicatorStyle.top}px)`,
              opacity: indicatorStyle.opacity,
              width: collapsed ? 'calc(100% - 8px)' : 'calc(100% - 8px)',
            }}
          />

          {menuItems.map((item) => {
            const Icon = item.icon;
            const isActive = currentView === item.id;

            return (
              <button
                key={item.id}
                onClick={() => onViewChange(item.id)}
                className={cn(
                  "nav-item-hover flex w-full items-center gap-3 rounded-lg text-sm relative z-10",
                  "transition-all duration-150",
                  collapsed ? "px-3 py-2.5 justify-center" : "px-4 py-2.5",
                  isActive
                    ? "text-foreground"
                    : "text-muted-foreground hover:text-foreground"
                )}
                style={{ height: 40, marginBottom: 4 }}
                title={collapsed ? item.label : undefined}
              >
                <Icon className={cn(
                  "h-4 w-4 flex-shrink-0 transition-colors duration-150",
                  isActive && "text-emerald-500"
                )} />
                <span className={cn(
                  "transition-all duration-200 whitespace-nowrap",
                  collapsed ? "opacity-0 w-0 overflow-hidden" : "opacity-100"
                )}>
                  {item.label}
                </span>
              </button>
            );
          })}
        </nav>

        {/* Footer */}
        <div className="border-t border-border p-3">
          <div className={cn(
            "flex items-center gap-2 mb-3 transition-all duration-200",
            collapsed ? "justify-center" : "px-2"
          )}>
            <div className={cn(
              "h-2 w-2 rounded-full flex-shrink-0",
              isChecking ? "bg-amber-500" : isHealthy ? "bg-emerald-500 pulse-live" : "bg-red-500"
            )} />
            <span className={cn(
              "text-xs text-muted-foreground transition-all duration-200 whitespace-nowrap",
              collapsed ? "opacity-0 w-0 overflow-hidden" : "opacity-100"
            )}>
              {isChecking ? "Connecting..." : isHealthy ? "API connected" : "API offline"}
            </span>
          </div>

          {/* Collapse toggle */}
          <button
            onClick={() => onCollapse(!collapsed)}
            className={cn(
              "flex items-center gap-2 w-full py-2 rounded-lg",
              "text-muted-foreground/60 hover:text-muted-foreground hover:bg-white/[0.02]",
              "transition-all duration-150",
              collapsed ? "justify-center" : "px-3"
            )}
            title={collapsed ? "Expand" : "Collapse"}
          >
            {collapsed ? (
              <ChevronRight className="h-3.5 w-3.5" />
            ) : (
              <>
                <ChevronLeft className="h-3.5 w-3.5" />
                <span className="text-xs">Collapse</span>
              </>
            )}
          </button>
        </div>
      </div>
    </aside>
  );
}
