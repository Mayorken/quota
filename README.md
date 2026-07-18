# Quota — Commission & Quota Tracker

Commission and quota tracking for small B2B sales teams (5–50 reps). Define a
comp plan → enter deals → the system calculates attainment and commission with
fully **transparent math** → reps and managers see the same numbers → export for
payroll. That transparency is the product: it's what eliminates commission
disputes.

This repo is the MVP core loop described in the build spec.

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) for local
setup, validation, and pull request guidelines.

## Stack

- **Backend:** Go + Gin, GORM. Postgres in production (Supabase), zero-setup
  SQLite for local dev.
- **Frontend:** React + Vite + TypeScript.
- **Auth:** JWT (email/password), bcrypt hashing, role-based access
  (rep / manager / admin).

## Quick start

Two terminals.

### 1. Backend (`:8080`)

```bash
cd backend
go run ./cmd/server
```

On first run with no `DATABASE_URL`, it creates a local `quota.db` SQLite file,
auto-migrates the schema, and **seeds a demo org** so you can explore
immediately:

- **Manager login:** `demo@quota.app` / `password123`
- Reps: `alex@quota.app`, `sam@quota.app`, `jordan@quota.app` (same password)

### 2. Frontend (`:5173`)

```bash
cd frontend
npm install
npm run dev
```

Open http://localhost:5173 and click **"Use demo account"** on the login screen.
Vite proxies `/api` to the Go server on `:8080`.

## Enabling Google Sign-In

Email/password works out of the box. To add "Continue with Google":

1. In [Google Cloud Console → Credentials](https://console.cloud.google.com/apis/credentials),
   create an **OAuth 2.0 Client ID** of type **Web application**.
2. Under **Authorized JavaScript origins**, add your frontend origin
   (e.g. `http://localhost:5173`, or whatever port Vite prints).
3. Copy the Client ID into `backend/.env`:

   ```
   GOOGLE_CLIENT_ID=xxxxx.apps.googleusercontent.com
   ```

4. Restart the backend. The login screen now shows the Google button
   automatically — the frontend reads the Client ID from `GET /api/auth/config`,
   so there's nothing to configure on the frontend.

How it works: Google Identity Services returns an ID token in the browser; the
frontend posts it to `POST /api/auth/google`; the backend verifies it with
Google, then finds-or-creates the user (first Google sign-in provisions an org
with that user as admin) and issues the app's own JWT. Existing email/password
accounts are linked by email on first Google sign-in.

## Using Postgres / Supabase

Set `DATABASE_URL` in `backend/.env` (copy from `.env.example`):

```
DATABASE_URL=postgres://user:pass@host:5432/dbname?sslmode=require
```

The same models auto-migrate on Postgres. `tiers`, `type_multipliers`, and the
commission `breakdown` are stored as JSON columns.

## The commission engine

The calculation lives in [`backend/internal/commission/engine.go`](backend/internal/commission/engine.go)
as a **pure function** with no database or HTTP dependency, so it's unit-tested
in isolation:

```bash
cd backend
go test ./...
```

It supports:

- **Flat rate** — one tier covering all revenue.
- **Accelerators** — e.g. 5% up to quota, 8% above (uncapped top tier).
- **Multi-tier accelerators** — any number of bands expressed as % of quota.
- **Deal-type multipliers** — e.g. renewals credited at 50% toward attainment.

Every payout produces an itemized `Breakdown` (revenue per tier, rate, and
commission per line) that is shown identically to reps and managers and included
in the CSV export.

## Data model

`organizations · users · comp_plans · rep_comp_assignments · deals ·
commission_calculations` — see [`backend/internal/models/models.go`](backend/internal/models/models.go).

## API surface

| Method | Path | Access | Purpose |
|--------|------|--------|---------|
| GET | `/api/auth/config` | public | Public client config (Google client ID) |
| POST | `/api/auth/signup` | public | Create org + first admin |
| POST | `/api/auth/login` | public | Get JWT |
| POST | `/api/auth/google` | public | Sign in / provision via Google ID token |
| GET | `/api/auth/me` | auth | Current user |
| GET | `/api/dashboard` | auth | Attainment (reps see only self) |
| GET | `/api/reps/:id/commission` | auth | Full breakdown for a rep |
| GET | `/api/deals` | auth | List deals (reps see only own) |
| POST | `/api/deals` | manager | Add a deal |
| DELETE | `/api/deals/:id` | manager | Remove a deal |
| GET/POST/PUT/DELETE | `/api/comp-plans` | manager | Comp plan CRUD |
| POST | `/api/comp-plans/assign` | manager | Assign a rep to a plan |
| GET | `/api/users`, POST `/api/users` | manager | Team management |
| GET | `/api/export/commissions.csv` | manager | Payroll CSV |
| POST | `/api/reps/:id/finalize` | manager | Snapshot a period (draft/approved/paid) |

## What's intentionally not here yet (post-revenue)

Per the spec, these wait for real customer demand: CSV **import** of deals,
Stripe billing + per-seat gating, CRM sync, Slack notifications, multi-currency,
approval workflows, mobile. The billing integration is the natural next build
(Week 5 in the plan) — the org model already carries `stripe_customer_id` and
`plan_tier` for it.
```
