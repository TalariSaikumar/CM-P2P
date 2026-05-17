/**
 * Sync package-lock.json with frontend/package.json (Next 14.2.35).
 * Run: node scripts/update-next-lock.mjs
 */
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.join(path.dirname(fileURLToPath(import.meta.url)), "..");
const lockPath = path.join(root, "package-lock.json");

const UPDATES = {
  next: "14.2.35",
  "eslint-config-next": "14.2.35",
  "@next/env": "14.2.35",
  "@next/eslint-plugin-next": "14.2.35",
  "@next/swc-darwin-arm64": "14.2.33",
  "@next/swc-darwin-x64": "14.2.33",
  "@next/swc-linux-arm64-gnu": "14.2.33",
  "@next/swc-linux-arm64-musl": "14.2.33",
  "@next/swc-linux-x64-gnu": "14.2.33",
  "@next/swc-linux-x64-musl": "14.2.33",
  "@next/swc-win32-arm64-msvc": "14.2.33",
  "@next/swc-win32-ia32-msvc": "14.2.33",
  "@next/swc-win32-x64-msvc": "14.2.33",
};

async function fetchMeta(name, version) {
  const url = `https://registry.npmjs.org/${name.replace("/", "%2F")}/${version}`;
  const res = await fetch(url);
  if (!res.ok) throw new Error(`Failed ${url}: ${res.status}`);
  const data = await res.json();
  const tarball = data.dist.tarball;
  const integrity = data.dist.integrity;
  if (!tarball || !integrity) throw new Error(`Missing dist for ${name}@${version}`);
  return { resolved: tarball, integrity, version };
}

function pkgKey(name) {
  return `node_modules/${name}`;
}

async function main() {
  const lock = JSON.parse(fs.readFileSync(lockPath, "utf8"));

  lock.packages["frontend"].dependencies.next = UPDATES.next;
  lock.packages["frontend"].devDependencies["eslint-config-next"] = UPDATES["eslint-config-next"];

  for (const [name, version] of Object.entries(UPDATES)) {
    const key = pkgKey(name);
    if (!lock.packages[key]) {
      console.warn("skip missing", key);
      continue;
    }
    const meta = await fetchMeta(name, version);
    lock.packages[key].version = meta.version;
    lock.packages[key].resolved = meta.resolved;
    lock.packages[key].integrity = meta.integrity;
    if (lock.packages[key].deprecated) delete lock.packages[key].deprecated;
    console.log("updated", name, version);
  }

  const next = lock.packages["node_modules/next"];
  next.dependencies["@next/env"] = UPDATES["@next/env"];
  for (const swc of Object.keys(UPDATES)) {
    if (swc.startsWith("@next/swc-") && next.optionalDependencies?.[swc]) {
      next.optionalDependencies[swc] = UPDATES[swc];
    }
  }

  const ecn = lock.packages["node_modules/eslint-config-next"];
  if (ecn?.dependencies) {
    ecn.dependencies["@next/eslint-plugin-next"] = UPDATES["@next/eslint-plugin-next"];
  }

  fs.writeFileSync(lockPath, JSON.stringify(lock, null, 2) + "\n");
  console.log("Wrote", lockPath);
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
