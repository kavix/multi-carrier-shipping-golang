# Service Documentation

This directory contains detailed documentation for each microservice in the platform.

## Services Overview

### Core Services
1. **[Shipment Service](./services/SHIPMENT.md)** - Shipment CRUD and lifecycle management
2. **[Carrier Integration Service](./services/CARRIER.md)** - Multi-carrier API integration
3. **[Rate Comparison Service](./services/RATE.md)** - Compare rates across carriers
4. **[Label Generation Service](./services/LABEL.md)** - Generate shipping labels
5. **[Tracking Service](./services/TRACKING.md)** - Real-time shipment tracking
6. **[Address Validation Service](./services/ADDRESS.md)** - Address validation and geocoding

### Business Services
7. **[Billing Service](./services/BILLING.md)** - Payments and invoicing
8. **[Return Service](./services/RETURN.md)** - Return management and refunds
9. **[Notification Service](./services/NOTIFICATION.md)** - Email/SMS notifications

### Infrastructure
10. **[API Gateway](./services/GATEWAY.md)** - Request routing and authentication

## Quick Links

| Service | Port | Database | Kafka |
|---------|------|----------|-------|
| API Gateway | 8080 | - | - |
| Shipment | 8081 | PostgreSQL | Producer |
| Carrier | 8082 | PostgreSQL | - |
| Rate | 8083 | PostgreSQL | Producer |
| Label | 8084 | PostgreSQL | Producer |
| Tracking | 8085 | PostgreSQL | Consumer/Producer |
| Address | 8086 | PostgreSQL | Producer |
| Billing | 8087 | PostgreSQL | Producer |
| Return | 8088 | PostgreSQL | Producer |
| Notification | - | - | Consumer |

## Service Architecture Pattern

All services follow this architecture:

```
HTTP Request
    ↓
Handler Layer (HTTP parsing, validation)
    ↓
Service Layer (Business logic, orchestration)
    ↓
Repository Layer (Data access, queries)
    ↓
Database Layer (PostgreSQL)
```

## Inter-Service Communication

### Direct HTTP Calls (Synchronous)
- Rate Service → Carrier Service (get rates)
- Label Service → Carrier Service (get carrier details)
- Address Service → Carrier Service (get pickup locations)

### Kafka Events (Asynchronous)
- Shipment Service publishes: `shipment.created`, `shipment.updated`, `shipment.status.changed`
- Tracking Service publishes: `tracking.updated`
- Billing Service publishes: `payment.processed`, `invoice.generated`
- Notification Service consumes all events

## Deployment

Each service is:
- **Containerized**: Docker image with Alpine base
- **Configurable**: Environment variables for configuration
- **Observable**: Structured logging, health endpoints
- **Scalable**: Stateless design for horizontal scaling

### Environment Variables

Every service expects these variables:

```
PORT=8081                          # Service port
DB_HOST=postgres-service           # Database host
DB_PORT=5432                       # Database port
DB_USER=postgres                   # Database user
DB_PASS=postgres                   # Database password
DB_NAME=shipments                  # Database name
KAFKA_BROKERS=kafka:29092          # Kafka brokers

# Service URLs (for inter-service calls)
SHIPMENT_SERVICE_URL=http://shipment-service:8081
CARRIER_SERVICE_URL=http://carrier-integration-service:8082
... (for all services)
```

## Development

### Local Development (without Docker)

```bash
cd shipment-service

# Install dependencies
go mod download

# Run migrations (manually)
# OR copy SQL to local psql

# Start service
PORT=8081 DB_HOST=localhost go run ./cmd
```

### Testing

```bash
cd shipment-service

# Run tests
go test ./...

# With coverage
go test -cover ./...

# Integration tests
go test -tags=integration ./...
```

## Monitoring

Each service exposes:
- **Health Endpoint**: `GET /health` (basic check)
- **Metrics**: (if configured)
- **Logs**: Structured JSON to stdout

### Health Check

```bash
curl -H "Authorization: Bearer test" http://localhost:8081/health
```

## Troubleshooting

### Common Issues

1. **Service won't connect to database**
   - Check DB_HOST, DB_PORT, DB_USER, DB_PASS
   - Verify database container is running
   - Check network connectivity

2. **Service won't connect to another service**
   - Check service URL environment variables
   - Verify target service is running
   - Check Docker network

3. **Kafka events not being processed**
   - Check KAFKA_BROKERS variable
   - Verify Kafka is running
   - Check consumer group lag

### Debug Commands

```bash
# Check service logs
docker logs -f container-name

# Check environment variables
docker exec container-name env | grep -i "service\|db\|kafka"

# Test connectivity
docker exec shipment-service curl http://carrier-service:8082/health

# Check database
docker exec postgres-shipment psql -U postgres -d shipments -c "SELECT COUNT(*) FROM shipments;"
```

## Next Steps

- Read specific service documentation in `services/` folder
- Review [API Guide](../API-GUIDE.md) for endpoints
- Check [Kafka Flows](../KAFKA-FLOWS.md) for event integration
- See [Deployment Guide](../DEPLOYMENT-GUIDE.md) for deployment instructions
