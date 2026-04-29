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

You can run commands from the **repo root** (`CM-P2P/`) or from **`frontend/`** — the root `package.json` uses an npm **workspace** so `npm install` and `npm run dev` work from either place.

1. Open a second terminal (repo root or `frontend/`).
2. **`frontend/.env`** is committed with **`APP_ENV=dev`**. Edit **`APP_ENV`** there for `stag` / `prod`, or add **`frontend/.env.local`** for overrides. The app loads **`frontend/config/{APP_ENV}.yaml`** for `api_url` and `support_email` (wired to `NEXT_PUBLIC_*` in `next.config.mjs`).
3. Install dependencies (from repo root **or** `frontend/`):
   - `npm install`
4. Start dev server:
   - `npm run dev`

**Windows (PowerShell):** If `npm` fails with `Cannot find module ... npm-prefix.js` / `Could not determine Node.js install directory`, your `PATH` is picking the wrong npm shim. Either fix `PATH` (remove `C:\Program Files\nodejs\node_modules\npm\bin`, keep `C:\Program Files\nodejs\` first—see earlier notes), or use the bundled scripts from `frontend`:

- `.\scripts\install.ps1`
- `.\scripts\dev.ps1`

Or call npm directly: `& "$env:ProgramFiles\nodejs\npm.cmd" install` and `& "$env:ProgramFiles\nodejs\npm.cmd" run dev`.

If PowerShell blocks scripts, use **cmd**: `frontend\scripts\install.cmd` then `frontend\scripts\dev.cmd`.

**Git Bash (Windows)** — the default `dev` script runs **`node ./scripts/run-next.cjs dev`**, which resolves `next` from the hoisted workspace `node_modules` and avoids **`next.cmd`** (that path often yields **`'node' is not recognized`** under `cmd.exe`). If it still fails, use either:

- From `frontend`: `npm run dev:gitbash` (runs `scripts/dev.sh` with a direct path to `node.exe`), or  
- `bash ./scripts/dev.sh`

Long-term fix: add **`C:\Program Files\nodejs`** to your **Windows** user or system **PATH** (Settings → System → About → Advanced system settings → Environment Variables), then reopen the terminal. Optionally in `~/.bashrc`: `export PATH="/c/Program Files/nodejs:$PATH"` so Bash and child tools agree.

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

From repo root: `npm install && npm run dev` — or from `frontend/`: `npm install && npm run dev`.

On Windows if `npm` is broken in PowerShell, use `frontend\scripts\install.cmd` and `frontend\scripts\dev.cmd`, or `npm run dev:gitbash` from root / `frontend` (Git Bash).
