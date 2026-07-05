---
name: testing-quota-frontend
description: Test the Quota frontend UI end-to-end. Use when verifying frontend changes, UI redesigns, or component updates.
---

# Testing Quota Frontend

## Local Dev Setup

1. Start backend: `export PATH=/usr/local/go/bin:$PATH && cd backend && go run ./cmd/server`
   - Runs on port 8080, auto-migrates DB and seeds demo data on first run
   - Uses SQLite locally (quota.db in backend dir)
2. Start frontend: `cd frontend && npm run dev`
   - Runs on port 5173 (or next available port if occupied)
   - Proxies `/api` requests to localhost:8080 via vite.config.ts

## Demo Credentials

- Email: `demo@quota.app`
- Password: `password123`
- Role: admin (sees all pages including manager-only routes)

## Pages to Test

| Page | Route | Role Required |
|------|-------|--------------|
| Login | `/login` | None |
| Dashboard | `/` | Any authenticated |
| Rep Detail | `/reps/:id` | Any authenticated |
| Deals | `/deals` | Manager/Admin |
| Comp Plans | `/comp-plans` | Manager/Admin |
| Team | `/team` | Manager/Admin |

## Key Things to Verify

- **Login page**: should be centered in viewport, brand mark visible, tabs work (Sign in / Create account)
- **Sidebar**: dark background, brand logo "Q" + "Quota", nav links with icons, active state highlights, user profile in footer
- **Dashboard**: 4 stat cards (Team Quota, Attained, Commission Owed with dark gradient, Active Reps), leaderboard with ranked rows and color-coded progress bars
- **Progress bars**: 4-tier color system: red (<40%), amber (40-69%), blue (70-99%), green (100%+)
- **Rep Detail**: 4-column header card, progress bar, commission breakdown summary, tier table with green total
- **Deals**: data table with type pills, Add Deal form toggle
- **Comp Plans**: plan cards in grid, period pills, tier lists
- **Team**: role badges (rep=gray, manager/admin=indigo), add member form, plan assignment dropdowns

## Common Issues

- Port conflicts: if 8080 or 5173 are already in use, kill the old process with `fuser -k <port>/tcp`
- Unicode in JSX: bare `\u2014` in JSX text renders as literal string; must use `{"\u2014"}` or actual `—` character
- Sidebar remount: the current layout wraps each route in `<AuthenticatedLayout>`, causing sidebar to technically remount on navigation. In practice this may not cause visible flicker but could be improved with React Router's `<Outlet>` pattern.

## Devin Secrets Needed

None — the demo account is seeded automatically and no external services are required for local testing.
