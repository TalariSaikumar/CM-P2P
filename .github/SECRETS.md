# GitHub repository secrets

Add these under **Settings → Secrets and variables → Actions → Repository secrets**  
([github.com/TalariSaikumar/CM-P2P/settings/secrets/actions](https://github.com/TalariSaikumar/CM-P2P/settings/secrets/actions)).

`.env` files only set `APP_ENV`; put secrets in `backend/config/{APP_ENV}.yaml` (do not commit real production secrets). `backend/.env` is gitignored.

---

## Frontend CI / Vercel (`frontend-ci.yml`)

| Secret | Required | Description |
|--------|----------|-------------|
| `NEXT_PUBLIC_API_URL` | Recommended | Public API URL for production builds, e.g. `https://your-api.example.com/api` |

Also set the same in **Vercel → Project → Settings → Environment Variables** for deployed sites.

| Variable | Example |
|----------|---------|
| `APP_ENV` | `stag` or `prod` |
| `NEXT_PUBLIC_API_URL` | `https://your-api.example.com/api` |

---

## Backend API (runtime — Azure Container Apps, Render, Railway, etc.)

Set on your **hosting platform**, not in the Docker image. Optional: mirror in GitHub secrets for future deploy workflows.

| Secret | Required | Description |
|--------|----------|-------------|
| `DATABASE_URL` | Yes | Postgres connection string |
| `JWT_SECRET` | Yes | Long random string for JWT signing |
| `RAZORPAY_KEY_ID` | For payments | `rzp_test_…` or `rzp_live_…` |
| `RAZORPAY_KEY_SECRET` | For payments | From Razorpay dashboard |
| `AZURE_STORAGE_ACCOUNT` | For uploads | Blob storage account name |
| `AZURE_STORAGE_KEY` | For uploads | Blob storage key |
| `AZURE_STORAGE_CONTAINER` | For uploads | Container name |
| `CORS_ALLOWED_ORIGINS` | For Vercel | `https://cm-p2-p-frontend.vercel.app` |
| `PUBLIC_APP_URL` | Razorpay notes / CORS | `https://cm-p2-p-frontend.vercel.app` |
| `TWILIO_ACCOUNT_SID` | Optional | SMS |
| `TWILIO_AUTH_TOKEN` | Optional | SMS |
| `TWILIO_FROM_NUMBER` | Optional | SMS |

---

## Docker → Azure Container Registry (`.github/workflows/docker-acr.yml`)

| Secret | Description |
|--------|-------------|
| `ACR_LOGIN_SERVER` | e.g. `myregistry.azurecr.io` |
| `ACR_USERNAME` | Registry username |
| `ACR_PASSWORD` | Registry password |
| `ACR_REPOSITORY` | Image repo name |
| `ACR_IMAGE_TAG` | Tag, e.g. `stag` |

**Repository variable** (optional): `DOCKER_APP_ENV` = `dev` | `stag` | `prod`

Razorpay and database secrets are **not** baked into the image. Configure them as **environment variables on the container** where the API runs.

---

## Quick setup checklist

1. GitHub → **Settings → Secrets and variables → Actions** → **New repository secret**
2. Add `NEXT_PUBLIC_API_URL`, `RAZORPAY_KEY_ID`, `RAZORPAY_KEY_SECRET`, `JWT_SECRET`, `DATABASE_URL` as needed
3. Vercel → same public vars (`NEXT_PUBLIC_*`, `APP_ENV`)
4. API host → all backend secrets + `CORS_ALLOWED_ORIGINS`
