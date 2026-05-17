# Deploy CM-P2P frontend on Vercel

The **Next.js app** (`frontend/`) runs on Vercel. The **Go API** (`backend/`) must be hosted separately (Render, Railway, Fly.io, etc.) because Vercel does not run this backend.

## 1. Deploy the frontend on Vercel

### Option A — Root Directory = `frontend` (recommended)

1. Push the repo to GitHub/GitLab/Bitbucket.
2. Go to [vercel.com/new](https://vercel.com/new) and import the repository.
3. **Root Directory:** click Edit → set to **`frontend`**.
4. Framework should auto-detect **Next.js**.
5. Add **Environment Variables** (Production and Preview):

   | Name | Example |
   |------|---------|
   | `APP_ENV` | `stag` or `prod` |
   | `NEXT_PUBLIC_API_URL` | `https://your-api.onrender.com/api` |

6. Click **Deploy**.

Your site will be at `https://<project>.vercel.app`. Use that URL in **Razorpay → Website**.

### Option B — Deploy from repo root (npm workspaces)

1. Import the repo with **Root Directory** left as **`.`** (repo root).
2. Vercel uses the root `vercel.json` (`npm run build` builds the `frontend` workspace).
3. Set the same environment variables as above.

## 2. Point the frontend at your API

Until the Go API is deployed, the Vercel site cannot log in or pay.

1. Deploy `backend/` to a host with PostgreSQL (e.g. Render web service).
2. Set on the API host:
   - `DATABASE_URL`, `JWT_SECRET`, `RAZORPAY_KEY_ID`, `RAZORPAY_KEY_SECRET`
   - `CORS_ALLOWED_ORIGINS=https://your-project.vercel.app` (comma-separated for multiple)
3. Set on Vercel:
   - `NEXT_PUBLIC_API_URL=https://<your-api-host>/api`

Redeploy Vercel after changing env vars.

## 3. CORS on the backend

Local dev allows `http://localhost:3000`. For Vercel, add in **`backend/.env`** (local) or your API host env:

```env
CORS_ALLOWED_ORIGINS=https://your-project.vercel.app,https://your-project-git-main-you.vercel.app
```

By default, `https://*.vercel.app` preview URLs are also allowed unless you set:

```env
CORS_ALLOW_VERCEL=false
```

## 4. Razorpay website URL

After the first Vercel deploy, use:

`https://<your-project>.vercel.app`

in Razorpay onboarding (**Website**). You can change it later in the Razorpay dashboard.

## 5. Optional: custom domain

Vercel → Project → **Domains** → add your domain → update `CORS_ALLOWED_ORIGINS` and Razorpay website URL to match.

## Quick checklist

- [ ] API deployed and `/api/health` works in the browser  
- [ ] `NEXT_PUBLIC_API_URL` set on Vercel  
- [ ] `CORS_ALLOWED_ORIGINS` includes your Vercel URL on the API  
- [ ] Razorpay keys on the API host  
- [ ] Razorpay dashboard website = your Vercel URL  
