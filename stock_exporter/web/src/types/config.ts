// TypeScript types mirroring Go config/config.go

export interface KiteConfig {
  api_key: string;
  api_secret: string;
  access_token: string;
  request_token: string;
  ticker_mode: string;
  currency: string;
  max_reconnect_attempts: number;
  reconnect_interval: string;
}

export interface RedPandaConfig {
  enabled: boolean;
  brokers: string[];
  topic: string;
  batch_size: number;
  linger_ms: number;
  compression: string;
  buffer_size: number;
}

export interface AppConfig {
  listen_address: string;
  metrics_path: string;
  exchange: string;
  kite: KiteConfig;
  redpanda: RedPandaConfig;
  stock_api_url: string;
  api_key: string;
  api_secret: string;
  scrape_interval: string;
  scrape_timeout: string;
  symbols: string[];
}

// ─── Reconfiguration Status (mirrors client.ReconfigStatus) ────────────────

export type ReconfigState = "idle" | "applying" | "ready" | "error";

export interface ConfigApplyStatus {
  state: ReconfigState;
  current_step: string;
  completed_steps: string[];
  error?: string;
  started_at?: string;
  finished_at?: string;
}
