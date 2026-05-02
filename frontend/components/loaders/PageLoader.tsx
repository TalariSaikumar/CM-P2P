"use client";

import { CarRoadLoader } from "./CarRoadLoader";

type Props = {
  title?: string;
  subtitle?: string;
  /** Default: comfortable vertical centering for main content areas */
  className?: string;
  /** Larger card + loader for hero-style empty states */
  variant?: "default" | "hero";
};

/** Centered car-themed loader for initial data fetch and full-page waits. */
export function PageLoader({ title = "Loading…", subtitle, className = "", variant = "default" }: Props) {
  const pad = variant === "hero" ? "px-12 py-14 sm:px-16 sm:py-16" : "px-10 py-11 sm:px-12 sm:py-12";
  const carSize = variant === "hero" ? "lg" : "md";

  return (
    <div
      className={`flex min-h-[min(420px,70dvh)] flex-col items-center justify-center gap-2 py-12 ${className}`}
      role="status"
      aria-live="polite"
      aria-busy="true"
    >
      <div
        className={`rounded-2xl border border-slate-200/90 bg-gradient-to-b from-white to-slate-50/90 shadow-xl shadow-slate-900/[0.06] ring-1 ring-slate-100/80 ${pad}`}
      >
        <CarRoadLoader size={carSize} className="mx-auto" />
        <p className="mt-8 text-center text-sm font-semibold tracking-wide text-slate-800">{title}</p>
        {subtitle ? (
          <p className="mt-1 max-w-xs text-center text-xs leading-relaxed text-slate-500">{subtitle}</p>
        ) : (
          <p className="mt-1 text-center text-xs text-slate-400">CarManage</p>
        )}
      </div>
    </div>
  );
}
