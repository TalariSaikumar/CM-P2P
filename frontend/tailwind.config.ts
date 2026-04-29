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
    },
  },
  plugins: [],
};

export default config;
