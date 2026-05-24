# Carrier Integration Service

**Port**: 8082  
**Database**: PostgreSQL (carriers)  
**Role**: Multi-Carrier API Integration  
**Direct Calls**: Called by Rate, Label, Tracking services

## Overview

The Carrier Integration Service manages connections to external carrier APIs (DHL, FedEx, UPS). It provides a unified interface for interacting with multiple carriers.

## Responsibilities

1. **Carrier Management**
   - Register carrier credentials
   - Manage API keys
   - Switch between carriers

2. **Rate Retrieval**
   - Get real-time rates from carriers
   - Cache rates for performance
   - Handle carrier API errors

3. **Tracking Integration**
   - Fetch tracking updates from carriers
   - Normalize tracking data
   - Handle multiple tracking formats

4. **Location Services**
   - Find pickup points
   - Find drop-off points
   - Calculate distance

## Architecture

### CarrierClient Interface

```go
type CarrierClient interface {
    GetRates(from, to string, weight float64) ([]CarrierRate, error)
    GetTracking(trackingNumber string) (*TrackingInfo, error)
    GetPickupLocations(address string, limit int) ([]Location, error)
    GetDropLocations(address string, limit int) ([]Location, error)
}
```

### Implementations

```
CarrierClient Interface
    ├─ DHLClient
    ├─ FedExClient
    ├─ UPSClient
    └─ CustomCarrierClient
```

## Carrier Support

### DHL
- **Endpoint**: https://api.dhl.com/v1
- **Methods**:
  - `/rates` - Get shipping rates
  - `/tracking` - Track shipments
  - `/locations` - Find service points

### FedEx
- **Endpoint**: https://apis.fedex.com/v1
- **Methods**:
  - `/rates/quotes` - Get rates
  - `/track/shipments` - Track shipments
  - `/location/find-locations` - Find locations

### UPS
- **Endpoint**: https://onlinetools.ups.com/rest/v1
- **Methods**
  - `/rates` - Get rates
  - `/track` - Track shipments
  - `/locations` - Find service centers

## API Endpoints

### POST /carriers - Register Carrier

**Request**:
```json
{
  "name": "DHL Express",
  "code": "dhl",
  "api_key": "your-dhl-key",
  "api_secret": "your-dhl-secret",
  "base_url": "https://api.dhl.com/v1"
}
```

### GET /carriers/rates - Get Rates

**Query Parameters**:
- `from` (required): Origin address
- `to` (required): Destination address
- `weight` (required): Package weight in kg

**Response**:
```json
{
  "rates": [
    {
      "carrier": "dhl",
      "carrier_name": "DHL Express",
      "service": "express",
      "rate": 45.99,
      "delivery_days": 2
    }
  ]
}
```

**Process**:
1. Parse request parameters
2. For each registered carrier:
   - Call carrier API with from/to/weight
   - Parse response
   - Convert to standard format
3. Aggregate results
4. Return all rates

### GET /carriers/tracking - Get Tracking

**Query Parameters**:
- `carrier` (required): Carrier code
- `tracking_number` (required): Tracking number

**Response**:
```json
{
  "tracking_number": "1234567890",
  "carrier": "dhl",
  "status": "in_transit",
  "events": [
    {
      "timestamp": "2026-05-24T15:30:00Z",
      "status": "picked_up",
      "location": "New York, NY",
      "description": "Package picked up"
    }
  ]
}
```

### GET /carriers/pickup-locations - Pickup Points

**Query Parameters**:
- `carrier` (required): Carrier code
- `address` (required): Address to search from
- `limit` (optional): Max results

### GET /carriers/drop-locations - Drop-off Points

**Query Parameters**: Same as pickup-locations

## Data Model

### Carrier Entity

```go
type Carrier struct {
    ID        string
    Name      string
    Code      string    // dhl, fedex, ups
    APIKey    string
    APISecret string
    BaseURL   string
    Active    bool
    CreatedAt time.Time
}
```

