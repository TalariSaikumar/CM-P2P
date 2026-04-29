#!/usr/bin/env bash
# Git Bash on Windows: npm's "next dev" runs through cmd.exe, which often lacks Node on PATH.
# This script calls node.exe directly (typical install paths + PATH fallback).
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
NODE=""
for c in \
  "/c/Program Files/nodejs/node.exe" \
  "/c/Program Files (x86)/nodejs/node.exe"; do
  if [ -f "$c" ]; then
    NODE="$c"
    break
  fi
done
if [ -z "$NODE" ]; then
  NODE="$(command -v node 2>/dev/null || true)"
fi
if [ -z "$NODE" ]; then
  echo "Could not find node.exe. Install Node.js and/or add it to your Windows PATH." >&2
  echo "See RUNNING.md (Git Bash / Windows)." >&2
  exit 1
fi
cd "$ROOT"
exec "$NODE" "./scripts/run-next.cjs" dev "$@"
