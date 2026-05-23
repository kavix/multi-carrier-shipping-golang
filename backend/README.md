# Backend Microservices

This directory contains Go microservices for the multi-carrier shipping platform.

## Services

- `cmd/api-gateway` — HTTP gateway that proxies requests to service backends
- `cmd/order-service` — shipping quote generation and order support
- `cmd/carrier-service` — carrier list and rate data
- `cmd/tracking-service` — shipment tracking

## Run

From `backend`:

```bash
go run ./cmd/order-service
go run ./cmd/carrier-service
go run ./cmd/tracking-service
go run ./cmd/api-gateway
```

The gateway listens on `http://localhost:8080` by default.

## Endpoints

- `GET /api/carriers`
- `GET /api/quote?origin=NYC&destination=LAX&weight=12`
- `GET /api/track?trackingNumber=TRACK123`
