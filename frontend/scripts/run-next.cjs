"use strict";
const path = require("path");
const { spawnSync } = require("child_process");

const frontendRoot = path.join(__dirname, "..");
const args = process.argv.slice(2);
if (args.length === 0) {
  console.error("usage: node scripts/run-next.cjs <next-args…>");
  process.exit(1);
}

let nextBin;
try {
  const nextPkg = require.resolve("next/package.json", { paths: [frontendRoot] });
  nextBin = path.join(path.dirname(nextPkg), "dist", "bin", "next");
} catch {
  console.error("Could not resolve 'next'. From repo root run: npm install");
  process.exit(1);
}

const r = spawnSync(process.execPath, [nextBin, ...args], {
  stdio: "inherit",
  cwd: frontendRoot,
  env: process.env,
  shell: false,
});
process.exit(r.status === null ? 1 : r.status);
