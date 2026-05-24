# System Architecture

## Overview

Multi-Carrier Shipping Platform is a distributed microservices-based system designed to handle complex shipping workflows across multiple carriers (DHL, FedEx, UPS, etc.). It uses event-driven architecture to decouple services and enable asynchronous processing.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                    CLIENT (Web/Mobile/API)                           │
└──────────────────────────┬──────────────────────────────────────────┘
                           │ HTTP/REST
                           ▼
                ┌──────────────────────┐
                │     API Gateway      │
                │      Port: 8080      │
                │  - Auth/Token Check  │
                │  - Request Routing   │
                │  - Load Balancing    │
                └──────────┬───────────┘
                           │
        ┌──────────────────┼──────────────────┬────────────────┐
        │                  │                  │                │
   HTTP/REST          HTTP/REST          HTTP/REST       HTTP/REST
        │                  │                  │                │
   ┌────▼─────┐   ┌────────▼─────┐   ┌────────▼────┐   ┌────▼──────┐
   │ Shipment  │   │  Carrier     │   │    Rate     │   │  Label    │
   │ Service   │   │ Integration  │   │ Comparison  │   │ Generation│
   │(8081)     │   │   (8082)     │   │   (8083)    │   │  (8084)   │
   └────┬──────┘   └──────────────┘   └─────┬───────┘   └───────────┘
        │                                     │
        │ Kafka Events                        │
        ▼                                     ▼
   ┌────────────────┐   ┌──────────────┐   ┌────────────────────┐
   │   Tracking     │   │   Address    │   │  API Gateway       │
   │   Service      │   │  Validation  │   │  (Notification     │
   │   (8085)       │   │   (8086)     │   │   Consumer)        │
   └────────────────┘   └──────────────┘   └────────────────────┘
        │                     │
        │ Kafka               │ Kafka
        │ (Events)            │ (Events)
        ▼                     ▼
   ┌──────────┐        ┌──────────┐
   │ Billing  │        │ Return   │
   │ Service  │        │ Service  │
   │ (8087)   │        │ (8088)   │
   └──────────┘        └──────────┘
```

## Detailed Architecture

### 1. API Gateway (Port 8080)

**Role**: Single entry point for all client requests

**Responsibilities**:
- Route requests to appropriate microservices
- Validate authentication tokens
- Log all requests
- Handle cross-cutting concerns (CORS, compression)

**Key Features**:
- Service discovery via environment variables
- Request proxying with headers preservation
- Error handling and circuit breaking

### 2. Service Topology

#### Core Microservices

| Service | Port | Database | Kafka Role | Purpose |
|---------|------|----------|-----------|---------|
| **Shipment** | 8081 | PostgreSQL | Producer | Core shipment management, CRUD operations |
| **Carrier Integration** | 8082 | PostgreSQL | - | Multi-carrier API clients (DHL, FedEx, UPS) |
| **Rate Comparison** | 8083 | PostgreSQL | Producer | Compare rates across all configured carriers |
| **Label Generation** | 8084 | PostgreSQL | Producer | Generate shipping labels, PDF creation |
| **Tracking** | 8085 | PostgreSQL | Consumer/Producer | Real-time tracking, status updates |
| **Address Validation** | 8086 | PostgreSQL | Producer | Address validation, geocoding, location services |
| **Billing & Invoice** | 8087 | PostgreSQL | Producer | Stripe payments, invoice generation |
| **Return & Reverse** | 8088 | PostgreSQL | Producer | Return management, refunds, reverse logistics |
| **Notification** | - | - | Consumer | Email/SMS notifications from Kafka events |

### 3. Data Storage

#### Database-Per-Service Pattern

Each service has its own PostgreSQL database:
- **Isolation**: Services cannot directly access other service databases
- **Scalability**: Each DB can be scaled independently
- **Autonomy**: Services manage their own schema changes
- **Reliability**: Database failure in one service doesn't cascade

```
┌─────────────────────────────────────────┐
│         Service 1                       │
│    ┌─────────────────────┐              │
│    │ PostgreSQL DB       │              │
│    │ (Shipments Table)   │              │
│    └─────────────────────┘              │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│         Service 2                       │
│    ┌─────────────────────┐              │
│    │ PostgreSQL DB       │              │
│    │ (Rates Table)       │              │
│    └─────────────────────┘              │
└─────────────────────────────────────────┘
```

### 4. Message Queue Architecture

**Technology**: Apache Kafka with Zookeeper

**Topics**:
```
shipment.created           ← Shipment Service
shipment.updated           ← Shipment Service
shipment.status.changed    ← Shipment Service
rates.compared             ← Rate Comparison Service
label.generated            ← Label Generation Service
tracking.updated           ← Tracking Service
address.validated          ← Address Validation Service
payment.processed          ← Billing Service
invoice.generated          ← Billing Service
return.created             ← Return Service
return.status.changed      ← Return Service
```

**Event Flow**:
1. Service A performs action → publishes event to Kafka
2. Service B listens to topic → receives event
3. Service B processes event asynchronously
4. No direct coupling between services

### 5. Communication Patterns

#### Synchronous (HTTP/REST)
- Client → API Gateway
- API Gateway → Individual Services
- Service → Carrier Integration Service (for rates, tracking)
- Service → Address Validation Service (for address lookup)

**Use Case**: Real-time operations requiring immediate response

#### Asynchronous (Kafka)
- Service publishes event after action
- Other interested services listen and react
- Decouples temporal coupling

**Use Case**: Notifications, auditing, eventual consistency

### 6. Service-to-Service Communication

```
Shipment Service creates shipment
         ↓
