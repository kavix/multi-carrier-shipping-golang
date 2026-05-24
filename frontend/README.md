# Multi-Carrier Shipping — Frontend

This is a minimal React frontend to exercise the API Gateway and backend services in the `multi-carrier-shipping-golang` monorepo.

Features
- Small UI to call common endpoints from `docs/API-GUIDE.md`
- Configure API base URL and Authorization token
- Simple forms for creating shipments, comparing rates, generating labels, validating addresses, billing, and returns

Getting started

1. Install dependencies

```bash
cd frontend
npm install
```

2. Run in development (Vite)

```bash
npm run dev
```

3. Open the app

Visit http://localhost:5173 (or the URL printed by Vite).

Configuration
- The app defaults to `http://localhost:8080` as the API gateway base URL. You can change it in the UI or set the environment variable `VITE_API_URL` before starting Vite.

Security
- Do not store secrets in the UI. Provide an Authorization token in the UI header field if your API requires authentication.

Notes
- This frontend is intentionally minimal and intended for local development and manual testing. You can expand components and add proper routing, validation, and authentication flows as needed.

