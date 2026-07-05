# Running Quota locally (restart guide)

The dev servers run as background processes and won't survive a machine restart.
Here's how to bring everything back up.

## Environment quirks on this machine

- **Go is installed at a non-standard path:** `%LOCALAPPDATA%\Programs\go`
  (it's not on PATH). Use the full path or add it to PATH:
  `C:\Users\HP\AppData\Local\Programs\go\bin`
- **Port 5173 is taken by another project (Spur)**, so the Vite dev server
  auto-picks a free port (e.g. 59416). The backend CORS already allows any
  localhost origin, so the dynamic port is fine.

## Start the backend (terminal 1)

```powershell
$env:Path = "C:\Users\HP\AppData\Local\Programs\go\bin;$env:Path"
cd "C:\Users\HP\Downloads\New saas project\backend"
go run ./cmd/server      # serves :8080, auto-migrates + seeds demo data
```

## Start the frontend (terminal 2)

```powershell
cd "C:\Users\HP\Downloads\New saas project\frontend"
npm run dev              # prints the URL, e.g. http://localhost:59416
```

Open the printed URL. Log in with **demo@quota.app** / **password123**.

## Handy commands

```powershell
# run the commission engine tests
cd backend; go test ./...

# production build + preview (what screenshots cleanly)
cd frontend; npm run build; npm run preview   # http://localhost:4173
```

## Where things stand

- MVP core loop is complete and verified: auth, comp plans, deals, attainment,
  transparent commission breakdown, payroll CSV export, role-based access.
- Google sign-in is coded — set `GOOGLE_CLIENT_ID` in `backend/.env` to enable.
- Next up (not started): Stripe billing + per-seat gating, CSV import of deals.
```
