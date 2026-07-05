# Deploying Quota

Architecture: **frontend on Vercel**, **backend on Render** (Docker), **Postgres
on Render** (managed). Deploy the backend first so you have its URL for the
frontend.

The repo already contains everything needed:
- `backend/Dockerfile` — builds the Go server
- `render.yaml` — Render blueprint (web service + Postgres)
- `frontend/vercel.json` — SPA routing
- API base URL is controlled by `VITE_API_BASE_URL` (empty in dev)

---

## 1. Backend + database on Render

1. Push this repo to GitHub (done: https://github.com/Mayorken/quota).
2. In the [Render dashboard](https://dashboard.render.com) → **New → Blueprint**,
   select this repo. Render reads `render.yaml` and proposes a **web service**
   (`quota-backend`) plus a **Postgres** database (`quota-db`).
3. Click **Apply**. Render will:
   - provision Postgres and wire `DATABASE_URL` into the service automatically,
   - generate a random `JWT_SECRET`,
   - set `SEED_DEMO=false` (no demo data in production).
4. When the service is live, note its URL, e.g.
   `https://quota-backend.onrender.com`. Check `‹url›/health` returns
   `{"status":"ok"}`.

> The Go server auto-migrates the schema on boot, so the tables are created on
> first deploy. No manual migration step.

## 2. Frontend on Vercel

1. In [Vercel](https://vercel.com/new), import the same GitHub repo.
2. Set **Root Directory** to `frontend`. Framework preset: **Vite**
   (build `npm run build`, output `dist`).
3. Add an environment variable:
   - `VITE_API_BASE_URL` = your Render backend URL
     (e.g. `https://quota-backend.onrender.com`, no trailing slash)
4. **Deploy.** Note the resulting URL, e.g. `https://quota.vercel.app`.

## 3. Connect the two (CORS)

Back in Render → `quota-backend` → **Environment**:
- Set `CORS_ORIGIN` to your Vercel URL (e.g. `https://quota.vercel.app`).
- Save — Render redeploys. The backend already allows any `localhost` origin
  for dev; this adds your production frontend.

## 4. (Optional) Google sign-in

1. In Google Cloud Console, add your Vercel URL to the OAuth client's
   **Authorized JavaScript origins**.
2. In Render, set `GOOGLE_CLIENT_ID` on `quota-backend` and redeploy. The button
   appears automatically.

---

## Notes & gotchas

- **Free tiers sleep.** Render's free web service spins down when idle; the first
  request after a nap takes ~30s. Render's free Postgres expires after ~30 days —
  fine for testing, upgrade (or move to Supabase) before real customers.
- **To use Supabase instead of Render Postgres:** skip the `databases:` block and
  set `DATABASE_URL` on the service to your Supabase connection string
  (`...?sslmode=require`).
- **First admin account:** with `SEED_DEMO=false` there's no demo user. Create
  your org via the app's **Create account** form (it makes you the admin).
- **Rotate `JWT_SECRET`** only when you intend to invalidate all sessions.
