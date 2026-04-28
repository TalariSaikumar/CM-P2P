# CarManage

P2P car rental monorepo: **Go (Gin) + GORM** API under `backend/`, **Next.js 14 + Tailwind** UI under `frontend/`.

Detailed run guide: see `RUNNING.md`.

## Run the API

1. Copy `backend/.env.example` to `backend/.env` and set `DATABASE_URL`, `JWT_SECRET`, and optionally Azure Blob + Twilio + `ALLOW_SELF_KYC_VERIFY=true` for local KYC demos.
2. From `backend/`: `go run ./cmd/server` — listens on `:8080`, routes under `/api`.

## Run the web app

1. Copy `frontend/.env.example` to `frontend/.env.local` and set `NEXT_PUBLIC_API_URL` (default `http://localhost:8080/api`).
2. From `frontend/`: `npm install` then `npm run dev` — opens on `http://localhost:3000` (CORS is allowed for this origin on the API).

## Main flows

- Register as **CUSTOMER** or **OWNER**; sign in receives a JWT.
- **KYC**: with `ALLOW_SELF_KYC_VERIFY=true`, use **Account → Mark KYC verified (demo)** after filling profile. Customers also need a **driving license** on the account before creating bookings.
- **Owner**: My fleet → add car → upload images (requires Azure blob env vars).
- **Customer**: Search cars → **Booking inquiry** → shared **booking chat**; owner sets **final price** (`PATCH /api/bookings/:id/price`); customer **confirms**; Twilio SMS fires when Twilio env is set.
