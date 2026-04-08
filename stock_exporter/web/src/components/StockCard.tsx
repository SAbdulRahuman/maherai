"use client";

import type { TickData } from "@/types/stock";

interface StockCardProps {
  tick: TickData;
}

export default function StockCard({ tick }: StockCardProps) {
  const isUp = tick.change_percent >= 0;
  const currency = tick.currency === "INR" ? "₹" : tick.currency + " ";

  return (
    <div className="bg-slate-900 border border-slate-700 rounded-xl p-4 hover:border-slate-600 transition-colors">
      <div className="flex items-center justify-between mb-3">
        <h3 className="font-bold text-base text-slate-200">{tick.symbol}</h3>
        <span
          className={`text-sm font-medium px-2 py-0.5 rounded-full ${
            isUp
              ? "bg-green-900/50 text-green-400"
              : "bg-red-900/50 text-red-400"
          }`}
        >
          {isUp ? "+" : ""}
          {tick.change_percent.toFixed(2)}%
        </span>
      </div>

      <div className="text-2xl font-bold text-slate-100 mb-3">
        {currency}{tick.last_price.toFixed(2)}
      </div>

      <div className="grid grid-cols-2 gap-x-4 gap-y-2 text-sm">
        <Row label="Open" value={`${currency}${tick.open_price.toFixed(2)}`} />
        <Row label="High" value={`${currency}${tick.high_price.toFixed(2)}`} />
        <Row label="Low" value={`${currency}${tick.low_price.toFixed(2)}`} />
        <Row label="Prev Close" value={`${currency}${tick.close_price.toFixed(2)}`} />
        <Row label="Volume" value={tick.volume_traded.toLocaleString()} />
        <Row label="VWAP" value={`${currency}${tick.average_trade_price.toFixed(2)}`} />
        <Row label="Bid" value={`${currency}${tick.bid_price.toFixed(2)} (${tick.bid_qty})`} />
        <Row label="Ask" value={`${currency}${tick.ask_price.toFixed(2)} (${tick.ask_qty})`} />
        <Row label="Spread" value={`${currency}${tick.spread.toFixed(2)}`} />
        <Row label="Buy Qty" value={tick.total_buy_quantity.toLocaleString()} />
      </div>
    </div>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between">
      <span className="text-slate-500">{label}</span>
      <span className="text-slate-300 font-mono">{value}</span>
    </div>
  );
}
