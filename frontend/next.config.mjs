import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

/**
 * Flat `key: value` YAML (one level, no lists/maps). Full-line # comments;
 * end-of-line comments must be preceded by two spaces then #.
 */
function parseFlatYaml(text) {
  /** @type {Record<string, string>} */
  const out = {};
  for (const line of text.split("\n")) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith("#")) continue;
    const m = trimmed.match(/^([A-Za-z_][A-Za-z0-9_]*)\s*:\s*(.*)$/);
    if (!m) continue;
    let val = m[2].trim();
    const idx = val.search(/\s+#/);
    if (idx !== -1) val = val.slice(0, idx).trim();
    if (
      (val.startsWith('"') && val.endsWith('"')) ||
      (val.startsWith("'") && val.endsWith("'"))
    ) {
      val = val.slice(1, -1);
    }
    if (val !== "") out[m[1]] = val;
  }
  return out;
}

const APP_ENV_MODES = ["dev", "stag", "prod"];

/** Map YAML keys under config/{APP_ENV}.yaml → process.env keys */
const YAML_TO_ENV = {
  api_url: "NEXT_PUBLIC_API_URL",
  support_email: "NEXT_PUBLIC_SUPPORT_EMAIL",
};

function resolveAppEnv() {
  const raw = (process.env.APP_ENV || "dev").trim().toLowerCase();
  if (APP_ENV_MODES.includes(raw)) return raw;
  if (raw) {
    console.warn(
      `[next.config] Invalid APP_ENV="${process.env.APP_ENV}". Expected one of: ${APP_ENV_MODES.join(", ")}. Using "dev".`,
    );
  }
  return "dev";
}

function loadYamlAppConfig(appEnv) {
  const file = path.join(__dirname, "config", `${appEnv}.yaml`);
  if (!fs.existsSync(file)) {
    console.warn(`[next.config] Missing ${file}, skipping YAML config.`);
    return null;
  }
  try {
    const doc = parseFlatYaml(fs.readFileSync(file, "utf8"));
    return Object.keys(doc).length ? doc : null;
  } catch (e) {
    console.warn(`[next.config] Failed to parse ${file}:`, e);
    return null;
  }
}

/**
 * Apply YAML values only when the target env var is still unset (so .env /
 * .env.local overrides win).
 */
function applyYamlToProcessEnv(doc) {
  if (!doc) return;
  for (const [yamlKey, envKey] of Object.entries(YAML_TO_ENV)) {
    if (process.env[envKey] !== undefined) continue;
    const raw = doc[yamlKey];
    if (raw == null) continue;
    const val = String(raw).trim();
    if (val === "") continue;
    process.env[envKey] = val;
  }
}

const appEnv = resolveAppEnv();
process.env.APP_ENV = appEnv;
if (process.env.NEXT_PUBLIC_APP_ENV === undefined) {
  process.env.NEXT_PUBLIC_APP_ENV = appEnv;
}
applyYamlToProcessEnv(loadYamlAppConfig(appEnv));

/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  images: {
    remotePatterns: [
      {
        protocol: "https",
        hostname: "*.blob.core.windows.net",
        pathname: "/**",
      },
    ],
  },
};

export default nextConfig;
