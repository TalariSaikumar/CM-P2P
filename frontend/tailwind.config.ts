import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ["var(--font-dm-sans)", "system-ui", "sans-serif"],
      },
      backgroundImage: {
        "grid-slate":
          "linear-gradient(to right, rgb(148 163 184 / 0.07) 1px, transparent 1px), linear-gradient(to bottom, rgb(148 163 184 / 0.07) 1px, transparent 1px)",
      },
      keyframes: {
        cmRoad: {
          "0%": { transform: "translateX(0)" },
          "100%": { transform: "translateX(-36px)" },
        },
        cmCarFloat: {
          "0%, 100%": { transform: "translateY(0)" },
          "50%": { transform: "translateY(-3px)" },
        },
        cmWheel: {
          to: { transform: "rotate(360deg)" },
        },
        cmShimmer: {
          "0%": { transform: "translateX(-100%)" },
          "100%": { transform: "translateX(200%)" },
        },
      },
      animation: {
        "cm-road": "cmRoad 0.5s linear infinite",
        "cm-float": "cmCarFloat 2.4s ease-in-out infinite",
        "cm-wheel": "cmWheel 0.68s linear infinite",
        "cm-shimmer": "cmShimmer 1.35s ease-in-out infinite",
      },
    },
  },
  plugins: [],
};

export default config;
