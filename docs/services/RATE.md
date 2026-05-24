# Rate Comparison Service

**Port**: 8083  
**Database**: PostgreSQL (rates)  
**Role**: Compare Shipping Rates Across Carriers

## Overview

The Rate Comparison Service compares shipping rates from all available carriers. It provides users with multiple carrier options at different price points for the same shipment.

## Responsibilities

1. **Rate Comparison**
   - Query multiple carriers for rates
   - Normalize pricing
   - Apply discounts/markups

2. **Rate History**
   - Store historical rates
   - Track rate trends
   - Analyze carrier pricing

3. **Recommendation**
   - Suggest best rates
   - Filter by criteria (speed, cost, carrier)

## API Endpoints

### POST /rates/compare - Compare Rates

**Request**:
```json
{
  "from_address": "New York, NY",
  "to_address": "Los Angeles, CA",
  "weight": 2.5,
  "filter_by": "cost" // or "speed"
}
```

**Response**:
```json
{
  "comparison_id": "COMP-001",
  "from": "New York, NY",
  "to": "Los Angeles, CA",
  "weight": 2.5,
  "rates": [
    {
      "carrier": "fedex",
      "service": "standard",
      "cost": 35.99,
      "delivery_days": 5,
      "score": 95
    },
    {
      "carrier": "ups",
      "service": "ground",
      "cost": 38.50,
      "delivery_days": 5,
      "score": 92
    },
    {
      "carrier": "dhl",
      "service": "express",
      "cost": 45.99,
      "delivery_days": 2,
      "score": 88
    }
  ],
  "recommended": {
    "carrier": "fedex",
    "reason": "Best value for cost"
  }
}
```

**Process**:
1. Parse request
2. Call Carrier Service: GET /carriers/rates
3. For each carrier rate:
   - Apply markup (platform fee)
   - Calculate delivery score
   - Apply any promotions
4. Sort by criteria (cost, speed)
5. Return comparison

### GET /rates/history - Rate History

**Query Parameters**:
- `from` (required)
- `to` (required)
- `days` (optional, default: 30)

**Response**:
```json
{
  "from": "New York, NY",
  "to": "Los Angeles, CA",
  "history": [
    {
      "date": "2026-05-24",
      "carrier": "fedex",
      "rate": 35.99,
      "delivery_days": 5
    }
  ],
  "trend": "stable"
}
```

## Data Model

### Rate Entity

```go
type Rate struct {
    ID            string
    ComparisonID  string
    Carrier       string
    Service       string
    BaseRate      float64
    Markup        float64
    FinalRate     float64
    DeliveryDays  int
    IsRecommended bool
    CreatedAt     time.Time
}
```

## Inter-Service Communication

### Calls Carrier Service

```
Rate Service
    ↓
GET /carriers/rates?from=X&to=Y&weight=Z
    ↓
Carrier Service
    ├─ Query DHL API
    ├─ Query FedEx API
    ├─ Query UPS API
    ↓
Return aggregated rates
```

### Event Publishing

**rates.compared**:
```json
{
  "event_type": "rates.compared",
  "comparison_id": "COMP-001",
  "best_rate": 35.99,
  "best_carrier": "fedex",
  "timestamp": "2026-05-24T10:30:00Z"
}
```

## Performance Optimization

### Parallel Rate Queries

```go
// Query all carriers in parallel
var wg sync.WaitGroup
rateChan := make(chan CarrierRate, 10)

for _, carrier := range carriers {
    wg.Add(1)
    go func(c string) {
        defer wg.Done()
        rate := queryCarrier(c)
        rateChan <- rate
    }(carrier)
}

wg.Wait()
close(rateChan)
```

### Caching

- Cache comparison results for 30 minutes
- Allow refresh if user requests

## Configuration

**Environment Variables**:
```
PORT=8083
CARRIER_SERVICE_URL=http://carrier-service:8082
RATE_MARKUP_PERCENTAGE=15  # 15% platform fee
```

## Troubleshooting

### Rates Not Comparing

```bash
# Verify Carrier Service is accessible
docker exec rate-service curl http://carrier-service:8082/health

# Check logs for carrier API errors
docker logs rate-service | grep ERROR
```
