# Multi-Carrier Shipping Platform

A complete microservices-based shipping platform built in Go with Gin, Kafka, and Docker.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         CLIENT (React/Web/Mobile)                    │
└──────────────────────────────┬──────────────────────────────────────┘
                               │ HTTP/REST
                               ▼
                    ┌──────────────────────┐
                    │     API Gateway        │
                    │      Port: 8080        │
                    │  Auth, Routing, Logging│
                    └──────────┬───────────┘
                               │
        ┌──────────────────────┼──────────────────────┐
        │                      │                      │
   HTTP/REST              HTTP/REST              HTTP/REST
        ▼                      ▼                      ▼
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│  Shipment    │    │   Carrier    │    │    Rate      │
│  Service     │    │ Integration  │    │ Comparison   │
│  Port: 8081  │    │ Port: 8082   │    │ Port: 8083   │
│  PostgreSQL  │    │ PostgreSQL   │    │ PostgreSQL   │
└──────┬───────┘    └──────────────┘    └──────┬───────┘
       │                                          │
       │ Kafka                                    │
       ▼                                          ▼
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│   Label      │    │  Tracking    │    │   Address    │
│ Generation   │    │  Service     │    │ Validation   │
│ Port: 8084   │    │ Port: 8085   │    │ Port: 8086   │
│ PostgreSQL   │    │ PostgreSQL   │    │ PostgreSQL   │
└──────────────┘    └──────┬───────┘    └──────────────┘
                           │
              ┌────────────┼────────────┐
              │            │            │
              ▼            ▼            ▼
       ┌──────────┐ ┌──────────┐ ┌──────────┐
       │ Billing  │ │ Return   │ │Notification│
       │ Service  │ │ Service  │ │ Service    │
       │Port:8087 │ │Port:8088 │ │ (Consumer) │
       │PostgreSQL│ │PostgreSQL│ │            │
       └──────────┘ └──────────┘ └──────────┘
```

## Services

| # | Service | Port | Database | Kafka Role | Description |
|---|---------|------|----------|------------|-------------|
| 1 | **API Gateway** | 8080 | — | — | Entry point, auth, routing |
| 2 | **Shipment Management** | 8081 | PostgreSQL | Producer | CRUD shipments, status management |
| 3 | **Carrier Integration** | 8082 | PostgreSQL | — | DHL, FedEx, UPS API clients |
| 4 | **Rate Comparison** | 8083 | PostgreSQL | Producer | Compare rates across carriers |
| 5 | **Label Generation** | 8084 | PostgreSQL | Producer | Generate shipping labels |
| 6 | **Real-time Tracking** | 8085 | PostgreSQL | Consumer/Producer | Track shipments, poll carriers |
| 7 | **Address Validation** | 8086 | PostgreSQL | Producer | Validate addresses, find locations |
| 8 | **Billing & Invoice** | 8087 | PostgreSQL | Producer | Payments, invoices (Stripe) |
| 9 | **Customer Notification** | — | — | Consumer | Email/SMS notifications |
| 10 | **Return & Reverse** | 8088 | PostgreSQL | Producer | Return requests, refunds |

## Kafka Topics

| Topic | Producer | Consumers | Purpose |
|-------|----------|-----------|---------|
| `shipment.created` | Shipment Service | Tracking, Notification | New shipment |
| `shipment.updated` | Shipment Service | Notification | Shipment modified |
| `shipment.status.changed` | Shipment Service | Notification | Status update |
| `rates.compared` | Rate Service | — | Rate comparison done |
| `label.generated` | Label Service | — | Label created |
| `tracking.updated` | Tracking Service | Notification | Tracking event |
| `address.validated` | Address Service | — | Address validated |
| `payment.processed` | Billing Service | Notification | Payment done |
| `invoice.generated` | Billing Service | — | Invoice created |
| `return.created` | Return Service | Notification | Return requested |
| `return.status.changed` | Return Service | Notification | Return updated |

## Quick Start

### 1. Start Everything (Docker)
```bash
make build
make up
```

### 2. Apply Migrations
```bash
# Shipment DB
psql postgres://postgres:postgres@localhost:5431/shipments -f shipment-service/migrations/001_create_shipments.sql

# Carrier DB
psql postgres://postgres:postgres@localhost:5432/carriers -f carrier-integration-service/migrations/001_create_carriers.sql

# Rate DB
psql postgres://postgres:postgres@localhost:5433/rates -f rate-comparison-service/migrations/001_create_rates.sql

# Label DB
psql postgres://postgres:postgres@localhost:5434/labels -f label-generation-service/migrations/001_create_labels.sql

# Tracking DB
psql postgres://postgres:postgres@localhost:5435/tracking -f tracking-service/migrations/001_create_tracking.sql

# Address DB
psql postgres://postgres:postgres@localhost:5436/addresses -f address-validation-service/migrations/001_create_addresses.sql

# Billing DB
psql postgres://postgres:postgres@localhost:5437/billing -f billing-service/migrations/001_create_billing.sql

# Return DB
psql postgres://postgres:postgres@localhost:5438/returns -f return-service/migrations/001_create_returns.sql
```

### 3. Test the Flow
```bash
make test-flow
```

### 4. Development Mode (Local Go, Docker for Infra)
```bash
make dev

