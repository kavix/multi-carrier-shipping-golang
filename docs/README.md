# Multi-Carrier Shipping Platform Documentation Index

This documentation provides comprehensive information about the multi-carrier shipping platform's architecture, services, APIs, and deployment procedures.

## 📚 Documentation Structure

### Foundation Documents

1. **[ARCHITECTURE.md](./ARCHITECTURE.md)** - System Design Overview
   - Complete system architecture
   - Service topology and relationships
   - Communication patterns
   - Scalability considerations
   - Monitoring and observability

2. **[API-GUIDE.md](./API-GUIDE.md)** - REST API Reference
   - Complete endpoint documentation
   - Request/response examples
   - Error codes and handling
   - Authentication details
   - Rate limiting information

3. **[KAFKA-FLOWS.md](./KAFKA-FLOWS.md)** - Event-Driven Architecture
   - All 11 Kafka topics documented
   - Complete event structures
   - Event flow examples
   - Consumer/producer relationships
   - End-to-end workflows

4. **[DEPLOYMENT-GUIDE.md](./DEPLOYMENT-GUIDE.md)** - Deployment Instructions
   - Quick start (Docker Compose)
   - Detailed step-by-step setup
   - Staging deployment guide
   - Kubernetes (production) deployment
   - Troubleshooting and rollback

### Service-Specific Documentation

#### Core Services

- **[Shipment Service](./services/SHIPMENT.md)** (Port 8081)
  - Shipment CRUD and lifecycle management
  - Status tracking and transitions
  - Event publishing

- **[Carrier Integration Service](./services/CARRIER.md)** (Port 8082)
  - Multi-carrier API abstraction
  - Real-time rate retrieval
  - Tracking integration
  - Location services

- **[Rate Comparison Service](./services/RATE.md)** (Port 8083)
  - Compare rates across carriers
  - Apply discounts/markups
  - Historical rate tracking

- **[Label Generation Service](./services/LABEL.md)** (Port 8084)
  - PDF label generation
  - Barcode creation
  - Batch label processing

- **[Tracking Service](./services/TRACKING.md)** (Port 8085)
  - Real-time shipment tracking
  - Carrier polling strategy
  - Tracking event management

- **[Address Validation Service](./services/ADDRESS.md)** (Port 8086)
  - Address validation & standardization
  - Geocoding integration
  - Location services

#### Business Services

- **[Billing Service](./services/BILLING.md)** (Port 8087)
  - Invoice generation
  - Payment processing via Stripe
  - Refund handling

- **[Return Service](./services/RETURN.md)** (Port 8088)
  - Return management
  - Refund processing
  - Reverse logistics

#### Infrastructure Services

- **[Notification Service](./services/NOTIFICATION.md)** (Background Worker)
  - Email notifications
  - SMS notifications
  - Event consumer for 7 Kafka topics

- **[API Gateway](./services/GATEWAY.md)** (Port 8080)
  - Central request router
  - Authentication & authorization
  - Service discovery & proxying

## 🗂️ Quick Navigation

### By Role

**For Developers**:
1. Start with [ARCHITECTURE.md](./ARCHITECTURE.md) to understand the system
2. Review [API-GUIDE.md](./API-GUIDE.md) for endpoints you'll use
3. Check specific service docs for implementation details
4. Use [DEPLOYMENT-GUIDE.md](./DEPLOYMENT-GUIDE.md) for local setup

**For DevOps/Operations**:
1. Read [DEPLOYMENT-GUIDE.md](./DEPLOYMENT-GUIDE.md) for deployment
2. Review [ARCHITECTURE.md](./ARCHITECTURE.md) for system overview
3. Check service docs for troubleshooting sections

**For Product/Business**:
1. Start with [ARCHITECTURE.md](./ARCHITECTURE.md) overview
2. Review [API-GUIDE.md](./API-GUIDE.md) for capabilities
3. Check [KAFKA-FLOWS.md](./KAFKA-FLOWS.md) for workflows

### By Task

**Setting up the platform**:
- [DEPLOYMENT-GUIDE.md](./DEPLOYMENT-GUIDE.md) - Quick Start section

