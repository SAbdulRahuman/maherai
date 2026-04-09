"use client";

import { useEffect, useState, useRef, useCallback } from "react";
import { getConfig, saveConfig, getConfigApplyStatus } from "@/lib/api";
import type { AppConfig, ConfigApplyStatus } from "@/types/config";
import ConfigApplyProgress from "@/components/ConfigApplyProgress";

export default function ConfigPage() {
  const [config, setConfig] = useState<AppConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);
  const [applyStatus, setApplyStatus] = useState<ConfigApplyStatus | null>(null);
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

  useEffect(() => {
    getConfig()
      .then(setConfig)
      .catch((err) =>
        setMessage({ type: "error", text: "Failed to load config: " + err.message })
      )
      .finally(() => setLoading(false));
  }, []);

  // Cleanup polling on unmount
  useEffect(() => {
    return () => {
      if (pollRef.current) clearInterval(pollRef.current);
    };
  }, []);

  const startPolling = useCallback(() => {
    if (pollRef.current) clearInterval(pollRef.current);

    const poll = async () => {
      try {
        const status = await getConfigApplyStatus();
        setApplyStatus(status);

        if (status.state === "ready" || status.state === "error") {
          if (pollRef.current) clearInterval(pollRef.current);
          pollRef.current = null;
          setSaving(false);

          if (status.state === "ready") {
            setMessage({ type: "success", text: "Configuration applied successfully!" });
            // Reload config to reflect any changes
            try {
              const freshConfig = await getConfig();
              setConfig(freshConfig);
            } catch { /* ignore */ }
          } else if (status.state === "error") {
            setMessage({ type: "error", text: status.error || "Configuration failed." });
          }
        }
      } catch {
        // Network error during polling — keep trying
      }
    };

    // Initial poll immediately
    poll();
    pollRef.current = setInterval(poll, 500);

    // Timeout after 30s
    setTimeout(() => {
      if (pollRef.current) {
        clearInterval(pollRef.current);
        pollRef.current = null;
        setSaving(false);
        setApplyStatus((prev) =>
          prev?.state === "applying"
            ? { ...prev, state: "error", error: "Reconfiguration timed out after 30 seconds." }
            : prev
        );
      }
    }, 30000);
  }, []);

  const handleSave = async () => {
    if (!config) return;
    setSaving(true);
    setMessage(null);
    setApplyStatus(null);
    try {
      const res = await saveConfig(config);
      if (res.status === "applying") {
        // Server is doing live reconfiguration — start polling
        setApplyStatus({
          state: "applying",
          current_step: "Initiating configuration update...",
          completed_steps: [],
        });
        startPolling();
      } else {
        // Fallback (no manager) — old behavior
        setSaving(false);
        setMessage({ type: "success", text: res.message || "Configuration saved!" });
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Unknown error";
      setSaving(false);
      setMessage({ type: "error", text: msg });
    }
  };

  const dismissProgress = () => {
    setApplyStatus(null);
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

  const updateRedPanda = <K extends keyof AppConfig["redpanda"]>(
    key: K,
    value: AppConfig["redpanda"][K]
  ) => {
    if (!config) return;
    setConfig({ ...config, redpanda: { ...config.redpanda, [key]: value } });
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
        <h1 className="text-3xl font-bold text-slate-100">Configuration</h1>
        <button
          onClick={handleSave}
          disabled={saving}
          className="px-4 py-2 bg-sky-600 hover:bg-sky-500 disabled:bg-slate-600 text-white rounded-lg text-base font-medium transition-colors"
        >
          {saving ? "Saving..." : "Save Configuration"}
        </button>
      </div>

      {message && (
        <div
          className={`px-4 py-3 rounded-lg text-base ${
            message.type === "success"
              ? "bg-green-900/50 text-green-300 border border-green-800"
              : "bg-red-900/50 text-red-300 border border-red-800"
          }`}
        >
          {message.text}
        </div>
      )}

      {/* Live reconfiguration progress overlay */}
      {applyStatus && applyStatus.state !== "idle" && (
        <ConfigApplyProgress status={applyStatus} onDismiss={dismissProgress} />
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
        <SecretField
          label="API Key"
          value={config.kite.api_key}
          onChange={(v) => updateKite("api_key", v)}
        />
        <SecretField
          label="API Secret"
          value={config.kite.api_secret}
          onChange={(v) => updateKite("api_secret", v)}
        />
        <SecretField
          label="Access Token"
          value={config.kite.access_token}
          onChange={(v) => updateKite("access_token", v)}
        />
        <SecretField
          label="Request Token"
          value={config.kite.request_token}
          onChange={(v) => updateKite("request_token", v)}
        />
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
              updateKite("max_reconnect_attempts", Number.parseInt(e.target.value) || 0)
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
        <SecretField
          label="API Key"
          value={config.api_key}
          onChange={(v) => updateField("api_key", v)}
        />
        <SecretField
          label="API Secret"
          value={config.api_secret}
          onChange={(v) => updateField("api_secret", v)}
        />
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

      {/* RedPanda / Kafka Settings */}
      <Section title="RedPanda / Kafka (Optional)">
        <div className="col-span-full">
          <div className={`inline-flex items-center gap-2 px-3 py-1 rounded-full text-sm font-medium ${
            config.redpanda?.enabled
              ? "bg-green-900/50 text-green-300 border border-green-800"
              : "bg-slate-800 text-slate-400 border border-slate-700"
          }`}>
            <span className={`w-2 h-2 rounded-full ${config.redpanda?.enabled ? "bg-green-400" : "bg-slate-500"}`} />
            {config.redpanda?.enabled ? "Connected" : "Disabled"}
          </div>
        </div>
        <Field label="Brokers">
          <input
            type="text"
            value={config.redpanda?.brokers?.join(", ") ?? ""}
            onChange={(e) =>
              updateRedPanda(
                "brokers",
                e.target.value
                  .split(",")
                  .map((s) => s.trim())
                  .filter(Boolean)
              )
            }
            className="input-field"
            placeholder="localhost:9092, localhost:9093"
          />
          <p className="text-xs text-slate-500 mt-1">Comma-separated broker addresses</p>
        </Field>
        <Field label="Topic">
          <input
            type="text"
            value={config.redpanda?.topic ?? ""}
            onChange={(e) => updateRedPanda("topic", e.target.value)}
            className="input-field"
            placeholder="stock-ticks"
          />
        </Field>
        <Field label="Batch Size">
          <input
            type="number"
            value={config.redpanda?.batch_size ?? 1000}
            onChange={(e) =>
              updateRedPanda("batch_size", Number.parseInt(e.target.value) || 1000)
            }
            className="input-field"
          />
        </Field>
        <Field label="Linger (ms)">
          <input
            type="number"
            value={config.redpanda?.linger_ms ?? 5}
            onChange={(e) =>
              updateRedPanda("linger_ms", Number.parseInt(e.target.value) || 5)
            }
            className="input-field"
          />
        </Field>
        <Field label="Compression">
          <select
            value={config.redpanda?.compression ?? "snappy"}
            onChange={(e) => updateRedPanda("compression", e.target.value)}
            className="input-field"
          >
            <option value="none">None</option>
            <option value="snappy">Snappy</option>
            <option value="lz4">LZ4</option>
            <option value="zstd">Zstd</option>
          </select>
        </Field>
        <Field label="Buffer Size">
          <input
            type="number"
            value={config.redpanda?.buffer_size ?? 131072}
            onChange={(e) =>
              updateRedPanda("buffer_size", Number.parseInt(e.target.value) || 131072)
            }
            className="input-field"
          />
        </Field>
      </Section>

      {/* Symbols Watchlist */}
      <Section title="Symbols Watchlist">
        <p className="text-sm text-slate-400 mb-2">
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
          className="input-field font-mono text-sm"
          placeholder={"INFY\nWIPRO\nTCS\n..."}
        />
      </Section>

      {/* Bottom Save Button */}
      <div className="flex justify-end pb-8">
        <button
          onClick={handleSave}
          disabled={saving}
          className="px-6 py-2.5 bg-sky-600 hover:bg-sky-500 disabled:bg-slate-600 text-white rounded-lg text-base font-medium transition-colors"
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
}: Readonly<{
  title: string;
  children: React.ReactNode;
}>) {
  return (
    <div className="bg-slate-900 border border-slate-700 rounded-xl p-6">
      <h2 className="text-xl font-semibold text-slate-200 mb-4 pb-2 border-b border-slate-700">
        {title}
      </h2>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">{children}</div>
    </div>
  );
}

function Field({
  label,
  children,
}: Readonly<{
  label: string;
  children: React.ReactNode;
}>) {
  return (
    <div className="space-y-1.5">
      <label className="block text-base font-medium text-slate-400">{label}</label>
      {children}
    </div>
  );
}

function maskValue(v: string): string {
  if (v.length <= 4) return "•".repeat(v.length);
  return "•".repeat(v.length - 4) + v.slice(-4);
}

function SecretField({
  label,
  value,
  onChange,
}: Readonly<{
  label: string;
  value: string;
  onChange: (value: string) => void;
}>) {
  const [visible, setVisible] = useState(false);

  return (
    <Field label={label}>
      <div className="relative">
        {visible ? (
          <input
            type="text"
            value={value}
            onChange={(e) => onChange(e.target.value)}
            className="input-field pr-10"
            autoComplete="off"
            data-1p-ignore
            data-lpignore="true"
          />
        ) : (
          <div
            className="input-field pr-10 cursor-text select-none overflow-hidden whitespace-nowrap text-slate-400 tracking-wider"
            onClick={() => setVisible(true)}
          >
            {value ? maskValue(value) : <span className="text-slate-600">••••••••</span>}
          </div>
        )}
        <button
          type="button"
          onClick={() => setVisible((v) => !v)}
          className="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-slate-400 hover:text-slate-200 transition-colors"
          title={visible ? "Hide" : "Show"}
        >
          {visible ? (
            <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M13.875 18.825A10.05 10.05 0 0112 19c-5 0-9.27-3.11-11-7.5a11.72 11.72 0 013.168-4.477M6.343 6.343A9.97 9.97 0 0112 5c5 0 9.27 3.11 11 7.5a11.7 11.7 0 01-4.373 5.157M6.343 6.343L3 3m3.343 3.343l2.829 2.829m4.486 4.486l2.829 2.829M6.343 6.343l11.314 11.314M14.121 14.121A3 3 0 009.879 9.879" />
            </svg>
          ) : (
            <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              <path strokeLinecap="round" strokeLinejoin="round" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
            </svg>
          )}
        </button>
      </div>
    </Field>
  );
}
