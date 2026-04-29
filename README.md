# CarManage

P2P car rental monorepo: **Go (Gin) + GORM** API under `backend/`, **Next.js 14 + Tailwind** UI under `frontend/`.

Detailed run guide: see `RUNNING.md`.

## Run the API

1. Copy `backend/.env.example` to `backend/.env` and set `APP_ENV` to `dev`, `stag`, or `prod`. Edit the matching file under `backend/config/` (for example `config/dev.yaml`) for database URL, JWT secret, Azure, and Twilio.
2. From `backend/`: `go run .` (or `go run main.go`) — listens on `:8080` by default, routes under `/api`.

## Run the web app

1. The repo includes **`frontend/.env`** with **`APP_ENV=dev`**. Change **`APP_ENV`** there to `stag` or `prod` when needed, or add **`frontend/.env.local`** for machine-only overrides. API URL and support email default from **`frontend/config/{APP_ENV}.yaml`** (see `next.config.mjs`); you can still set **`NEXT_PUBLIC_*`** in `.env.local`.
2. From the **repo root** or **`frontend/`**: `npm install` then `npm run dev` — opens on `http://localhost:3000` (CORS is allowed for this origin on the API). The root `package.json` wires scripts to the `frontend` workspace.

## Main flows

- Register as **CUSTOMER** or **OWNER**; sign in receives a JWT.
- **KYC**: with `allow_self_kyc_verify: true` in your YAML (e.g. `config/dev.yaml`), use **Account → Mark KYC verified (demo)** after filling profile. Customers also need a **driving license** on the account before creating bookings.
- **Owner**: My fleet → add car → upload images (requires Azure keys in YAML).
- **Customer**: Search cars → **Book** (rental dates, pickup & drop-off required) → optional **chat** to negotiate; owner sets **final price** (`PATCH /api/bookings/:id/price`) and **confirms**; customer may **withdraw** until a final price is set; Twilio SMS may notify the customer when Twilio is configured in YAML.
