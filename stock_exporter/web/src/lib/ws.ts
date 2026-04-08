// WebSocket hook for real-time tick streaming
"use client";

import { useEffect, useRef, useState, useCallback } from "react";
import type { TickData } from "@/types/stock";

interface UseTickStreamOptions {
  symbols?: string[];
  enabled?: boolean;
}

export function useTickStream({ symbols, enabled = true }: UseTickStreamOptions = {}) {
  const [ticks, setTicks] = useState<TickData[]>([]);
  const [connected, setConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout>(undefined);

  const connect = useCallback(() => {
    if (!enabled) return;

    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const params = symbols?.length ? `?symbols=${symbols.join(",")}` : "";
    const url = `${protocol}//${window.location.host}/api/ws/ticks${params}`;

    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onopen = () => {
      setConnected(true);
    };

    ws.onmessage = (event) => {
      try {
        const data: TickData[] = JSON.parse(event.data);
        setTicks(data);
      } catch {
        // ignore parse errors
      }
    };

    ws.onclose = () => {
      setConnected(false);
      wsRef.current = null;
      // Auto-reconnect after 2 seconds
      reconnectTimeoutRef.current = setTimeout(connect, 2000);
    };

    ws.onerror = () => {
      ws.close();
    };
  }, [enabled, symbols]);

  useEffect(() => {
    connect();
    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [connect]);

  return { ticks, connected };
}
