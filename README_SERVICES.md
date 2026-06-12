# Multi-Carrier Shipping Platform Architecture

This document outlines the microservices architecture, how services are interconnected, and a detailed API reference for each service.

## 🏗️ Architecture Overview

The platform is built using a **Microservices Architecture** with a **Database-per-Service** pattern. It leverages both synchronous and asynchronous communication to ensure high availability and loose coupling.

### 🔌 Service Connections

1.  **Synchronous (HTTP/REST)**:
    *   **API Gateway**: Acts as the single entry point. It handles authentication (JWT/Bearer) and proxies requests to downstream services.
    *   **Rate Comparison Service → Carrier Service**: Calls the Carrier Service synchronously to fetch real-time rates from external carrier APIs.
    *   **Return Service → Carrier Service**: Calls the Carrier Service to fetch carrier-specific return locations.

2.  **Asynchronous (Kafka Events)**:
    *   **Shipment Service**: Publishes `shipment.created`, `shipment.updated`, and `shipment.status.changed`.
    *   **Label Generation Service**: Consumes `shipment.created` to automatically generate shipping labels and updates the shipment record via a private event.
    *   **Address Validation Service**: Consumes `shipment.created` to validate addresses and emits `shipment.address.validated`.
    *   **Billing Service**: Consumes `shipment.created` to generate invoices and publishes `payment.processed` upon success.
    *   **Notification Service**: Consumes all major events to send Email/SMS alerts to users.

---

## 🛠️ Service API Reference

All requests to the API Gateway require an `Authorization` header:
`Authorization: Bearer test-token`

### 1. Shipment Service
*Core service for managing the shipment lifecycle.*

───────────────────────────┬─────────────┬───────────────────────────────────────────────┬─────────────────────────────┐
│ Endpoint                  │ HTTP Method │ Description                                   │ Technology Used             │
├───────────────────────────┼─────────────┼───────────────────────────────────────────────┼─────────────────────────────┤
│ `/api/shipments`          │ POST        │ Create a new shipment                         │ Go (Gin), PostgreSQL, Kafka │
│ `/api/shipments`          │ GET         │ List all shipments for the user               │ Go (Gin), PostgreSQL        │
│ `/api/shipments/{id}`     │ GET         │ Get details of a specific shipment            │ Go (Gin), PostgreSQL        │
│ `/api/shipments/{id}`     │ PUT         │ Update shipment details (Pending only)        │ Go (Gin), PostgreSQL, Kafka │
│ `/api/shipments/{id}/status`│ PATCH       │ Update shipment status (Internal/Admin)       │ Go (Gin), PostgreSQL, Kafka │
│ `/api/shipments/{id}`     │ DELETE      │ Cancel and delete a pending shipment          │ Go (Gin), PostgreSQL, Kafka │
└───────────────────────────┴─────────────┴───────────────────────────────────────────────┴─────────────────────────────┘

#### Example CURLs:
```bash
# Create Shipment
curl -X POST -H "Authorization: Bearer test-token" -H "Content-Type: application/json" -d '{"sender_name": "John Doe", "sender_address": "123 Main St, NY", "receiver_name": "Jane Smith", "receiver_address": "456 Oak Ave, LA", "weight": 2.5, "carrier": "fedex", "service_type": "FEDEX_GROUND"}' http://localhost:8080/shipments

# List Shipments
curl -H "Authorization: Bearer test-token" http://localhost:8080/shipments
```

### 2. Carrier Integration Service
*Gateway to external carrier APIs (DHL, FedEx, UPS).*

───────────────────────────────┬─────────────┬───────────────────────────────────────────────┬─────────────────────────────┐
│ Endpoint                      │ HTTP Method │ Description                                   │ Technology Used             │
├───────────────────────────────┼─────────────┼───────────────────────────────────────────────┼─────────────────────────────┤
│ `/api/carriers`               │ POST        │ Register a new carrier integration            │ Go (Gin), PostgreSQL        │
│ `/api/carriers/rates`         │ GET         │ Fetch live rates from carrier APIs            │ Go (Gin), External APIs     │
│ `/api/carriers/tracking`      │ GET         │ Get live tracking data from carrier           │ Go (Gin), External APIs     │
│ `/api/carriers/pickup-locations`│ GET         │ Find carrier pickup points near address       │ Go (Gin), External APIs     │
└───────────────────────────────┴─────────────┴───────────────────────────────────────────────┴─────────────────────────────┘

#### Example CURLs:
```bash
# Get Live Rates
curl -H "Authorization: Bearer test-token" "http://localhost:8080/carriers/rates?from=10001&to=90001&weight=5"

# Get Live Tracking
curl -H "Authorization: Bearer test-token" "http://localhost:8080/carriers/tracking?carrier=fedex&tracking_number=123456789"
```

### 3. Rate Comparison Service
*Aggregates and compares rates to find the best deal.*

───────────────────────────┬─────────────┬───────────────────────────────────────────────┬─────────────────────────────┐
│ Endpoint                  │ HTTP Method │ Description                                   │ Technology Used             │
├───────────────────────────┼─────────────┼───────────────────────────────────────────────┼─────────────────────────────┤
│ `/api/rates/compare`      │ POST        │ Compare rates for a specific shipment         │ Go (Gin), Kafka, HTTP Client│
│ `/api/rates/comparison`   │ GET         │ Retrieve a saved rate comparison              │ Go (Gin), PostgreSQL        │
└───────────────────────────┴─────────────┴───────────────────────────────────────────────┴─────────────────────────────┘

#### Example CURLs:
```bash
# Compare Rates
curl -X POST -H "Authorization: Bearer test-token" -H "Content-Type: application/json" -d '{"shipment_id": "SHIP-123", "from": "New York", "to": "Los Angeles", "weight": 2.5}' http://localhost:8080/rates/compare
```