**Creating a shipment**:
- [KAFKA-FLOWS.md](./KAFKA-FLOWS.md) - "Complete Customer Journey" section
- [Shipment Service](./services/SHIPMENT.md) - POST /shipments endpoint
- [API-GUIDE.md](./API-GUIDE.md) - Shipment endpoints

**Tracking a shipment**:
- [Tracking Service](./services/TRACKING.md)
- [API-GUIDE.md](./API-GUIDE.md) - GET /tracking endpoints
- [KAFKA-FLOWS.md](./KAFKA-FLOWS.md) - Tracking topic

**Processing returns**:
- [Return Service](./services/RETURN.md)
- [API-GUIDE.md](./API-GUIDE.md) - Return endpoints
- [KAFKA-FLOWS.md](./KAFKA-FLOWS.md) - Return events

**Managing payments**:
- [Billing Service](./services/BILLING.md)
- [API-GUIDE.md](./API-GUIDE.md) - Billing endpoints
- [KAFKA-FLOWS.md](./KAFKA-FLOWS.md) - Payment events

## 📊 System Overview

### 10 Services

```
┌─ API Gateway (Port 8080)
│
├─ Shipment Service (Port 8081)
├─ Carrier Integration Service (Port 8082)
├─ Rate Comparison Service (Port 8083)
├─ Label Generation Service (Port 8084)
├─ Tracking Service (Port 8085)
├─ Address Validation Service (Port 8086)
├─ Billing Service (Port 8087)
├─ Return Service (Port 8088)
└─ Notification Service (Background)
```

### 11 Kafka Topics

- `shipment.created` - New shipment created
- `shipment.updated` - Shipment details changed
- `shipment.status.changed` - Status transition
- `label.generated` - Label created
- `tracking.updated` - Tracking information updated
- `payment.processed` - Payment completed
- `invoice.generated` - Invoice created
- `address.validated` - Address validated
- `return.created` - Return request created
- `return.status.changed` - Return status updated
- `rates.compared` - Rates comparison completed

### 8 Databases (PostgreSQL)

- postgres-shipment (Port 5431)
- postgres-carrier (Port 5432)
- postgres-rate (Port 5433)
- postgres-label (Port 5434)
- postgres-tracking (Port 5435)
- postgres-address (Port 5436)
- postgres-billing (Port 5437)
- postgres-return (Port 5438)

### Infrastructure

- Apache Kafka 7.5.0
- Zookeeper 7.5.0
- All services in Docker containers
- Docker Compose for development
- Kubernetes for production

## 🚀 Getting Started

### 1. Quick Start (5 minutes)

```bash
cd /path/to/multi-carrier-shipping
./start-all.sh
```

