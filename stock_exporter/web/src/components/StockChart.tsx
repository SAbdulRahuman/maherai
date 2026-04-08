"use client";

import { useEffect, useRef, useCallback } from "react";
import type { TickData } from "@/types/stock";

// Dynamic import to avoid SSR issues with lightweight-charts
let createChartModule: typeof import("lightweight-charts") | null = null;

interface StockChartProps {
  symbol: string;
  ticks: TickData[];
}

export default function StockChart({ symbol, ticks }: StockChartProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<ReturnType<typeof import("lightweight-charts").createChart> | null>(null);
  const seriesRef = useRef<ReturnType<ReturnType<typeof import("lightweight-charts").createChart>["addSeries"]> | null>(null);
  const dataRef = useRef<{ time: number; value: number }[]>([]);

  const initChart = useCallback(async () => {
    if (!containerRef.current) return;

    if (!createChartModule) {
      createChartModule = await import("lightweight-charts");
    }

    const { createChart, ColorType, LineStyle } = createChartModule;

    // Clean up existing chart
    if (chartRef.current) {
      chartRef.current.remove();
    }

    const chart = createChart(containerRef.current, {
      width: containerRef.current.clientWidth,
      height: 280,
      layout: {
        background: { type: ColorType.Solid, color: "#1e293b" },
        textColor: "#94a3b8",
        fontSize: 11,
      },
      grid: {
        vertLines: { color: "#334155", style: LineStyle.Dotted },
        horzLines: { color: "#334155", style: LineStyle.Dotted },
      },
      crosshair: {
        vertLine: { color: "#475569", labelBackgroundColor: "#334155" },
        horzLine: { color: "#475569", labelBackgroundColor: "#334155" },
      },
      rightPriceScale: {
        borderColor: "#334155",
      },
      timeScale: {
        borderColor: "#334155",
        timeVisible: true,
        secondsVisible: false,
      },
    });

    const series = chart.addSeries(createChartModule.LineSeries, {
      color: "#38bdf8",
      lineWidth: 2,
      priceFormat: { type: "price", precision: 2, minMove: 0.01 },
    });

    chartRef.current = chart;
    seriesRef.current = series;

    // Handle resize
    const resizeObserver = new ResizeObserver((entries) => {
      for (const entry of entries) {
        chart.applyOptions({ width: entry.contentRect.width });
      }
    });
    resizeObserver.observe(containerRef.current);

    return () => resizeObserver.disconnect();
  }, []);

  useEffect(() => {
    initChart();
    return () => {
      if (chartRef.current) {
        chartRef.current.remove();
        chartRef.current = null;
      }
    };
  }, [initChart]);

  // Update data when ticks change
  useEffect(() => {
    const tick = ticks.find((t) => t.symbol === symbol);
    if (!tick || !seriesRef.current) return;

    const now = Math.floor(Date.now() / 1000);
    dataRef.current.push({ time: now, value: tick.last_price });

    // Keep last 500 data points
    if (dataRef.current.length > 500) {
      dataRef.current = dataRef.current.slice(-500);
    }

    seriesRef.current.setData(
      dataRef.current as { time: import("lightweight-charts").UTCTimestamp; value: number }[]
    );
  }, [ticks, symbol]);

  return (
    <div className="bg-slate-800 border border-slate-700 rounded-xl overflow-hidden">
      <div className="px-4 py-2 border-b border-slate-700 flex items-center justify-between">
        <span className="font-semibold text-sm text-slate-200">{symbol}</span>
        {(() => {
          const tick = ticks.find((t) => t.symbol === symbol);
          if (!tick) return null;
          const isUp = tick.change_percent >= 0;
          return (
            <div className="flex items-center gap-3">
              <span className="text-lg font-bold text-slate-100">
                {tick.currency === "INR" ? "₹" : ""}{tick.last_price.toFixed(2)}
              </span>
              <span
                className={`text-sm font-medium ${
                  isUp ? "text-green-400" : "text-red-400"
                }`}
              >
                {isUp ? "+" : ""}{tick.change_percent.toFixed(2)}%
              </span>
            </div>
          );
        })()}
      </div>
      <div ref={containerRef} />
    </div>
  );
}
