# Multi-Carrier Shipping Platform

This workspace contains a Next.js frontend and a Go backend microservice architecture for a multi-carrier shipping platform.

## Structure

- `frontend/` — Next.js application with TypeScript and App Router.
- `backend/` — Go backend containing independent microservices:
  - `api-gateway` — public gateway and aggregator
  - `order-service` — quote generation and order creation
  - `carrier-service` — carrier inventory and rates
  - `tracking-service` — shipment tracking

## Run locally

1. Start the backend services:
   - `cd backend`
   - `go run ./cmd/order-service`
   - `go run ./cmd/carrier-service`
   - `go run ./cmd/tracking-service`
   - `go run ./cmd/api-gateway`

2. Start the frontend:
   - `cd frontend`
   - `npm install`
   - `npm run dev`

3. Open `http://localhost:3000`

## Notes

- The frontend calls the gateway at `http://localhost:8080`.
- Each backend microservice listens on its own port and the gateway aggregates responses.
