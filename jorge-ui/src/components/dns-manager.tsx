"use client";

import { useState } from "react";
import { Plus, MoreHorizontal, RefreshCw, Loader2, AlertCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import {
  useResources,
  createResource,
  deleteResource,
  updateResource,
  refreshResources,
} from "@/lib/hooks";
import { Resource } from "@/lib/api";

export function DNSManager() {
  const { resources, isLoading, isRefreshing, error } = useResources();
  const [search, setSearch] = useState("");
  const [isOpen, setIsOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [editingRecord, setEditingRecord] = useState<Resource | null>(null);

  // Form state
  const [formName, setFormName] = useState("");
  const [formDns, setFormDns] = useState("");
  const [formError, setFormError] = useState("");

  const filtered = resources.filter((r) =>
    r.name.toLowerCase().includes(search.toLowerCase()) ||
    r.dns.toLowerCase().includes(search.toLowerCase())
  );

  const resetForm = () => {
    setFormName("");
    setFormDns("");
    setFormError("");
    setEditingRecord(null);
  };

  const handleOpenChange = (open: boolean) => {
    setIsOpen(open);
    if (!open) resetForm();
  };

  const handleEdit = (record: Resource) => {
    setEditingRecord(record);
    setFormName(record.name);
    setFormDns(record.dns);
    setIsOpen(true);
  };

  const handleSubmit = async () => {
    if (!formName.trim() || !formDns.trim()) {
      setFormError("Name and DNS are required");
      return;
    }

    setIsSubmitting(true);
    setFormError("");

    try {
      if (editingRecord) {
        await updateResource(editingRecord.ID, { name: formName, dns: formDns });
      } else {
        await createResource({ name: formName, dns: formDns });
      }
      handleOpenChange(false);
    } catch (err) {
      setFormError(err instanceof Error ? err.message : "Failed to save");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDelete = async (id: number) => {
    try {
      await deleteResource(id);
    } catch (err) {
      console.error("Failed to delete:", err);
    }
  };

  // Format date
  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-medium tracking-tight">DNS Records</h1>
          <p className="text-sm text-muted-foreground mt-1">
            {isLoading ? "Loading..." : `${resources.length} records`}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="sm"
            className="h-8"
            onClick={() => refreshResources()}
            disabled={isRefreshing}
          >
            <RefreshCw className={`h-4 w-4 ${isRefreshing ? "animate-spin" : ""}`} />
          </Button>
          <Dialog open={isOpen} onOpenChange={handleOpenChange}>
            <DialogTrigger asChild>
              <Button size="sm" className="h-8">
                <Plus className="h-4 w-4 mr-1" />
                Add record
              </Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-md">
              <DialogHeader>
                <DialogTitle>
                  {editingRecord ? "Edit Record" : "Add DNS Record"}
                </DialogTitle>
              </DialogHeader>
              <div className="space-y-4 pt-4">
                {formError && (
                  <div className="flex items-center gap-2 text-sm text-red-500 bg-red-500/10 px-3 py-2 rounded-lg">
                    <AlertCircle className="h-4 w-4" />
                    {formError}
                  </div>
                )}
                <div className="space-y-2">
                  <Label className="text-xs text-muted-foreground">Name</Label>
                  <Input
                    placeholder="api.example.com"
                    value={formName}
                    onChange={(e) => setFormName(e.target.value)}
                  />
                </div>
                <div className="space-y-2">
                  <Label className="text-xs text-muted-foreground">DNS / IP Address</Label>
                  <Input
                    placeholder="192.168.1.100 or cdn.example.com"
                    value={formDns}
                    onChange={(e) => setFormDns(e.target.value)}
                  />
                </div>
                <div className="flex justify-end gap-2 pt-4">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => handleOpenChange(false)}
                    disabled={isSubmitting}
                  >
                    Cancel
                  </Button>
                  <Button size="sm" onClick={handleSubmit} disabled={isSubmitting}>
                    {isSubmitting ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : editingRecord ? (
                      "Save"
                    ) : (
                      "Add"
                    )}
                  </Button>
                </div>
              </div>
            </DialogContent>
          </Dialog>
        </div>
      </div>

      {error ? (
        <div className="glass-card rounded-lg p-8 text-center">
          <AlertCircle className="h-8 w-8 text-red-500 mx-auto mb-3" />
          <p className="text-sm text-muted-foreground mb-4">
            Failed to connect to API
          </p>
          <p className="text-xs text-muted-foreground/70 mb-4 font-mono">
            {error}
          </p>
          <Button size="sm" variant="outline" onClick={() => refreshResources()}>
            Retry
          </Button>
        </div>
      ) : (
        <>
          <Input
            placeholder="Search records..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="max-w-sm"
          />

          <div className="glass-card rounded-lg overflow-hidden">
            {isLoading ? (
              <div className="flex items-center justify-center py-12">
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
              </div>
            ) : filtered.length === 0 ? (
              <div className="text-center py-12">
                <p className="text-sm text-muted-foreground">
                  {search ? "No records match your search" : "No DNS records yet"}
                </p>
                {!search && (
                  <Button
                    size="sm"
                    variant="ghost"
                    className="mt-2"
                    onClick={() => setIsOpen(true)}
                  >
                    Add your first record
                  </Button>
                )}
              </div>
            ) : (
              <table className="w-full">
                <thead>
                  <tr className="border-b border-border bg-white/[0.02]">
                    <th className="text-left text-xs font-medium text-muted-foreground px-4 py-3">
                      Name
                    </th>
                    <th className="text-left text-xs font-medium text-muted-foreground px-4 py-3">
                      DNS / IP
                    </th>
                    <th className="text-left text-xs font-medium text-muted-foreground px-4 py-3 w-36">
                      Updated
                    </th>
                    <th className="w-10"></th>
                  </tr>
                </thead>
                <tbody>
                  {filtered.map((record) => (
                    <tr
                      key={record.ID}
                      className="border-b border-border last:border-0 hover:bg-white/[0.02] transition-colors duration-150"
                    >
                      <td className="px-4 py-3 text-sm">{record.name}</td>
                      <td className="px-4 py-3 text-sm font-mono text-muted-foreground truncate max-w-xs">
                        {record.dns}
                      </td>
                      <td className="px-4 py-3 text-xs text-muted-foreground">
                        {formatDate(record.UpdatedAt)}
                      </td>
                      <td className="px-4 py-3">
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                              <MoreHorizontal className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem onClick={() => handleEdit(record)}>
                              Edit
                            </DropdownMenuItem>
                            <DropdownMenuItem
                              className="text-red-500"
                              onClick={() => handleDelete(record.ID)}
                            >
                              Delete
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </>
      )}
    </div>
  );
}
