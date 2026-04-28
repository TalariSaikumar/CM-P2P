#!/usr/bin/env bash
set -euo pipefail

direction="${1:-up}"
database_url="${DATABASE_URL:-}"

if [[ "$direction" != "up" && "$direction" != "down" ]]; then
  echo "Usage: ./scripts/migrate.sh [up|down]" >&2
  exit 1
fi

if [[ -z "$database_url" ]]; then
  echo "DATABASE_URL is not set." >&2
  exit 1
fi

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
migrations_dir="$root_dir/migrations"

if [[ ! -d "$migrations_dir" ]]; then
  echo "Migrations directory not found: $migrations_dir" >&2
  exit 1
fi

if [[ "$direction" == "up" ]]; then
  mapfile -t files < <(ls "$migrations_dir"/*.up.sql 2>/dev/null | sort)
else
  mapfile -t files < <(ls "$migrations_dir"/*.down.sql 2>/dev/null | sort -r)
fi

if [[ "${#files[@]}" -eq 0 ]]; then
  echo "No migration files found for direction: $direction"
  exit 0
fi

for file in "${files[@]}"; do
  echo "Applying $direction migration: $(basename "$file")"
  psql "$database_url" -v ON_ERROR_STOP=1 -f "$file"
done

echo "All '$direction' migrations applied successfully."
