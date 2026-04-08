// API client for the stock exporter backend
import type { AppConfig } from "@/types/config";
import type { TickData, ExporterStatus } from "@/types/stock";

// In production (embedded), the API is on the same origin.
// In development, proxy via next.config.ts rewrites or use env var.
const API_BASE = typeof window !== "undefined" ? window.location.origin : "";

async function fetchJSON<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { "Content-Type": "application/json" },
    ...init,
  });
  if (!res.ok) {
    const body = await res.text();
    throw new Error(`API ${res.status}: ${body}`);
  }
  return res.json();
}

export async function getConfig(): Promise<AppConfig> {
  return fetchJSON<AppConfig>("/api/config");
}

export async function saveConfig(
  config: AppConfig
): Promise<{ status: string; message: string }> {
  return fetchJSON("/api/config", {
    method: "PUT",
    body: JSON.stringify(config),
  });
}

export async function getTicks(symbols?: string[]): Promise<TickData[]> {
  const params = symbols?.length ? `?symbols=${symbols.join(",")}` : "";
  return fetchJSON<TickData[]>(`/api/ticks${params}`);
}

export async function getSymbols(): Promise<string[]> {
  return fetchJSON<string[]>("/api/symbols");
}

export async function getStatus(): Promise<ExporterStatus> {
  return fetchJSON<ExporterStatus>("/api/status");
}
