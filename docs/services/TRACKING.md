# Tracking Service

**Port**: 8085  
**Database**: PostgreSQL (tracking)  
**Role**: Real-Time Shipment Tracking  
**Kafka**: Consumer (shipment.created), Producer (tracking.updated)

## Overview

The Tracking Service maintains real-time tracking information for shipments. It subscribes to shipment creation events and polls carrier APIs for tracking updates.

## Responsibilities

1. **Tracking Events**
   - Record tracking updates from carriers
   - Maintain tracking history
   - Detect status changes

2. **Tracking Queries**
   - Get current tracking info
   - Get tracking history
   - Calculate ETA

3. **Kafka Event Consumer**
   - Consume: shipment.created
   - Initialize tracking record
   - Publish: tracking.updated

## Architecture

```
Kafka Event (shipment.created)
    ↓
Tracking Service
    ├─ Create tracking record
    ├─ Poll carrier API (every 1 hour)
    ├─ Process tracking events
    └─ Publish tracking.updated
        ↓
    Notification Service subscribes
    Billing Service subscribes
```

## API Endpoints

### GET /tracking/:shipment_id - Get Tracking

**Response**:
```json
{
  "shipment_id": "SHIP-001",
  "tracking_number": "1234567890",
  "carrier": "fedex",
  "status": "in_transit",
  "current_location": "Memphis, TN",
  "estimated_delivery": "2026-05-26T12:00:00Z",
  "last_update": "2026-05-24T15:30:00Z",
  "events": [
    {
      "timestamp": "2026-05-24T15:30:00Z",
      "status": "in_transit",
      "location": "Memphis, TN",
      "description": "Package in transit"
    },
    {
      "timestamp": "2026-05-24T10:15:00Z",
      "status": "picked_up",
      "location": "New York, NY",
      "description": "Package picked up from origin"
    }
  ]
}
```

### GET /tracking/:shipment_id/history - Full History

**Response**:
```json
{
  "shipment_id": "SHIP-001",
  "total_events": 5,
  "events": [
    // All tracking events
  ]
}
```

## Data Model

### Tracking Entity

```go
type Tracking struct {
    ID                string
    ShipmentID        string
    TrackingNumber    string
    Carrier           string
    CurrentStatus     string
    CurrentLocation   string
    EstimatedDelivery time.Time
    LastPolledAt      time.Time
    CreatedAt         time.Time
}

type TrackingEvent struct {
    ID             string
    TrackingID     string
    Timestamp      time.Time
    Status         string
    Location       string
    Description    string
    RawData        string  // Original carrier response
    CreatedAt      time.Time
}
```

### Database Schema

```sql
CREATE TABLE tracking (
    id VARCHAR(50) PRIMARY KEY,
    shipment_id VARCHAR(50) NOT NULL UNIQUE,
    tracking_number VARCHAR(100) NOT NULL,
    carrier VARCHAR(50) NOT NULL,
    current_status VARCHAR(50),
    current_location VARCHAR(255),
    estimated_delivery TIMESTAMP,
    last_polled_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE tracking_events (
    id VARCHAR(50) PRIMARY KEY,
    tracking_id VARCHAR(50) REFERENCES tracking(id),
    timestamp TIMESTAMP NOT NULL,
    status VARCHAR(50),
    location VARCHAR(255),
    description TEXT,
    raw_data JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_tracking_shipment_id ON tracking(shipment_id);
CREATE INDEX idx_tracking_events_tracking_id ON tracking_events(tracking_id);
CREATE INDEX idx_tracking_events_timestamp ON tracking_events(timestamp DESC);
```

## Kafka Integration

### Consumer: shipment.created

**Subscribes to**: `shipment.created` topic

