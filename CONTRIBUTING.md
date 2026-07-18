# Contributing to Quota

Thanks for helping improve Quota. Keep contributions focused, easy to review,
and backed by the relevant checks.

## Local setup

1. Start the Go API:

   ```bash
   cd backend
   go run ./cmd/server
   ```

2. In another terminal, start the frontend:

   ```bash
   cd frontend
   npm install
   npm run dev
   ```

3. Open `http://localhost:5173` and use the demo account described in the
   README.

## Before opening a pull request

- Create a short, descriptive branch from `main`.
- Keep unrelated changes in separate pull requests.
- Add or update tests when behavior changes.
- Never commit `.env` files, credentials, database files, or customer data.
- Run the checks for every area you changed.

### Backend checks

```bash
cd backend
go test ./...
```

### Frontend checks

```bash
cd frontend
npm ci
npm run build
```

## Pull request notes

Explain what changed, why it changed, and how you verified it. For interface
changes, include a screenshot or short recording. For commission-calculation
changes, include an example showing the expected payout breakdown.

