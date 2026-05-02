"use client";

/** Shimmer placeholders while car search results load. */
export function SearchResultsSkeleton({ rows = 4 }: { rows?: number }) {
  return (
    <ul className="space-y-3" aria-hidden>
      {Array.from({ length: rows }).map((_, i) => (
        <li
          key={i}
          className="relative flex flex-col gap-3 overflow-hidden rounded-lg border border-slate-200 bg-white p-4 shadow-sm sm:flex-row sm:items-center sm:justify-between"
        >
          <div className="pointer-events-none absolute inset-0 animate-cm-shimmer bg-gradient-to-r from-transparent via-white/75 to-transparent" />
          <div className="min-w-0 flex-1 space-y-2">
            <div className="h-4 w-[min(280px,78%)] rounded bg-slate-200/90" />
            <div className="h-3 w-[min(360px,92%)] rounded bg-slate-100" />
            <div className="h-3 w-[min(200px,48%)] rounded bg-slate-100" />
          </div>
          <div className="h-10 w-full shrink-0 rounded-md bg-slate-100 sm:w-24" />
        </li>
      ))}
    </ul>
  );
}