# Then in separate terminals:
cd shipment-service && go mod tidy && go run ./cmd
cd carrier-integration-service && go mod tidy && go run ./cmd
cd rate-comparison-service && go mod tidy && go run ./cmd
cd label-generation-service && go mod tidy && go run ./cmd
cd tracking-service && go mod tidy && go run ./cmd
cd address-validation-service && go mod tidy && go run ./cmd
cd billing-service && go mod tidy && go run ./cmd
cd return-service && go mod tidy && go run ./cmd
cd notification-service && go mod tidy && go run ./cmd
cd api-gateway && go mod tidy && go run ./cmd
```

## API Endpoints

### Shipment Management
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/shipments` | Create shipment |
| GET | `/shipments` | List user shipments |
| GET | `/shipments/:id` | Get shipment |
| PUT | `/shipments/:id` | Update shipment |
| PATCH | `/shipments/:id/status` | Update status |
| DELETE | `/shipments/:id` | Delete shipment |

### Carrier Integration
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/carriers` | Register carrier |
| GET | `/carriers/rates` | Get rates from all carriers |
| GET | `/carriers/tracking` | Track shipment |
| GET | `/carriers/pickup-locations` | Find pickup points |
| GET | `/carriers/drop-locations` | Find drop-off points |

### Rate Comparison
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/rates/compare` | Compare all carrier rates |
| GET | `/rates/comparison` | Get comparison result |

### Label Generation
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/labels` | Generate label |
| GET | `/labels/:id` | Get label |
| GET | `/labels/:id/download` | Download PDF |

### Tracking
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/tracking/:shipment_id` | Get tracking history |
| POST | `/tracking/events` | Add tracking event |

### Address Validation
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/addresses/validate` | Validate address |
| GET | `/addresses/pickup-locations` | Get pickup locations |
| GET | `/addresses/drop-locations` | Get drop-off locations |

### Billing
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/billing/invoices` | Create invoice |
| POST | `/billing/payments` | Process payment |
| GET | `/billing/invoices/:id` | Get invoice |

### Returns
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/returns` | Request return |
| POST | `/returns/:id/approve` | Approve return |
| POST | `/returns/:id/refund` | Process refund |
| GET | `/returns/:id` | Get return |

## Design Decisions

1. **Database-per-Service**: Each service owns its own PostgreSQL instance — no shared databases
2. **API Gateway Pattern**: Single entry point handles auth, routing, and request logging
3. **Async Communication**: Kafka decouples services (e.g., Shipment Service doesn't know Notification Service exists)
4. **Clean Architecture**: `handler → service → repository → domain` layers in every service
5. **Carrier Abstraction**: `CarrierClient` interface with implementations for DHL, FedEx, UPS — easy to add more
6. **Event-Driven Notifications**: Notification Service listens to Kafka topics, sends emails automatically

## Adding a New Carrier

1. Create a new client in `carrier-integration-service/internal/client/`
2. Implement the `CarrierClient` interface
3. Add to `CarrierClientFactory`
4. Register via API: `POST /carriers`

## Project Structure

```
multi-carrier-shipping/
├── api-gateway/              # Reverse proxy
├── shipment-service/         # Shipment CRUD + Kafka producer
├── carrier-integration-service/  # DHL, FedEx, UPS clients
├── rate-comparison-service/  # Compare rates across carriers
├── label-generation-service/ # Generate shipping labels
├── tracking-service/         # Track shipments + Kafka consumer/producer
├── address-validation-service/ # Validate addresses, find locations
├── billing-service/          # Invoices, payments (Stripe)
├── notification-service/   # Email/SMS consumer (no HTTP)
├── return-service/         # Return requests, refunds
├── shared/                 # Common libraries
│   ├── pkg/kafka/          # Producer, Consumer, Topics
│   ├── pkg/logger/         # Zap logging
│   ├── pkg/middleware/     # Auth, request logging
│   ├── pkg/errors/         # Common errors
│   └── pkg/utils/          # ID generation
├── docker-compose.yml      # Full stack orchestration
├── Makefile              # Common commands
└── .env.example          # Environment variables
```

## Carrier APIs

### DHL
- **Rates**: `POST /rates` — Get shipping rates
- **Tracking**: `GET /tracking` — Track shipments
- **Locations**: `GET /locations` — Find service points
- **Labels**: `POST /labels` — Create shipping labels

### FedEx
- **Rates API**: `POST /rate/v1/rates/quotes`
- **Track API**: `POST /track/v1/trackingnumbers`
- **Address Validation**: `POST /address/v1/addresses/resolve`

### UPS
- **Rating API**: `POST /api/rating/v1/Rate`
- **Tracking API**: `GET /api/track/v1/details/{inquiryNumber}`
- **Locator API**: `GET /api/locations/v1/search/availabilities/{geocode}`

> **Note**: The current implementation uses simulated responses. Replace with actual API calls using your real API credentials.

## Next Steps

1. Replace simulated carrier APIs with real DHL/FedEx/UPS API integrations
2. Add JWT authentication with refresh tokens
3. Implement distributed tracing with OpenTelemetry
4. Add Redis caching for address validation and carrier rates
5. Deploy to Kubernetes with Helm charts
6. Add Prometheus metrics and Grafana dashboards

## Release Management

The platform includes automated tools to build, tag, and publish Docker images for all 10 microservices to a container registry (e.g. GitHub Container Registry or Docker Hub).

### 1. Build and Tag Images Locally
To compile the microservices and tag them locally under a release version without publishing them:
```bash
make release-local TAG=v1.0.0
```

### 2. Build and Publish to Registry
To build, tag, and publish the images to the registry (defaults to `ghcr.io/kavix/multi-carrier-shipping-golang` but can be overridden with `REGISTRY`):
```bash
make release TAG=v1.0.0 REGISTRY=ghcr.io/your-username/your-repo
```

You can also run the release script directly to enable interactive tagging and draft a GitHub release using the GitHub CLI (`gh`):
```bash
./release-images.sh --github-release v1.0.0
```
