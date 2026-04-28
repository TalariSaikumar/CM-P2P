# Running Backend and Frontend

This project has two apps:

- `backend`: Go (Gin + GORM) API
- `frontend`: Next.js web app

Run each app in a separate terminal.

## Prerequisites

- Go `1.23+`
- Node.js `18+` and npm
- PostgreSQL database (URL is configured in `backend/config/<APP_ENV>.yaml`)

## 1) Run the backend API

1. Open a terminal in `backend`.
2. Copy `backend/.env.example` to `backend/.env` and set **`APP_ENV`** to one of **`dev`**, **`stag`**, or **`prod`**. Only this file is read from dotenv; everything else comes from YAML.
3. Edit the matching config file:
   - `APP_ENV=dev` → `backend/config/dev.yaml`
   - `APP_ENV=stag` → `backend/config/stag.yaml`
   - `APP_ENV=prod` → `backend/config/prod.yaml`
   
   At minimum set **`database_url`** and **`jwt_secret`** (see keys in `dev.yaml`).
4. Install dependencies (optional):
   - `go mod download`
5. Start the server from `backend`:
   - `go run .` or `go run main.go`

Expected result:

- API listens on `http://localhost:8080`
- Routes are under `http://localhost:8080/api`

### Connect PostgreSQL with pgAdmin (localhost)

Use the same host, port, database name, and user as in `database_url` inside your active YAML (for local dev, usually `backend/config/dev.yaml`). If your local PostgreSQL user has **no password** (common with `trust` / `peer` auth), leave **Password** empty in pgAdmin and use a URL **without** `:PASSWORD` — for example `postgres://postgres@localhost:5432/carmanage?sslmode=disable`.

Example local setup (adjust names/passwords to match what you create in PostgreSQL):

1. In pgAdmin: **Object → Register → Server**.
2. **General** tab: **Name** — e.g. `CM-P2P local`.
3. **Connection** tab:

| Field | Example (local) |
| --- | --- |
| Host name/address | `localhost` |
| Port | `5432` |
| Maintenance database | `postgres` (or your DB name if it already exists) |
| Username | `carmanage` |
| Password | Leave blank if your local user has no password; otherwise enter it |
| Save password? | Optional |

4. Create an application database if needed: **Databases → Create → Database** — e.g. `carmanage`.
5. Set `database_url` in `backend/config/dev.yaml` (or the YAML for your `APP_ENV`) to match, for example:

- With password: `postgres://carmanage:YOUR_PASSWORD@localhost:5432/carmanage?sslmode=disable`
- **No password:** `postgres://carmanage@localhost:5432/carmanage?sslmode=disable` (no `:` before `@`)

For local PostgreSQL without TLS, `sslmode=disable` is typical. Remote/Azure usually uses `sslmode=require`.

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
- Set **`DATABASE_URL`** in the shell to the same value as **`database_url`** in your YAML (the migrate scripts do not read YAML). Example: copy `database_url` from `config/dev.yaml`, then in PowerShell `$env:DATABASE_URL = "postgres://..."` before running the script.

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

`cd backend && go run .`

Terminal 2 (frontend):

`cd frontend && npm install && npm run dev`
