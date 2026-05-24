# Shipment Service

**Port**: 8081  
**Database**: PostgreSQL (shipments)  
**Role**: Shipment Lifecycle Management

## Overview

The Shipment Service manages the entire lifecycle of shipments - from creation through delivery. It's the core service that orchestrates other services.

## Responsibilities

1. **Shipment CRUD Operations**
   - Create new shipments
   - Retrieve shipment details
   - Update shipment information
   - List user shipments
   - Delete shipments (only pending)

2. **Status Management**
   - Track shipment status transitions
   - Publish status change events
   - Trigger downstream services

3. **Event Publishing**
   - Publishes: `shipment.created`, `shipment.updated`, `shipment.status.changed`
   - Enables other services to react
   - Decouples service dependencies

## Architecture

```
HTTP Handler
    ↓
ShipmentService (Business Logic)
    ├─ Validation
    ├─ Orchestration
    └─ Event Publishing
    ↓
ShipmentRepository (Data Access)
    ├─ CRUD operations
    ├─ Query building
    └─ Transaction management
    ↓
PostgreSQL Database
    └─ Shipments Table
```

## Data Model

### Shipment Entity

```go
type Shipment struct {
    ID              string    // Unique ID (SHIP-001, SHIP-002, ...)
    UserID          string    // Owner of shipment
    SenderName      string
    SenderAddress   string
    SenderPhone     string
    SenderEmail     string
    ReceiverName    string
    ReceiverAddress string
    ReceiverPhone   string
    ReceiverEmail   string
    Weight          float64   // in kg
    Dimensions      string    // WxLxH
    Description     string
    Status          string    // pending, in_transit, delivered, failed, cancelled
    Carrier         string    // dhl, fedex, ups, etc.
    TrackingNumber  string
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

### Database Schema

```sql
CREATE TABLE shipments (
    id VARCHAR(50) PRIMARY KEY,
    user_id VARCHAR(50) NOT NULL,
    sender_name VARCHAR(255) NOT NULL,
    sender_address TEXT NOT NULL,
    sender_phone VARCHAR(20),
    sender_email VARCHAR(255),
    receiver_name VARCHAR(255) NOT NULL,
    receiver_address TEXT NOT NULL,
    receiver_phone VARCHAR(20),
    receiver_email VARCHAR(255),
    weight DECIMAL(10, 2) NOT NULL,
    dimensions VARCHAR(50),
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    carrier VARCHAR(50),
    tracking_number VARCHAR(100),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_shipments_user_id ON shipments(user_id);
CREATE INDEX idx_shipments_status ON shipments(status);
CREATE INDEX idx_shipments_created_at ON shipments(created_at DESC);
```

## API Endpoints

### POST /shipments - Create Shipment

**Request**:
```json
{
  "sender_name": "John Doe",
  "sender_address": "123 Main St, New York, NY",
  "receiver_name": "Jane Smith",
  "receiver_address": "456 Oak Ave, Los Angeles, CA",
  "weight": 2.5,
  "carrier": "dhl",
  "service_type": "express"
}
```

**Response**: 201 Created
```json
{
  "id": "SHIP-001",
  "status": "pending",
  "tracking_number": "1234567890",
  "created_at": "2026-05-24T10:30:00Z"
}
```

**Process**:
1. Validate input
2. Generate tracking number
3. Create shipment record
4. Publish `shipment.created` event

### GET /shipments - List User Shipments

**Query Parameters**:
- `status` (optional): Filter by status
- `carrier` (optional): Filter by carrier
- `limit` (optional, default: 20)
- `offset` (optional, default: 0)

### GET /shipments/:id - Get Specific Shipment

### PUT /shipments/:id - Update Shipment

Only pending shipments can be updated.

### PATCH /shipments/:id/status - Update Status

**Status Transitions**:
```
pending → in_transit → out_for_delivery → delivered
  ↓                ↓
cancelled       failed
```

**Publishes**: `shipment.status.changed` event

### DELETE /shipments/:id - Delete Shipment

Only pending shipments can be deleted.

## Workflow

### Create Shipment Workflow

```
1. User submits POST /shipments
   ├─ Validate all required fields
   ├─ Check weight > 0
   └─ Validate email addresses

2. Shipment Service
   ├─ Generate unique ID (SHIP-001)
   ├─ Generate tracking number
   ├─ Create Shipment record in DB
   └─ Publish shipment.created event

3. Event Subscribers React
   ├─ Tracking Service
   │  ├─ Create initial tracking record
   │  └─ Set status "created"
   ├─ Notification Service
   │  └─ Send confirmation email
   └─ Billing Service
      ├─ Calculate charges
      └─ Create invoice

4. Return Response to User
   ├─ Status: 201 Created
   ├─ Include shipment ID
   └─ Include tracking number
```

## Event Publishing

### shipment.created

**When**: New shipment is created  
**Event**:
```json
{
  "event_type": "shipment.created",
  "timestamp": "2026-05-24T10:30:00Z",
  "shipment_id": "SHIP-001",
  "tracking_number": "1234567890",
  "carrier": "dhl"
}
```

### shipment.updated

**When**: Shipment details are modified  
**Event**:
```json
{
  "event_type": "shipment.updated",
  "timestamp": "2026-05-24T10:35:00Z",
  "shipment_id": "SHIP-001",
  "changed_fields": ["receiver_address"]
}
```

### shipment.status.changed

**When**: Shipment status transitions  
**Event**:
```json
{
  "event_type": "shipment.status.changed",
  "timestamp": "2026-05-24T10:45:00Z",
  "shipment_id": "SHIP-001",
  "old_status": "pending",
  "new_status": "in_transit"
}
```

## Inter-Service Communication

### No Direct Calls to Other Services

Shipment Service doesn't directly call other services. Communication is via events:

```
Shipment Service → Kafka Event → Other Services
```

This ensures loose coupling and resilience.

## Configuration

**Environment Variables**:
```
PORT=8081
DB_HOST=postgres-shipment
DB_PORT=5432
DB_USER=postgres
DB_PASS=postgres
DB_NAME=shipments
KAFKA_BROKERS=kafka:29092
```

## Development

### Local Setup

```bash
cd shipment-service

# Install dependencies
go mod download

# Run local database
# Use: docker run -p 5431:5432 postgres:15-alpine

# Set environment
export DB_HOST=localhost DB_PORT=5431

# Apply migrations
psql postgres://postgres:postgres@localhost:5431/shipments \
  -f migrations/001_create_shipments.sql

# Run service
go run ./cmd
```

### Testing

```bash
# Run tests
go test ./...

# With coverage
go test -cover ./...

# Test specific scenario
go test -run TestCreateShipment ./internal/service
```

## Error Handling

### Validation Errors
```json
{
  "error": "invalid weight: must be greater than 0"
}
```

### Database Errors
```json
{
  "error": "database error: could not create shipment"
}
```

### Kafka Errors
```json
{
  "error": "failed to publish event"
}
```

## Performance Considerations

### Database Indexes

Optimize common queries:
- `user_id`: List user shipments
- `status`: Filter by status
- `created_at`: Sort by creation time

### Query Optimization

```sql
-- Good: Use indexes
SELECT * FROM shipments WHERE user_id = 'user-123' ORDER BY created_at DESC;

-- Avoid: Expensive full scan
SELECT * FROM shipments WHERE LOWER(sender_name) = 'john';
```

### Pagination

Always use limit/offset:
```go
query := "SELECT * FROM shipments WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3"
rows, _ := db.Query(query, userID, limit, offset)
```

## Monitoring

### Health Check

```bash
curl -H "Authorization: Bearer test" http://localhost:8081/health
```

### Key Metrics

- Shipments created per day
- Average creation time
- Status distribution
- Database query performance

### Logs

```bash
# View logs
docker logs shipment-service

# Search for errors
docker logs shipment-service | grep ERROR

# Follow logs
docker logs -f shipment-service
```

## Troubleshooting

### Shipment not created

```bash
# Check database connection
docker exec shipment-service curl http://postgres-shipment:5432

# Check logs for validation errors
docker logs shipment-service | grep ERROR

# Verify database has table
docker exec postgres-shipment psql -U postgres -d shipments -c "\dt"
```

### Events not published

```bash
# Check Kafka connection
docker exec shipment-service curl http://kafka:9092

# Check Kafka topics
docker exec kafka kafka-topics --list --bootstrap-server kafka:9092

# Check for errors
docker logs shipment-service | grep kafka
```

## Future Enhancements

1. **Batch shipment creation**: Create multiple shipments at once
2. **Shipment templates**: Save common configurations
3. **Scheduling**: Schedule shipment creation for future date
4. **Analytics**: Shipment trends and patterns
5. **Webhooks**: Notify external systems of status changes
