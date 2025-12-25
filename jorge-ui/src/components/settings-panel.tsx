"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

export function SettingsPanel() {
  const [saved, setSaved] = useState(false);

  const handleSave = () => {
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
  };

  return (
    <div className="space-y-12">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-medium tracking-tight">Settings</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Manage preferences
          </p>
        </div>
        <Button size="sm" className="h-8" onClick={handleSave}>
          {saved ? "Saved" : "Save changes"}
        </Button>
      </div>

      <div className="max-w-xl space-y-6">
        <div className="glass-card rounded-lg p-5 space-y-4">
          <h2 className="text-sm font-medium">General</h2>
          <div className="grid gap-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label className="text-xs text-muted-foreground">Timezone</Label>
                <Select defaultValue="utc-3">
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="utc-3">America/Sao_Paulo</SelectItem>
                    <SelectItem value="utc-5">America/New_York</SelectItem>
                    <SelectItem value="utc">UTC</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label className="text-xs text-muted-foreground">Check interval</Label>
                <Select defaultValue="30">
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="10">10 seconds</SelectItem>
                    <SelectItem value="30">30 seconds</SelectItem>
                    <SelectItem value="60">1 minute</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
          </div>
        </div>

        <div className="glass-card rounded-lg p-5 space-y-4">
          <h2 className="text-sm font-medium">Notifications</h2>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm">Email notifications</p>
                <p className="text-xs text-muted-foreground">admin@example.com</p>
              </div>
              <Switch defaultChecked />
            </div>
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm">Slack</p>
                <p className="text-xs text-muted-foreground">#alerts</p>
              </div>
              <Switch defaultChecked />
            </div>
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm">Critical alerts only</p>
                <p className="text-xs text-muted-foreground">Reduce notification volume</p>
              </div>
              <Switch />
            </div>
          </div>
        </div>

        <div className="glass-card rounded-lg p-5 space-y-4">
          <h2 className="text-sm font-medium">API</h2>
          <div className="space-y-2">
            <Label className="text-xs text-muted-foreground">API Key</Label>
            <div className="flex gap-2">
              <Input
                type="password"
                value="sk-jorge-xxxx-xxxx-xxxx"
                className="font-mono"
                readOnly
              />
              <Button variant="outline" size="sm">
                Regenerate
              </Button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
