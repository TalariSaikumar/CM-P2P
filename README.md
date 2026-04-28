# CarManage

P2P car rental monorepo: **Go (Gin) + GORM** API under `backend/`, **Next.js 14 + Tailwind** UI under `frontend/`.

Detailed run guide: see `RUNNING.md`.

## Run the API

1. Copy `backend/.env.example` to `backend/.env` and set `APP_ENV` to `dev`, `stag`, or `prod`. Edit the matching file under `backend/config/` (for example `config/dev.yaml`) for database URL, JWT secret, Azure, and Twilio.
2. From `backend/`: `go run .` (or `go run main.go`) — listens on `:8080` by default, routes under `/api`.

## Run the web app

1. Copy `frontend/.env.example` to `frontend/.env.local` and set `NEXT_PUBLIC_API_URL` (default `http://localhost:8080/api`).
2. From `frontend/`: `npm install` then `npm run dev` — opens on `http://localhost:3000` (CORS is allowed for this origin on the API).

## Main flows

- Register as **CUSTOMER** or **OWNER**; sign in receives a JWT.
- **KYC**: with `allow_self_kyc_verify: true` in your YAML (e.g. `config/dev.yaml`), use **Account → Mark KYC verified (demo)** after filling profile. Customers also need a **driving license** on the account before creating bookings.
- **Owner**: My fleet → add car → upload images (requires Azure keys in YAML).
- **Customer**: Search cars → **Booking inquiry** → shared **booking chat**; owner sets **final price** (`PATCH /api/bookings/:id/price`); customer **confirms**; Twilio SMS fires when Twilio fields are set in YAML.
