"use client";

/** Compact steering-wheel style spinner for primary buttons (car-themed, no layout shift). */
export function ButtonCarSpinner({ className = "" }: { className?: string }) {
  return (
    <svg
      className={`inline-block h-[1.1em] w-[1.1em] shrink-0 animate-spin text-emerald-200 ${className}`}
      viewBox="0 0 24 24"
      fill="none"
      aria-hidden
    >
      <circle cx="12" cy="12" r="9" stroke="currentColor" strokeWidth="1.5" opacity="0.35" />
      <path
        d="M12 3v3M12 18v3M3 12h3M18 12h3M5.6 5.6l2.1 2.1M16.3 16.3l2.1 2.1M5.6 18.4l2.1-2.1M16.3 7.7l2.1-2.1"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
      />
      <circle cx="12" cy="12" r="2.2" fill="currentColor" />
    </svg>
  );
}