Publishes "shipment.created" event to Kafka
         ↓
         ├─→ Tracking Service listens
         │   ├─ Creates initial tracking record
         │   └─ Publishes "tracking.updated"
         │
         ├─→ Notification Service listens
         │   └─ Sends confirmation email
         │
         └─→ Billing Service listens
             └─ Creates invoice

Direct HTTP calls (when immediate response needed):
- Rate Comparison → Carrier Integration (get rates)
- Label Generation → Carrier Integration (get carrier details)
- Address Validation → Carrier Integration (for pickup locations)
```

## Design Principles

### 1. Clean Architecture

Every service follows this layered structure:

```
┌────────────────────────────────────────┐
│        HTTP Layer (Handler)            │
│   - Request parsing                    │
│   - Response formatting                │
│   - HTTP status codes                  │
└────────────────────────────────────────┘
                  ↓
┌────────────────────────────────────────┐
│     Business Logic (Service)           │
│   - Validation                         │
│   - Orchestration                      │
│   - Rules and workflows                │
└────────────────────────────────────────┘
                  ↓
┌────────────────────────────────────────┐
│      Data Access (Repository)          │
│   - Database queries                   │
│   - Transaction management             │
│   - Query optimization                 │
└────────────────────────────────────────┘
                  ↓
┌────────────────────────────────────────┐
│      Domain Models (Domain)            │
│   - Business entities                  │
│   - Value objects                      │
│   - Constraints                        │
└────────────────────────────────────────┘
```

### 2. Dependency Inversion

- High-level modules depend on abstractions
- Example: `CarrierClient` interface with multiple implementations

```go
type CarrierClient interface {
    GetRates(...) ([]CarrierRate, error)
    GetTracking(...) (*TrackingInfo, error)
    GetPickupLocations(...) ([]Location, error)
}

// Implementations:
type DHLClient struct { ... }
type FedExClient struct { ... }
type UPSClient struct { ... }
```

### 3. Separation of Concerns

- Each service handles specific domain
- Kafka separates temporal concerns
- API Gateway handles cross-cutting concerns

### 4. Scalability

- **Horizontal Scaling**: Deploy multiple instances of any service
- **Load Balancing**: Services behind load balancer
- **Database Scaling**: Each service's DB scaled independently
- **Kafka Scalability**: Topics can have multiple partitions

## Security Architecture

### 1. Authentication

- **Token-based**: JWT or custom tokens
- **API Gateway**: Validates all tokens before routing
- **Header Propagation**: Authorization header passed to downstream services

```
Client Request
    ↓
Authorization: Bearer <token>
    ↓
API Gateway validates token
    ↓
Token valid? → Route to service with header
Token invalid? → Return 401 Unauthorized
```

### 2. Service-to-Service Communication

- Internal network (Docker network)
- Services accessible only within container network
- External access only through API Gateway

### 3. Database Security

- Each service has own database credentials
- Credentials via environment variables
- No hardcoded secrets

## Deployment Architecture

### Development
```
Single Docker Compose file
  ├─ All services in containers
  ├─ All databases in containers
  ├─ Kafka/Zookeeper in containers
  └─ Shared Docker network
```

### Production

```
Kubernetes Cluster
  ├─ Multiple nodes
  ├─ Deployment manifests per service
  ├─ ConfigMaps for configuration
  ├─ Secrets for credentials
  ├─ StatefulSets for databases
  └─ Services for networking
```

## Error Handling

### Strategy

1. **Validation**: Input validation at handler level
2. **Business Logic**: Domain error checks in service
3. **Infrastructure**: Retry logic in repository
4. **External APIs**: Circuit breaker for carrier APIs

### Propagation

```
Client → API Gateway
           ↓
       Service handler
           ↓
       Service layer (business logic)
           ↓
       Repository (database)
           ↓
       Error: returns error with context
           ↓
       Logged and returned to client
```

## Monitoring & Observability

### Logging

- Structured logging to stdout
- JSON format for easy parsing
- Context propagation across services

### Metrics

- Service health endpoints
- Request latency
- Database query performance
- Kafka message processing

### Tracing

- Correlation IDs in requests
- Trace propagation through services
- End-to-end request tracing

## Scalability Considerations

### Horizontal Scaling

1. **Stateless Services**: All services are stateless (no session storage)
2. **Load Balancing**: Multiple instances handled by load balancer
3. **Kafka Partitioning**: Topics partitioned for parallelism

### Vertical Scaling

- Increase CPU/memory for resource-intensive services
- Database optimization with indexes
- Connection pooling for databases

### Caching

- In-memory caching for frequently accessed data
- Cache invalidation on updates
- Distributed cache for multi-instance scenarios

## Future Enhancements

1. **Circuit Breaker**: Resilience for external API calls
2. **Service Mesh**: Istio/Linkerd for advanced routing
3. **API Versioning**: Backward compatibility handling
4. **GraphQL Layer**: Alternative to REST
5. **Caching Layer**: Redis for performance
6. **Rate Limiting**: Per-user request limits
7. **Advanced Analytics**: ML for rate prediction
