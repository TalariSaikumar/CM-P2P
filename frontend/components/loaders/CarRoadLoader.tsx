"use client";

import { useId } from "react";

type Size = "sm" | "md" | "lg";

const widthClass: Record<Size, string> = {
  sm: "w-20",
  md: "w-32",
  lg: "w-44",
};

function svgIds(hook: string) {
  const s = hook.replace(/\W/g, "") || "x";
  return {
    clip: `cmclip${s}`,
    body: `cmbody${s}`,
    glass: `cmglass${s}`,
  };
}

/** Stylized side-view car with moving road and spinning wheels — for full-page and card loaders. */
export function CarRoadLoader({ size = "md", className = "" }: { size?: Size; className?: string }) {
  const { clip, body, glass } = svgIds(useId());

  return (
    <div className={`relative ${widthClass[size]} shrink-0 ${className}`} aria-hidden>
      <svg viewBox="0 0 160 58" className="h-auto w-full overflow-visible drop-shadow-sm">
        <defs>
          <linearGradient id={body} x1="0%" y1="0%" x2="100%" y2="0%">
            <stop offset="0%" stopColor="#475569" />
            <stop offset="45%" stopColor="#64748b" />
            <stop offset="100%" stopColor="#334155" />
          </linearGradient>
          <linearGradient id={glass} x1="0%" y1="0%" x2="0%" y2="100%">
            <stop offset="0%" stopColor="#e2e8f0" stopOpacity="0.95" />
            <stop offset="100%" stopColor="#94a3b8" stopOpacity="0.5" />
          </linearGradient>
          <clipPath id={clip}>
            <rect x="0" y="44" width="160" height="14" rx="1" />
          </clipPath>
        </defs>

        <rect x="0" y="44" width="160" height="14" fill="#0f172a" rx="2" />
        <g clipPath={`url(#${clip})`}>
          <g className="origin-left animate-cm-road will-change-transform">
            {Array.from({ length: 14 }).map((_, i) => (
              <rect key={i} x={i * 28 - 28} y="49" width="16" height="2.5" rx="1" fill="#64748b" opacity="0.85" />
            ))}
          </g>
        </g>

        <g className="animate-cm-float will-change-transform" style={{ transformOrigin: "80px 42px" }}>
          <path
            fill={`url(#${body})`}
            d="M22 38c0-4 3-7 8-7h72c6 0 10 3 12 8l6 14H24l-2-15z"
            stroke="#1e293b"
            strokeWidth="0.5"
          />
          <path fill={`url(#${glass})`} d="M38 24h52c4 0 7 2 8 5l3 9H32l4-10c1-3 4-4 8-4z" opacity="0.9" />
          <path fill="#0f172a" d="M28 38h8l-1-6c-1-3-3-4-6-4h-4l3 10z" opacity="0.35" />
          <rect x="46" y="30" width="38" height="6" rx="1" fill="#1e293b" opacity="0.25" />
          <path fill="#10b981" d="M118 32h12l4 6h-10z" opacity="0.5" />

          <g transform="translate(46, 46)">
            <circle r="7.5" fill="#0f172a" stroke="#334155" strokeWidth="1" />
            <g className="animate-cm-wheel" style={{ transformOrigin: "0px 0px" }}>
              <line x1="0" y1="-5" x2="0" y2="5" stroke="#64748b" strokeWidth="1.2" strokeLinecap="round" />
              <line x1="-5" y1="0" x2="5" y2="0" stroke="#64748b" strokeWidth="1.2" strokeLinecap="round" />
            </g>
          </g>
          <g transform="translate(118, 46)">
            <circle r="7.5" fill="#0f172a" stroke="#334155" strokeWidth="1" />
            <g className="animate-cm-wheel" style={{ transformOrigin: "0px 0px" }}>
              <line x1="0" y1="-5" x2="0" y2="5" stroke="#64748b" strokeWidth="1.2" strokeLinecap="round" />
              <line x1="-5" y1="0" x2="5" y2="0" stroke="#64748b" strokeWidth="1.2" strokeLinecap="round" />
            </g>
          </g>
        </g>
      </svg>
    </div>
  );
}