### 4. Label Generation Service
*Produces shipping labels (PDF/ZPL) and integrates with S3.*

───────────────────────────┬─────────────┬───────────────────────────────────────────────┬─────────────────────────────┐
│ Endpoint                  │ HTTP Method │ Description                                   │ Technology Used             │
├───────────────────────────┼─────────────┼───────────────────────────────────────────────┼─────────────────────────────┤
│ `/api/labels`             │ POST        │ Manually trigger label generation             │ Go (Gin), Kafka, AWS S3     │
│ `/api/labels/{id}`        │ GET         │ Get label metadata                            │ Go (Gin), PostgreSQL        │
│ `/api/labels/{id}/download`│ GET         │ Download the label PDF file                   │ Go (Gin), S3/Local Storage  │
└───────────────────────────┴─────────────┴───────────────────────────────────────────────┴─────────────────────────────┘

#### Example CURLs:
```bash
# Download Label
curl -H "Authorization: Bearer test-token" http://localhost:8080/labels/LABEL-123/download --output label.pdf
```

### 5. Tracking Service
*Unified tracking history and event logging.*

───────────────────────────┬─────────────┬───────────────────────────────────────────────┬─────────────────────────────┐
│ Endpoint                  │ HTTP Method │ Description                                   │ Technology Used             │
├───────────────────────────┼─────────────┼───────────────────────────────────────────────┼─────────────────────────────┤
│ `/api/tracking/{id}`      │ GET         │ Get full tracking history for a shipment      │ Go (Gin), PostgreSQL, Kafka │
└───────────────────────────┴─────────────┴───────────────────────────────────────────────┴─────────────────────────────┘

#### Example CURLs:
```bash
# Get Tracking History
curl -H "Authorization: Bearer test-token" http://localhost:8080/tracking/SHIP-123
```

### 6. Address Validation Service
*Standardizes addresses and finds locations.*

───────────────────────────────┬─────────────┬───────────────────────────────────────────────┬─────────────────────────────┐
│ Endpoint                      │ HTTP Method │ Description                                   │ Technology Used             │
├───────────────────────────────┼─────────────┼───────────────────────────────────────────────┼─────────────────────────────┤
│ `/api/addresses/validate`     │ POST        │ Validate and standardize an address           │ Go (Gin), External APIs     │
│ `/api/addresses/pickup-locations`│ GET         │ Find nearby pickup locations                  │ Go (Gin), HTTP Client       │
└───────────────────────────────┴─────────────┴───────────────────────────────────────────────┴─────────────────────────────┘

#### Example CURLs:
```bash
# Validate Address
curl -X POST -H "Authorization: Bearer test-token" -H "Content-Type: application/json" -d '{"address": "123 Main St, New York"}' http://localhost:8080/addresses/validate
```

### 7. Billing Service
*Handles invoices, payments, and refunds.*

───────────────────────────┬─────────────┬───────────────────────────────────────────────┬─────────────────────────────┐
│ Endpoint                  │ HTTP Method │ Description                                   │ Technology Used             │
├───────────────────────────┼─────────────┼───────────────────────────────────────────────┼─────────────────────────────┤
│ `/api/billing/invoices`   │ POST        │ Manually create an invoice                    │ Go (Gin), PostgreSQL, Kafka │
│ `/api/billing/invoices`   │ GET         │ List invoices (filterable by shipment_id)     │ Go (Gin), PostgreSQL        │
│ `/api/billing/payments`   │ POST        │ Process a payment for an invoice              │ Go (Gin), Stripe, Kafka     │
└───────────────────────────┴─────────────┴───────────────────────────────────────────────┴─────────────────────────────┘

#### Example CURLs:
```bash
# Process Payment
curl -X POST -H "Authorization: Bearer test-token" -H "Content-Type: application/json" -d '{"invoice_id": "INV-123", "method": "stripe"}' http://localhost:8080/billing/payments
```

### 8. Return Service
*Manages the return lifecycle.*

───────────────────────────┬─────────────┬───────────────────────────────────────────────┬─────────────────────────────┐
│ Endpoint                  │ HTTP Method │ Description                                   │ Technology Used             │
├───────────────────────────┼─────────────┼───────────────────────────────────────────────┼─────────────────────────────┤
│ `/api/returns`            │ POST        │ Initiate a return for a shipment              │ Go (Gin), PostgreSQL, Kafka │
│ `/api/returns/{id}/approve`│ POST        │ Approve a return and trigger label generation │ Go (Gin), PostgreSQL        │
│ `/api/returns/{id}/refund` │ POST        │ Process a refund for a returned item          │ Go (Gin), PostgreSQL        │
│ `/api/returns/{id}`        │ GET         │ Retrieve return request details               │ Go (Gin), PostgreSQL        │
└───────────────────────────┴─────────────┴───────────────────────────────────────────────┴─────────────────────────────┘

#### Example CURLs:
```bash
# Initiate Return
curl -X POST -H "Authorization: Bearer test-token" -H "Content-Type: application/json" -d '{"shipment_id": "SHIP-123", "reason": "Damaged"}' http://localhost:8080/returns
```

---

## 📡 Technology Stack Summary

*   **Language**: Go 1.21
*   **API Framework**: Gin Gonic
*   **Database**: PostgreSQL 15 (Containerized)
*   **Message Broker**: Apache Kafka (Confluent Schema Registry ready)
*   **Logging**: Structured JSON with Uber-Zap
*   **Deployment**: Docker & Docker Compose
*   **Cloud Integrations**: AWS S3 (Labels), Resend/SMTP (Notifications), Stripe (Payments)