### CarrierRate

```go
type CarrierRate struct {
    Carrier      string
    CarrierName  string
    Service      string
    Rate         float64
    Currency     string
    DeliveryDays int
    Available    bool
}
```

### Location

```go
type Location struct {
    ID         string
    Name       string
    Address    string
    City       string
    Country    string
    PostalCode string
    Latitude   float64
    Longitude  float64
    Distance   float64  // km
    Type       string   // pickup or drop
    Hours      string
}
```

## Inter-Service Communication

### Called By

```
Rate Service → GET /carriers/rates
Label Service → GET /carriers/rates, GET /carriers/drop-locations
Tracking Service → GET /carriers/tracking
Address Service → GET /carriers/pickup-locations, GET /carriers/drop-locations
```

### Caching Strategy

```
Rate Request
    ↓
Check cache (1 hour TTL)
    ├─ Hit: Return cached rate
    └─ Miss: Call carrier API
            ├─ Parse response
            ├─ Cache result
            └─ Return to caller
```

## Error Handling

### Carrier API Errors

```json
{
  "error": "carrier api error",
  "carrier": "dhl",
  "message": "Invalid API key"
}
```

### Retry Logic

```go
for attempt := 0; attempt < 3; attempt++ {
    response, err := callCarrierAPI()
    if err == nil {
        return response
    }
    if attempt < 2 {
        time.Sleep(time.Duration(math.Pow(2, float64(attempt))) * time.Second)
    }
}
```

### Circuit Breaker

If carrier API fails 5 times in a row:
- Return cached result
- Log warning
- After 30 seconds, try again

## Configuration

**Environment Variables**:
```
DHL_API_KEY=your-dhl-key
DHL_API_SECRET=your-dhl-secret
FEDEX_API_KEY=your-fedex-key
UPS_API_KEY=your-ups-key

# Cache TTL
RATES_CACHE_TTL=3600  # 1 hour
TRACKING_CACHE_TTL=300  # 5 minutes
```

## Development

### Testing Carriers

```bash
# Test DHL rates
curl "http://localhost:8082/carriers/rates?from=New+York&to=Los+Angeles&weight=2.5"

# Test tracking
curl "http://localhost:8082/carriers/tracking?carrier=dhl&tracking_number=1234567890"

# Test locations
curl "http://localhost:8082/carriers/pickup-locations?carrier=dhl&address=New+York"
```

### Mock Carriers

For testing without real API credentials:

```go
type MockCarrier struct {}

func (m *MockCarrier) GetRates(from, to string, weight float64) ([]CarrierRate, error) {
    return []CarrierRate{
        {
            Carrier: "mock",
            Rate: 50.0,
            DeliveryDays: 2,
        },
    }, nil
}
```

## Monitoring

### Key Metrics

- Carrier API response time
- Cache hit rate
- API error rate per carrier
- Rate updates per day

### Logs

```bash
# View carrier API calls
docker logs carrier-service | grep "api call"

# View cache hits/misses
docker logs carrier-service | grep "cache"

# View errors
docker logs carrier-service | grep ERROR
```

## Performance

### Caching

- **Rates**: Cached for 1 hour (relatively stable)
- **Tracking**: Cached for 5 minutes (frequently updated)
- **Locations**: Cached for 1 day (static data)

### Parallel Requests

When multiple rates requested, query all carriers in parallel:

```go
for carrier := range carriers {
    go getRatesFromCarrier(carrier)
}
return aggregateResults()
```

## Future Enhancements

1. **More Carriers**: Add DPD, GLS, TNT, etc.
2. **Real-time Rates**: Direct API calls instead of caching
3. **Rate Negotiation**: Custom rates per customer
4. **Insurance**: Offer shipment insurance
5. **Customs**: Handle international customs
6. **Analytics**: Carrier performance tracking
