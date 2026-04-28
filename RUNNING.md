# Running Backend and Frontend

This project has two apps:

- `backend`: Go (Gin + GORM) API
- `frontend`: Next.js web app

Run each app in a separate terminal.

## Prerequisites

- Go `1.23+`
- Node.js `18+` and npm
- PostgreSQL database accessible via `DATABASE_URL`

## 1) Run the backend API

1. Open a terminal in `backend`.
2. Copy env template:
   - PowerShell: `Copy-Item .env.example .env`
   - Bash: `cp .env.example .env`
3. Update `backend/.env` (minimum required):
   - `DATABASE_URL`
   - `JWT_SECRET`
4. Install dependencies (optional, `go run` can also fetch automatically):
   - `go mod download`
5. Start server:
   - `go run ./cmd/server`

Expected result:

- API listens on `http://localhost:8080`
- Routes are under `http://localhost:8080/api`

## 1.1) Apply SQL migration manually (optional)

If you want explicit SQL-based schema setup, use:

- `backend/migrations/*.up.sql`

Run with `psql` from `backend`:

- PowerShell:
  - `Get-ChildItem .\migrations\*.up.sql | Sort-Object Name | ForEach-Object { psql "$env:DATABASE_URL" -f $_.FullName }`
- Bash:
  - `for f in migrations/*.up.sql; do psql "$DATABASE_URL" -f "$f"; done`

Each migration is idempotent (`CREATE TABLE IF NOT EXISTS` and `ADD COLUMN IF NOT EXISTS`) so it can be used for both fresh and existing databases.

The first file `000000_extensions_and_timestamps.up.sql` enables `pgcrypto` (for `gen_random_uuid()` on older PostgreSQL) and defines `touch_row_timestamps_utc()`. Every domain table uses `id uuid PRIMARY KEY DEFAULT gen_random_uuid()` and `BEFORE INSERT OR UPDATE` triggers so `created_at` and `updated_at` are set from the database clock on insert and on each update (`timestamptz` is stored internally as UTC).

## 1.2) Apply migrations via scripts (recommended)

From `backend`:

- PowerShell `up`: `.\scripts\migrate.ps1 -Direction up`
- PowerShell `down`: `.\scripts\migrate.ps1 -Direction down`
- Bash `up`: `./scripts/migrate.sh up`
- Bash `down`: `./scripts/migrate.sh down`

Requirements:

- `psql` must be installed and available in `PATH`
- `DATABASE_URL` must be set (or pass `-DatabaseUrl` to `migrate.ps1`)

## 2) Run the frontend web app

1. Open a second terminal in `frontend`.
2. Copy env template:
   - PowerShell: `Copy-Item .env.example .env.local`
   - Bash: `cp .env.example .env.local`
3. Ensure `frontend/.env.local` has:
   - `NEXT_PUBLIC_API_URL=http://localhost:8080/api`
4. Install dependencies:
   - `npm install`
5. Start dev server:
   - `npm run dev`

Expected result:

- Web app runs at `http://localhost:3000`

## Optional services

In `backend/.env`, these are optional for local development:

- Azure Blob (`AZURE_STORAGE_ACCOUNT`, `AZURE_STORAGE_KEY`, `AZURE_STORAGE_CONTAINER`)
- Twilio (`TWILIO_ACCOUNT_SID`, `TWILIO_AUTH_TOKEN`, `TWILIO_FROM_NUMBER`)
- Demo KYC shortcut: `ALLOW_SELF_KYC_VERIFY=true`

## Quick startup summary

Terminal 1 (backend):

`cd backend && go run ./cmd/server`

Terminal 2 (frontend):

`cd frontend && npm install && npm run dev`
