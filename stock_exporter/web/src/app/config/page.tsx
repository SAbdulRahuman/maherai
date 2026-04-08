"use client";

import { useEffect, useState } from "react";
import { getConfig, saveConfig } from "@/lib/api";
import type { AppConfig } from "@/types/config";

export default function ConfigPage() {
  const [config, setConfig] = useState<AppConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  useEffect(() => {
    getConfig()
      .then(setConfig)
      .catch((err) =>
        setMessage({ type: "error", text: "Failed to load config: " + err.message })
      )
      .finally(() => setLoading(false));
  }, []);

  const handleSave = async () => {
    if (!config) return;
    setSaving(true);
    setMessage(null);
    try {
      const res = await saveConfig(config);
      setMessage({ type: "success", text: res.message || "Configuration saved!" });
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Unknown error";
      setMessage({ type: "error", text: msg });
    } finally {
      setSaving(false);
    }
  };

  const updateField = <K extends keyof AppConfig>(key: K, value: AppConfig[K]) => {
    if (!config) return;
    setConfig({ ...config, [key]: value });
  };

  const updateKite = <K extends keyof AppConfig["kite"]>(
    key: K,
    value: AppConfig["kite"][K]
  ) => {
    if (!config) return;
    setConfig({ ...config, kite: { ...config.kite, [key]: value } });
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-slate-400 animate-pulse">Loading configuration...</div>
      </div>
    );
  }

  if (!config) {
    return (
      <div className="text-red-400 text-center py-8">
        Failed to load configuration.
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-slate-100">Configuration</h1>
        <button
          onClick={handleSave}
          disabled={saving}
          className="px-4 py-2 bg-sky-600 hover:bg-sky-500 disabled:bg-slate-600 text-white rounded-lg text-sm font-medium transition-colors"
        >
          {saving ? "Saving..." : "Save Configuration"}
        </button>
      </div>

      {message && (
        <div
          className={`px-4 py-3 rounded-lg text-sm ${
            message.type === "success"
              ? "bg-green-900/50 text-green-300 border border-green-800"
              : "bg-red-900/50 text-red-300 border border-red-800"
          }`}
        >
          {message.text}
        </div>
      )}

      {/* Server Settings */}
      <Section title="Server Settings">
        <Field label="Listen Address">
          <input
            type="text"
            value={config.listen_address}
            onChange={(e) => updateField("listen_address", e.target.value)}
            className="input-field"
            placeholder=":9101"
          />
        </Field>
        <Field label="Metrics Path">
          <input
            type="text"
            value={config.metrics_path}
            onChange={(e) => updateField("metrics_path", e.target.value)}
            className="input-field"
            placeholder="/metrics"
          />
        </Field>
        <Field label="Exchange">
          <select
            value={config.exchange}
            onChange={(e) => updateField("exchange", e.target.value)}
            className="input-field"
          >
            <option value="NSE">NSE</option>
            <option value="TADAWUL">TADAWUL</option>
            <option value="IB">IB</option>
          </select>
        </Field>
      </Section>

      {/* Kite Connect Settings */}
      <Section title="Kite Connect Settings">
        <Field label="API Key">
          <input
            type="text"
            value={config.kite.api_key}
            onChange={(e) => updateKite("api_key", e.target.value)}
            className="input-field"
          />
        </Field>
        <Field label="API Secret">
          <input
            type="password"
            value={config.kite.api_secret}
            onChange={(e) => updateKite("api_secret", e.target.value)}
            className="input-field"
          />
        </Field>
        <Field label="Access Token">
          <input
            type="password"
            value={config.kite.access_token}
            onChange={(e) => updateKite("access_token", e.target.value)}
            className="input-field"
          />
        </Field>
        <Field label="Request Token">
          <input
            type="password"
            value={config.kite.request_token}
            onChange={(e) => updateKite("request_token", e.target.value)}
            className="input-field"
          />
        </Field>
        <Field label="Ticker Mode">
          <select
            value={config.kite.ticker_mode}
            onChange={(e) => updateKite("ticker_mode", e.target.value)}
            className="input-field"
          >
            <option value="ltp">LTP (Last Traded Price)</option>
            <option value="quote">Quote</option>
            <option value="full">Full (includes market depth)</option>
          </select>
        </Field>
        <Field label="Currency">
          <input
            type="text"
            value={config.kite.currency}
            onChange={(e) => updateKite("currency", e.target.value)}
            className="input-field"
            placeholder="INR"
          />
        </Field>
        <Field label="Max Reconnect Attempts">
          <input
            type="number"
            value={config.kite.max_reconnect_attempts}
            onChange={(e) =>
              updateKite("max_reconnect_attempts", parseInt(e.target.value) || 0)
            }
            className="input-field"
          />
        </Field>
        <Field label="Reconnect Interval">
          <input
            type="text"
            value={config.kite.reconnect_interval}
            onChange={(e) => updateKite("reconnect_interval", e.target.value)}
            className="input-field"
            placeholder="5s"
          />
        </Field>
      </Section>

      {/* Legacy REST API */}
      <Section title="Legacy REST API (Fallback)">
        <Field label="Stock API URL">
          <input
            type="text"
            value={config.stock_api_url}
            onChange={(e) => updateField("stock_api_url", e.target.value)}
            className="input-field"
          />
        </Field>
        <Field label="API Key">
          <input
            type="text"
            value={config.api_key}
            onChange={(e) => updateField("api_key", e.target.value)}
            className="input-field"
          />
        </Field>
        <Field label="API Secret">
          <input
            type="password"
            value={config.api_secret}
            onChange={(e) => updateField("api_secret", e.target.value)}
            className="input-field"
          />
        </Field>
      </Section>

      {/* Scrape Settings */}
      <Section title="Scrape Settings">
        <Field label="Scrape Interval">
          <input
            type="text"
            value={config.scrape_interval}
            onChange={(e) => updateField("scrape_interval", e.target.value)}
            className="input-field"
            placeholder="15s"
          />
        </Field>
        <Field label="Scrape Timeout">
          <input
            type="text"
            value={config.scrape_timeout}
            onChange={(e) => updateField("scrape_timeout", e.target.value)}
            className="input-field"
            placeholder="10s"
          />
        </Field>
      </Section>

      {/* Symbols Watchlist */}
      <Section title="Symbols Watchlist">
        <p className="text-xs text-slate-400 mb-2">
          One symbol per line. Total: {config.symbols.length} symbols.
        </p>
        <textarea
          value={config.symbols.join("\n")}
          onChange={(e) =>
            updateField(
              "symbols",
              e.target.value
                .split("\n")
                .map((s) => s.trim())
                .filter(Boolean)
            )
          }
          rows={12}
          className="input-field font-mono text-xs"
          placeholder={"INFY\nWIPRO\nTCS\n..."}
        />
      </Section>

      {/* Bottom Save Button */}
      <div className="flex justify-end pb-8">
        <button
          onClick={handleSave}
          disabled={saving}
          className="px-6 py-2.5 bg-sky-600 hover:bg-sky-500 disabled:bg-slate-600 text-white rounded-lg text-sm font-medium transition-colors"
        >
          {saving ? "Saving..." : "Save Configuration"}
        </button>
      </div>
    </div>
  );
}

function Section({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <div className="bg-slate-900 border border-slate-700 rounded-xl p-6">
      <h2 className="text-lg font-semibold text-slate-200 mb-4 pb-2 border-b border-slate-700">
        {title}
      </h2>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">{children}</div>
    </div>
  );
}

function Field({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div className="space-y-1.5">
      <label className="block text-sm font-medium text-slate-400">{label}</label>
      {children}
    </div>
  );
}
