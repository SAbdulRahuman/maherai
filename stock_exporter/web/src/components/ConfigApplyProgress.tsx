"use client";

import type { ConfigApplyStatus } from "@/types/config";

interface Props {
  status: ConfigApplyStatus;
  onDismiss: () => void;
}

export default function ConfigApplyProgress({ status, onDismiss }: Readonly<Props>) {
  const isTerminal = status.state === "ready" || status.state === "error";

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div className="bg-slate-900 border border-slate-700 rounded-xl p-6 max-w-md w-full mx-4 shadow-2xl">
        {/* Header */}
        <div className="flex items-center gap-3 mb-4">
          {status.state === "applying" && (
            <div className="h-5 w-5 border-2 border-sky-400 border-t-transparent rounded-full animate-spin" />
          )}
          {status.state === "ready" && (
            <svg className="h-5 w-5 text-green-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
            </svg>
          )}
          {status.state === "error" && (
            <svg className="h-5 w-5 text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          )}
          <h3 className="text-lg font-semibold text-slate-100">
            {status.state === "applying" && "Applying Configuration..."}
            {status.state === "ready" && "Configuration Applied"}
            {status.state === "error" && "Configuration Failed"}
            {status.state === "idle" && "Ready"}
          </h3>
        </div>

        {/* Current Step */}
        {status.current_step && (
          <p className={`text-sm mb-4 ${
            status.state === "error" ? "text-red-300" : "text-slate-300"
          }`}>
            {status.current_step}
          </p>
        )}

        {/* Completed Steps */}
        {status.completed_steps && status.completed_steps.length > 0 && (
          <div className="space-y-1.5 mb-4">
            {status.completed_steps.map((step) => (
              <div key={step} className="flex items-center gap-2 text-sm">
                <svg className="h-4 w-4 text-green-400 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                </svg>
                <span className="text-slate-400">{step}</span>
              </div>
            ))}
          </div>
        )}

        {/* Error message */}
        {status.error && (
          <div className="bg-red-900/40 border border-red-800 rounded-lg px-3 py-2 mb-4">
            <p className="text-sm text-red-300">{status.error}</p>
          </div>
        )}

        {/* Dismiss button (only when terminal) */}
        {isTerminal && (
          <div className="flex justify-end">
            <button
              onClick={onDismiss}
              className={`px-4 py-1.5 rounded-lg text-sm font-medium transition-colors ${
                status.state === "ready"
                  ? "bg-green-700 hover:bg-green-600 text-white"
                  : "bg-slate-700 hover:bg-slate-600 text-slate-200"
              }`}
            >
              {status.state === "ready" ? "Done" : "Dismiss"}
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
