"use client";

import { useEffect, useState, useMemo } from "react";
import { getSymbols, getTicks } from "@/lib/api";
import { useTickStream } from "@/lib/ws";
import StockChart from "@/components/StockChart";
import TickerStrip from "@/components/TickerStrip";
import StockCard from "@/components/StockCard";
import type { TickData } from "@/types/stock";

const DEFAULT_CHART_SYMBOLS = ["INFY", "TCS", "WIPRO", "RELIANCE"];

export default function HomePage() {
  const [allSymbols, setAllSymbols] = useState<string[]>([]);
  const [selectedSymbols, setSelectedSymbols] = useState<string[]>(DEFAULT_CHART_SYMBOLS);
  const [searchTerm, setSearchTerm] = useState("");
  const [showSearch, setShowSearch] = useState(false);
  const [fallbackTicks, setFallbackTicks] = useState<TickData[]>([]);

  // Real-time WebSocket stream
  const { ticks: wsTicks, connected } = useTickStream({ enabled: true });

  // Use WS ticks if connected, otherwise fall back to REST polling
  const ticks = connected && wsTicks.length > 0 ? wsTicks : fallbackTicks;

  useEffect(() => {
    getSymbols().then(setAllSymbols).catch(() => {});
  }, []);

  // REST polling fallback
  useEffect(() => {
    if (connected && wsTicks.length > 0) return;

    const poll = () => {
      getTicks().then(setFallbackTicks).catch(() => {});
    };
    poll();
    const interval = setInterval(poll, 3000);
    return () => clearInterval(interval);
  }, [connected, wsTicks.length]);

  const filteredSymbols = useMemo(() => {
    if (!searchTerm) return allSymbols.slice(0, 20);
    const term = searchTerm.toUpperCase();
    return allSymbols.filter((s) => s.includes(term)).slice(0, 20);
  }, [allSymbols, searchTerm]);

  const addSymbol = (symbol: string) => {
    if (!selectedSymbols.includes(symbol)) {
      setSelectedSymbols([...selectedSymbols, symbol]);
    }
    setSearchTerm("");
    setShowSearch(false);
  };

  const removeSymbol = (symbol: string) => {
    setSelectedSymbols(selectedSymbols.filter((s) => s !== symbol));
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-slate-100">Live Market</h1>
        <div className="flex items-center gap-3">
          <span
            className={`inline-flex items-center gap-1.5 text-sm px-2.5 py-1 rounded-full ${
              connected
                ? "bg-green-900/50 text-green-400"
                : "bg-yellow-900/50 text-yellow-400"
            }`}
          >
            <span
              className={`w-2 h-2 rounded-full ${
                connected ? "bg-green-500" : "bg-yellow-500 animate-pulse"
              }`}
            />
            {connected ? "Live" : "Polling"}
          </span>
        </div>
      </div>

      {/* Ticker Strip */}
      <TickerStrip ticks={ticks} />

      {/* Symbol Selector */}
      <div className="relative">
        <div className="flex items-center gap-2 flex-wrap">
          {selectedSymbols.map((sym) => (
            <span
              key={sym}
              className="inline-flex items-center gap-1 px-3 py-1.5 bg-sky-900/50 text-sky-300 text-sm rounded-full border border-sky-800"
            >
              {sym}
              <button
                onClick={() => removeSymbol(sym)}
                className="ml-0.5 hover:text-white"
              >
                x
              </button>
            </span>
          ))}
          <div className="relative">
            <input
              type="text"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              onFocus={() => setShowSearch(true)}
              placeholder="Add symbol..."
              className="w-44 px-3 py-1.5 bg-slate-800 border border-slate-600 rounded-lg text-sm text-slate-200 placeholder-slate-500 focus:border-sky-500 focus:outline-none"
            />
            {showSearch && filteredSymbols.length > 0 && (
              <div className="absolute top-full mt-1 left-0 w-48 max-h-48 overflow-y-auto bg-slate-800 border border-slate-600 rounded-lg shadow-xl z-50">
                {filteredSymbols.map((sym) => (
                  <button
                    key={sym}
                    onClick={() => addSymbol(sym)}
                    className="w-full text-left px-3 py-2 text-sm text-slate-300 hover:bg-slate-700 hover:text-white"
                  >
                    {sym}
                  </button>
                ))}
              </div>
            )}
          </div>
          {showSearch && (
            <button
              onClick={() => { setShowSearch(false); setSearchTerm(""); }}
              className="text-sm text-slate-500 hover:text-slate-300"
            >
              Close
            </button>
          )}
        </div>
      </div>

      {/* Charts Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {selectedSymbols.map((sym) => (
          <StockChart key={sym} symbol={sym} ticks={ticks} />
        ))}
      </div>

      {/* Stock Cards Grid */}
      {ticks.length > 0 && (
        <>
          <h2 className="text-xl font-semibold text-slate-200 mt-8">
            Stock Details
          </h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            {selectedSymbols.map((sym) => {
              const tick = ticks.find((t) => t.symbol === sym);
              if (!tick) return null;
              return <StockCard key={sym} tick={tick} />;
            })}
          </div>
        </>
      )}
    </div>
  );
}
