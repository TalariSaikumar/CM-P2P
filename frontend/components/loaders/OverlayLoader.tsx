"use client";

import { CarRoadLoader } from "./CarRoadLoader";

/** Dim overlay with loader for blocking actions (e.g. payment submit). */
export function OverlayLoader({ message = "Processing…" }: { message?: string }) {
  return (
    <div
      className="fixed inset-0 z-[100] flex items-center justify-center bg-slate-900/45 p-6 backdrop-blur-[2px]"
      role="status"
      aria-modal="true"
      aria-busy="true"
      aria-live="polite"
    >
      <div className="max-w-sm rounded-2xl border border-white/20 bg-white/95 px-8 py-10 text-center shadow-2xl shadow-slate-900/20">
        <CarRoadLoader size="md" className="mx-auto" />
        <p className="mt-6 text-sm font-semibold text-slate-800">{message}</p>
        <p className="mt-1 text-xs text-slate-500">Please do not close this window.</p>
      </div>
    </div>
  );
}
