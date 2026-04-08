"use client";

import { useEffect, useState, useMemo } from "react";
import { getTicks } from "@/lib/api";
import { useTickStream } from "@/lib/ws";
import type { TickData } from "@/types/stock";

type SortKey = keyof TickData;
type SortDir = "asc" | "desc";

export default function MetricsPage() {
  const [fallbackTicks, setFallbackTicks] = useState<TickData[]>([]);
  const [search, setSearch] = useState("");
  const [sortKey, setSortKey] = useState<SortKey>("symbol");
  const [sortDir, setSortDir] = useState<SortDir>("asc");
  const [selectedSymbol, setSelectedSymbol] = useState<string | null>(null);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);

  const { ticks: wsTicks, connected } = useTickStream({ enabled: true });

  const ticks = connected && wsTicks.length > 0 ? wsTicks : fallbackTicks;

  // REST polling fallback
  useEffect(() => {
    if (connected && wsTicks.length > 0) {
      setLastUpdated(new Date());
      return;
    }
    const poll = () => {
      getTicks()
        .then((data) => {
          setFallbackTicks(data);
          setLastUpdated(new Date());
        })
        .catch(() => {});
    };
    poll();
    const interval = setInterval(poll, 2000);
    return () => clearInterval(interval);
  }, [connected, wsTicks.length]);

  // Update timestamp on WS data
  useEffect(() => {
    if (wsTicks.length > 0) setLastUpdated(new Date());
  }, [wsTicks]);

  const filtered = useMemo(() => {
    let result = ticks;
    if (search) {
      const term = search.toUpperCase();
      result = result.filter((t) => t.symbol.includes(term));
    }
    result.sort((a, b) => {
      const aVal = a[sortKey];
      const bVal = b[sortKey];
      if (typeof aVal === "string" && typeof bVal === "string") {
        return sortDir === "asc"
          ? aVal.localeCompare(bVal)
          : bVal.localeCompare(aVal);
      }
      if (typeof aVal === "number" && typeof bVal === "number") {
        return sortDir === "asc" ? aVal - bVal : bVal - aVal;
      }
      return 0;
    });
    return result;
  }, [ticks, search, sortKey, sortDir]);

  const toggleSort = (key: SortKey) => {
    if (sortKey === key) {
      setSortDir(sortDir === "asc" ? "desc" : "asc");
    } else {
      setSortKey(key);
      setSortDir("asc");
    }
  };

  const sortIcon = (key: SortKey) => {
    if (sortKey !== key) return "↕";
    return sortDir === "asc" ? "↑" : "↓";
  };

  const selectedTick = selectedSymbol
    ? ticks.find((t) => t.symbol === selectedSymbol)
    : null;

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between flex-wrap gap-4">
        <h1 className="text-3xl font-bold text-slate-100">Stock Metrics</h1>
        <div className="flex items-center gap-4">
          <span className="text-sm text-slate-500">
            {filtered.length} / {ticks.length} symbols
          </span>
          {lastUpdated && (
            <span className="text-sm text-slate-500">
              Updated: {lastUpdated.toLocaleTimeString()}
            </span>
          )}
        </div>
      </div>

      {/* Search */}
      <div className="relative">
        <input
          type="text"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search symbols (e.g. INFY, WIPRO, TCS...)"
          className="w-full max-w-md px-4 py-2.5 bg-slate-900 border border-slate-700 rounded-xl text-base text-slate-200 placeholder-slate-500 focus:border-sky-500 focus:outline-none focus:ring-1 focus:ring-sky-500"
        />
        {search && (
          <button
            onClick={() => setSearch("")}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-500 hover:text-slate-300 text-base"
          >
            Clear
          </button>
        )}
      </div>

      {/* Table + Detail Panel */}
      <div className="flex gap-4">
        {/* Main Table */}
        <div
          className={`flex-1 overflow-x-auto bg-slate-900 border border-slate-700 rounded-xl ${
            selectedTick ? "max-w-[calc(100%-320px)]" : ""
          }`}
        >
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-slate-700 text-slate-400">
                {(
                  [
                    ["symbol", "Symbol"],
                    ["exchange", "Exch"],
                    ["last_price", "Last Price"],
                    ["open_price", "Open"],
                    ["high_price", "High"],
                    ["low_price", "Low"],
                    ["close_price", "Prev Close"],
                    ["change_percent", "Change %"],
                    ["volume_traded", "Volume"],
                    ["average_trade_price", "VWAP"],
                    ["bid_price", "Bid"],
                    ["ask_price", "Ask"],
                    ["spread", "Spread"],
                  ] as [SortKey, string][]
                ).map(([key, label]) => (
                  <th
                    key={key}
                    onClick={() => toggleSort(key)}
                    className="px-3 py-2.5 text-left font-medium cursor-pointer hover:text-slate-200 whitespace-nowrap select-none"
                  >
                    {label}{" "}
                    <span className="text-slate-600">{sortIcon(key)}</span>
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {filtered.map((tick) => {
                const isUp = tick.change_percent >= 0;
                const isSelected = selectedSymbol === tick.symbol;
                return (
                  <tr
                    key={tick.symbol}
                    onClick={() =>
                      setSelectedSymbol(
                        selectedSymbol === tick.symbol ? null : tick.symbol
                      )
                    }
                    className={`border-b border-slate-800 cursor-pointer transition-colors ${
                      isSelected
                        ? "bg-sky-900/30"
                        : "hover:bg-slate-800/50"
                    }`}
                  >
                    <td className="px-3 py-2 font-medium text-slate-200">
                      {tick.symbol}
                    </td>
                    <td className="px-3 py-2 text-slate-400">{tick.exchange}</td>
                    <td className="px-3 py-2 font-mono text-slate-200">
                      {tick.last_price.toFixed(2)}
                    </td>
                    <td className="px-3 py-2 font-mono text-slate-400">
                      {tick.open_price.toFixed(2)}
                    </td>
                    <td className="px-3 py-2 font-mono text-slate-400">
                      {tick.high_price.toFixed(2)}
                    </td>
                    <td className="px-3 py-2 font-mono text-slate-400">
                      {tick.low_price.toFixed(2)}
                    </td>
                    <td className="px-3 py-2 font-mono text-slate-400">
                      {tick.close_price.toFixed(2)}
                    </td>
                    <td
                      className={`px-3 py-2 font-mono font-medium ${
                        isUp ? "text-green-400" : "text-red-400"
                      }`}
                    >
                      {isUp ? "+" : ""}
                      {tick.change_percent.toFixed(2)}%
                    </td>
                    <td className="px-3 py-2 font-mono text-slate-400">
                      {tick.volume_traded.toLocaleString()}
                    </td>
                    <td className="px-3 py-2 font-mono text-slate-400">
                      {tick.average_trade_price.toFixed(2)}
                    </td>
                    <td className="px-3 py-2 font-mono text-slate-400">
                      {tick.bid_price.toFixed(2)}
                    </td>
                    <td className="px-3 py-2 font-mono text-slate-400">
                      {tick.ask_price.toFixed(2)}
                    </td>
                    <td className="px-3 py-2 font-mono text-slate-400">
                      {tick.spread.toFixed(2)}
                    </td>
                  </tr>
                );
              })}
              {filtered.length === 0 && (
                <tr>
                  <td
                    colSpan={13}
                    className="px-3 py-8 text-center text-slate-500"
                  >
                    {search
                      ? `No symbols matching "${search}"`
                      : "No tick data available yet"}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>

        {/* Detail Drawer */}
        {selectedTick && (
          <div className="w-80 shrink-0 bg-slate-900 border border-slate-700 rounded-xl p-4 self-start sticky top-20">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-bold text-slate-100">
                {selectedTick.symbol}
              </h3>
              <button
                onClick={() => setSelectedSymbol(null)}
                className="text-slate-500 hover:text-slate-300 text-lg"
              >
                x
              </button>
            </div>

            <div
              className={`text-3xl font-bold mb-1 ${
                selectedTick.change_percent >= 0
                  ? "text-green-400"
                  : "text-red-400"
              }`}
            >
              {selectedTick.currency === "INR" ? "₹" : ""}
              {selectedTick.last_price.toFixed(2)}
            </div>
            <div
              className={`text-base mb-4 ${
                selectedTick.change_percent >= 0
                  ? "text-green-400"
                  : "text-red-400"
              }`}
            >
              {selectedTick.change_percent >= 0 ? "+" : ""}
              {selectedTick.change_percent.toFixed(2)}%
            </div>

            <div className="space-y-2 text-sm">
              <DetailRow label="Exchange" value={selectedTick.exchange} />
              <DetailRow label="Currency" value={selectedTick.currency} />
              <DetailRow label="Open" value={selectedTick.open_price.toFixed(2)} />
              <DetailRow label="High" value={selectedTick.high_price.toFixed(2)} />
              <DetailRow label="Low" value={selectedTick.low_price.toFixed(2)} />
              <DetailRow label="Prev Close" value={selectedTick.close_price.toFixed(2)} />
              <DetailRow label="Volume" value={selectedTick.volume_traded.toLocaleString()} />
              <DetailRow label="Buy Qty" value={selectedTick.total_buy_quantity.toLocaleString()} />
              <DetailRow label="Sell Qty" value={selectedTick.total_sell_quantity.toLocaleString()} />
              <DetailRow label="Last Traded Qty" value={selectedTick.last_traded_qty.toLocaleString()} />
              <DetailRow label="VWAP" value={selectedTick.average_trade_price.toFixed(2)} />
              <DetailRow label="Bid" value={`${selectedTick.bid_price.toFixed(2)} (${selectedTick.bid_qty})`} />
              <DetailRow label="Ask" value={`${selectedTick.ask_price.toFixed(2)} (${selectedTick.ask_qty})`} />
              <DetailRow label="Spread" value={selectedTick.spread.toFixed(2)} />
              {selectedTick.last_trade_time && (
                <DetailRow
                  label="Last Trade"
                  value={new Date(selectedTick.last_trade_time).toLocaleString()}
                />
              )}
              {selectedTick.exchange_time && (
                <DetailRow
                  label="Exchange Time"
                  value={new Date(selectedTick.exchange_time).toLocaleString()}
                />
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

function DetailRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between py-1 border-b border-slate-800">
      <span className="text-slate-500">{label}</span>
      <span className="text-slate-200 font-mono">{value}</span>
    </div>
  );
}
