"use client";

import { useEffect, useState } from "react";
import { getStatus } from "@/lib/api";
import type { ExporterStatus } from "@/types/stock";

export default function StatusBadge() {
  const [status, setStatus] = useState<ExporterStatus | null>(null);

  useEffect(() => {
    const fetchStatus = async () => {
      try {
        const s = await getStatus();
        setStatus(s);
      } catch {
        setStatus(null);
      }
    };
    fetchStatus();
    const interval = setInterval(fetchStatus, 5000);
    return () => clearInterval(interval);
  }, []);

  if (!status) {
    return (
      <span className="inline-flex items-center gap-1.5 text-xs px-2 py-1 rounded-full bg-red-900/50 text-red-400">
        <span className="w-2 h-2 rounded-full bg-red-500 animate-pulse" />
        Disconnected
      </span>
    );
  }

  return (
    <div className="flex items-center gap-3">
      <span className="inline-flex items-center gap-1.5 text-xs px-2 py-1 rounded-full bg-green-900/50 text-green-400">
        <span className="w-2 h-2 rounded-full bg-green-500" />
        {status.exchange}
      </span>
      <span className="text-xs text-slate-400">
        {status.instruments} instruments
      </span>
      <span className="text-xs text-slate-500">v{status.version}</span>
    </div>
  );
}
