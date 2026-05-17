# Deploy CM-P2P frontend on Vercel

The **Next.js app** (`frontend/`) runs on Vercel. The **Go API** (`backend/`) must be hosted separately (Render, Railway, Fly.io, etc.) because Vercel does not run this backend.

## 1. Deploy the frontend on Vercel

### Option A — Root Directory = `frontend` (recommended)

1. Push the repo to GitHub/GitLab/Bitbucket.
2. Go to [vercel.com/new](https://vercel.com/new) and import the repository.
3. **Root Directory:** click Edit → set to **`frontend`**.
4. Framework should auto-detect **Next.js**.
5. Add **Environment Variables** (Production and Preview): only **`APP_ENV`** (`dev`, `stag`, or `prod`). Edit `frontend/config/{APP_ENV}.yaml` in the repo for `api_url`, etc.

6. Click **Deploy**.

Your site is at **`https://cm-p2-p-frontend.vercel.app`**. Use that URL in **Razorpay → Website** (no trailing slash).

### Option B — Deploy from repo root (npm workspaces)

1. Import the repo with **Root Directory** left as **`.`** (repo root).
2. Vercel uses the root `vercel.json` (`npm run build` builds the `frontend` workspace).
3. Set the same environment variables as above.

## 2. Point the frontend at your API

Until the Go API is deployed, the Vercel site cannot log in or pay.

1. Deploy `backend/` to a host with PostgreSQL (e.g. Render web service).
2. Run the API with `APP_ENV=dev` in `backend/.env`; configure `backend/config/dev.yaml` (database, JWT, Razorpay, CORS).
3. On Vercel set only `APP_ENV=dev` in project env; set `api_url` in `frontend/config/dev.yaml`, then redeploy.

Redeploy Vercel after changing `dev.yaml` or env vars.

## 3. CORS on the backend

With `APP_ENV=dev`, `backend/config/dev.yaml` lists localhost and `https://cm-p2-p-frontend.vercel.app` under `cors.allowed_origins`. All `https://*.vercel.app` preview URLs are also allowed by default.

## 4. Razorpay website URL

After the first Vercel deploy, use:

**`https://cm-p2-p-frontend.vercel.app`**

in Razorpay → **Account & Settings** → **Website**. Payment orders also store this URL in Razorpay order notes (`app_url`).

## 5. Optional: custom domain

Vercel → Project → **Domains** → add your domain → update `CORS_ALLOWED_ORIGINS` and Razorpay website URL to match.

## GitHub Actions secrets

For CI builds and Docker deploys, add secrets in GitHub:  
**Settings → Secrets and variables → Actions → New repository secret**

Full list: [`.github/SECRETS.md`](.github/SECRETS.md)

Minimum for Vercel + payments:

- `NEXT_PUBLIC_API_URL` (GitHub + Vercel)
- `RAZORPAY_KEY_ID`, `RAZORPAY_KEY_SECRET` (API host only — never in frontend)

## Quick checklist

- [ ] API deployed and `/api/health` works in the browser  
- [ ] `APP_ENV=dev` on Vercel and `api_url` set in `frontend/config/dev.yaml`  
- [ ] `CORS_ALLOWED_ORIGINS` includes your Vercel URL on the API  
- [ ] Razorpay keys on the API host  
- [ ] Razorpay dashboard website = your Vercel URL  
