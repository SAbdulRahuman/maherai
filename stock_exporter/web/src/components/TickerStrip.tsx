"use client";

import type { TickData } from "@/types/stock";

interface TickerStripProps {
  ticks: TickData[];
}

export default function TickerStrip({ ticks }: TickerStripProps) {
  if (ticks.length === 0) return null;

  return (
    <div className="overflow-hidden bg-slate-900/80 border border-slate-800 rounded-lg mb-6">
      <div className="flex overflow-x-auto gap-0 py-2 px-2 scrollbar-hide">
        {ticks.slice(0, 50).map((tick) => {
          const isUp = tick.change_percent >= 0;
          return (
            <div
              key={tick.symbol}
              className="flex items-center gap-2 px-3 py-1 whitespace-nowrap text-xs border-r border-slate-800 last:border-r-0 min-w-fit"
            >
              <span className="font-medium text-slate-300">{tick.symbol}</span>
              <span className="text-slate-200 font-mono">
                {tick.last_price.toFixed(2)}
              </span>
              <span
                className={`font-medium ${
                  isUp ? "text-green-400" : "text-red-400"
                }`}
              >
                {isUp ? "▲" : "▼"} {Math.abs(tick.change_percent).toFixed(2)}%
              </span>
            </div>
          );
        })}
      </div>
    </div>
  );
}
