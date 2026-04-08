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

export interface AppConfig {
  listen_address: string;
  metrics_path: string;
  exchange: string;
  kite: KiteConfig;
  stock_api_url: string;
  api_key: string;
  api_secret: string;
  scrape_interval: string;
  scrape_timeout: string;
  symbols: string[];
}
