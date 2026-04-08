// TypeScript types mirroring Go structs from internal/client/tick_store.go and internal/api/api.go

export interface TickData {
  symbol: string;
  exchange: string;
  currency: string;
  last_price: number;
  open_price: number;
  high_price: number;
  low_price: number;
  close_price: number;
  change_percent: number;
  volume_traded: number;
  total_buy_quantity: number;
  total_sell_quantity: number;
  last_traded_qty: number;
  average_trade_price: number;
  bid_price: number;
  ask_price: number;
  bid_qty: number;
  ask_qty: number;
  spread: number;
  last_trade_time?: string;
  exchange_time?: string;
}

export interface ExporterStatus {
  version: string;
  exchange: string;
  instruments: number;
  uptime: string;
  kite_enabled: boolean;
  ws_clients: number;
}