See [DEPLOYMENT-GUIDE.md](./DEPLOYMENT-GUIDE.md#quick-start---development) for details.

### 2. Make First Request

```bash
curl -H "Authorization: Bearer test" http://localhost:8080/health
```

### 3. Create a Shipment

See [API-GUIDE.md](./API-GUIDE.md#post-shipments---create-shipment) for request format.

## 🔗 Cross-Reference Guide

### Services by Feature

**Authentication & Routing**:
- API Gateway → Routes to all services

**Shipment Lifecycle**:
- Shipment Service → Creates/updates
- Address Service → Validates addresses
- Rate Service → Gets rates
- Label Service → Generates labels
- Billing Service → Creates invoices

**Tracking & Fulfillment**:
- Tracking Service → Maintains tracking
- Carrier Service → Gets tracking from carriers
- Notification Service → Sends updates

**Returns & Refunds**:
- Return Service → Manages returns
- Billing Service → Processes refunds

### Kafka Event Flow

```
Shipment Service publishes: shipment.created
    ↓
Subscribed by:
  - Tracking Service (create tracking record)
  - Billing Service (create invoice)
  - Notification Service (send confirmation)
    ↓
Each publishes additional events:
  - Tracking Service publishes: tracking.updated
  - Billing Service publishes: invoice.generated, payment.processed
    ↓
Notification Service subscribes to all and sends notifications
```

## 📖 Common Workflows

### Workflow 1: Create & Ship Package

1. User creates shipment (POST /shipments)
2. Shipment Service stores and publishes `shipment.created`
3. Tracking Service creates tracking record
4. Billing Service creates invoice
5. Notification Service sends confirmation
6. User gets rates (POST /rates/compare)
7. User selects carrier and generates label (POST /labels)
8. Shipment dispatched to carrier

**Documentation**: [KAFKA-FLOWS.md](./KAFKA-FLOWS.md#complete-customer-journey)

### Workflow 2: Track Package

1. User checks tracking (GET /tracking/:id)
2. Tracking Service polls carrier API every hour
3. Updates received, publishes `tracking.updated`
4. Notification Service sends status email
5. User sees real-time location and ETA

**Documentation**: [Tracking Service](./services/TRACKING.md)

### Workflow 3: Handle Return

1. User initiates return (POST /returns)
2. Return Service auto-approves eligible returns
3. Generates return label (calls Label Service)
4. User ships item back
5. Warehouse receives, updates status
6. Refund processed (calls Billing Service)
7. Notification Service sends refund confirmation

**Documentation**: [Return Service](./services/RETURN.md)

## 🛠️ Troubleshooting Quick Links

- Service won't start → [DEPLOYMENT-GUIDE.md](./DEPLOYMENT-GUIDE.md#troubleshooting)
- API Gateway issues → [Gateway Service](./services/GATEWAY.md#troubleshooting)
- Database problems → [DEPLOYMENT-GUIDE.md](./DEPLOYMENT-GUIDE.md#troubleshooting)
- Kafka issues → [DEPLOYMENT-GUIDE.md](./DEPLOYMENT-GUIDE.md#kafka-setup)
- Service connectivity → Check [ARCHITECTURE.md](./ARCHITECTURE.md#service-communication)

## 📋 Checklists

### Pre-Deployment

- [ ] Read [DEPLOYMENT-GUIDE.md](./DEPLOYMENT-GUIDE.md) - Environment Setup
- [ ] Install Docker and Docker Compose
- [ ] Clone repository
- [ ] Configure `.env` file
- [ ] Build Docker images
- [ ] Apply database migrations

### Post-Deployment

- [ ] Verify all services running: `docker compose ps`
- [ ] Check health endpoint: `curl http://localhost:8080/health`
- [ ] Review logs: `docker compose logs`
- [ ] Test sample request (create shipment)
- [ ] Monitor Kafka: Check topics auto-created

### Development Setup

- [ ] Install Go 1.21+
- [ ] Install development tools
- [ ] Set up IDE with Go extension
- [ ] Read service-specific docs for development

## 📝 Document Maintenance

- Last Updated: May 24, 2026
- Version: 1.0
- Services: 10
- Documentation Files: 11 (.md files)
- Total Pages: 600+

### How to Update Documentation

1. Edit relevant `.md` file
2. Update this index if adding new documents
3. Follow markdown formatting standards
4. Include code examples for clarity
5. Add troubleshooting sections when applicable

## 🔐 Security Notes

- All services use HTTP internally (Docker network isolation)
- API Gateway enforces authentication (Bearer token)
- Payment processing uses Stripe (PCI compliant)
- Database credentials in environment variables
- HTTPS enforced in production
- See [DEPLOYMENT-GUIDE.md](./DEPLOYMENT-GUIDE.md#security) for details

## 📞 Support & Questions

For questions about:
- **Architecture** → See [ARCHITECTURE.md](./ARCHITECTURE.md)
- **APIs** → See [API-GUIDE.md](./API-GUIDE.md)
- **Events** → See [KAFKA-FLOWS.md](./KAFKA-FLOWS.md)
- **Deployment** → See [DEPLOYMENT-GUIDE.md](./DEPLOYMENT-GUIDE.md)
- **Specific Service** → See [services/](./services/) directory

## 🎯 Next Steps

1. **Understand Architecture**: Read [ARCHITECTURE.md](./ARCHITECTURE.md)
2. **Learn APIs**: Review [API-GUIDE.md](./API-GUIDE.md)
3. **Deploy Platform**: Follow [DEPLOYMENT-GUIDE.md](./DEPLOYMENT-GUIDE.md)
4. **Explore Services**: Check [KAFKA-FLOWS.md](./KAFKA-FLOWS.md) for workflows
5. **Deep Dive**: Read specific service documentation

---

**Happy Shipping! 📦**