**Process**:
```
1. Receive event: shipment.created
   {
     "shipment_id": "SHIP-001",
     "tracking_number": "1234567890",
     "carrier": "fedex"
   }

2. Create tracking record
   INSERT INTO tracking (...) VALUES (...)

3. Initialize first event
   INSERT INTO tracking_events (...) 
   VALUES (status='created', ...)

4. Start polling schedule
   - Every 1 hour: poll carrier API
   - Update tracking_events with new info
   - Publish tracking.updated if changed
```

### Producer: tracking.updated

**Publishes to**: `tracking.updated` topic

**Event**:
```json
{
  "event_type": "tracking.updated",
  "shipment_id": "SHIP-001",
  "tracking_number": "1234567890",
  "old_status": "pending",
  "new_status": "in_transit",
  "current_location": "Memphis, TN",
  "timestamp": "2026-05-24T15:30:00Z"
}
```

## Polling Strategy

### Background Job

```
Every 1 hour:
    ├─ Query all active shipments
    ├─ For each shipment:
    │  ├─ Call Carrier Service: GET /carriers/tracking
    │  ├─ Parse response
    │  ├─ Compare with last known status
    │  └─ If changed:
    │     ├─ Insert new event
    │     ├─ Update tracking record
    │     └─ Publish tracking.updated event
    └─ Repeat
```

### Smart Polling

```go
// Adjust polling frequency based on status
frequency := map[string]time.Duration{
    "pending": 24 * time.Hour,
    "in_transit": 1 * time.Hour,
    "out_for_delivery": 15 * time.Minute,
    "delivered": 0,  // Don't poll
    "failed": 0,     // Don't poll
}

nextPoll := time.Now().Add(frequency[tracking.CurrentStatus])
```

## Inter-Service Communication

### Calls Carrier Service

```
GET /carriers/tracking?carrier=fedex&tracking_number=1234567890
    ↓
Returns:
{
  "status": "in_transit",
  "location": "Memphis, TN",
  "events": [...]
}
```

## Performance

### Database Optimization

```sql
-- Index for active shipments
CREATE INDEX idx_tracking_current_status 
ON tracking(current_status) 
WHERE current_status NOT IN ('delivered', 'failed');

-- Index for polling queries
CREATE INDEX idx_tracking_last_polled 
ON tracking(last_polled_at);
```

### Batch Processing

```go
// Process tracking updates in batches
const batchSize = 100

trackingRecords := getActiveShipments(batchSize)
for _, tracking := range trackingRecords {
    go updateTracking(tracking)
}
```

## Error Handling

### Carrier API Failures

```go
// If carrier API fails, keep existing tracking
// Try again next hour
if err != nil {
    log.Printf("Failed to get tracking for %s: %v", trackingNumber, err)
    return nil  // Don't update
}
```

### Missing Carrier Data

```go
// If carrier returns no events, check last update
// If older than 24 hours, mark as stale
if len(events) == 0 && time.Since(lastUpdate) > 24*time.Hour {
    // Consider shipment lost
    PublishEvent("tracking.stale", trackingID)
}
```

## Monitoring

### Key Metrics

- Active shipments being tracked
- Average tracking events per shipment
- Status distribution
- Carrier API response time

### Logs

```bash
# View tracking updates
docker logs tracking-service | grep "tracking.updated"

# View polling activity
docker logs tracking-service | grep "polling"

# View errors
docker logs tracking-service | grep ERROR
```

## Configuration

**Environment Variables**:
```
PORT=8085
CARRIER_SERVICE_URL=http://carrier-service:8082
KAFKA_BROKERS=kafka:29092

# Polling configuration
POLLING_INTERVAL=3600          # 1 hour
INITIAL_POLL_INTERVAL=300      # 5 minutes for just-created
FREQUENCY_BY_STATUS=true       # Enable smart polling
```

## Future Enhancements

1. **Real-time Updates**: Webhook from carriers instead of polling
2. **Predictive Tracking**: ML-based ETA prediction
3. **Geolocation**: Map visualization of shipment location
4. **Delays**: Automatically detect and alert on delays
5. **Proof of Delivery**: Capture signature/photo upon delivery
