# Address Validation Service

**Port**: 8086  
**Database**: PostgreSQL (addresses)  
**Role**: Address Validation, Standardization, Geocoding  
**Kafka**: Producer (address.validated)

## Overview

The Address Validation Service validates, standardizes, and geocodes addresses. It ensures accurate delivery by verifying recipient addresses before shipment creation.

## Responsibilities

1. **Address Validation**
   - Validate address format
   - Check address existence
   - Detect delivery issues

2. **Address Standardization**
   - Normalize address format
   - Expand abbreviations (St → Street)
   - Correct formatting

3. **Geocoding**
   - Convert addresses to coordinates
   - Find service points
   - Calculate distances

## API Endpoints

### POST /addresses/validate - Validate Address

**Request**:
```json
{
  "street": "123 Main St",
  "city": "New York",
  "state": "NY",
  "postal_code": "10001",
  "country": "USA"
}
```

**Response**:
```json
{
  "valid": true,
  "standardized_address": {
    "street": "123 MAIN STREET",
    "city": "NEW YORK",
    "state": "NY",
    "postal_code": "10001-1234",
    "country": "USA"
  },
  "geocode": {
    "latitude": 40.7128,
    "longitude": -74.0060
  },
  "delivery_point_valid": true,
  "warnings": []
}
```

**Validation Checks**:
```
1. Format validation (required fields)
2. State/country code validation
3. Postal code format
4. Address database lookup
5. Geocoding success
6. Deliverability check
```

### GET /addresses/pickup-locations - Find Pickup Points

**Query Parameters**:
- `address` (required): Full address or coordinates
- `carrier` (optional): Specific carrier
- `limit` (optional, default: 5)

**Response**:
```json
{
  "pickup_locations": [
    {
      "id": "LOC-001",
      "name": "UPS Store - Main St",
      "address": "100 Main St, New York, NY",
      "distance": 0.5,  // km
      "hours": "Mon-Sat 9am-6pm",
      "phone": "(212) 555-1234"
    }
  ]
}
```

### GET /addresses/drop-locations - Find Drop-off Points

**Query Parameters**: Same as pickup-locations

**Response**: Same structure as pickup-locations

## Data Model

### Address Entity

```go
type Address struct {
    ID               string
    Street           string
    City             string
    State            string
    PostalCode       string
    Country          string
    StandardizedAddr string  // Normalized version
    Latitude         float64
    Longitude        float64
    Valid            bool
    Deliverable      bool
    LastValidated    time.Time
    CreatedAt        time.Time
}

type Geocode struct {
    AddressID string
    Latitude  float64
    Longitude float64
}

type Location struct {
    ID       string
    Name     string
    Type     string  // pickup, drop, service_center
    Address  string
    Distance float64
    Hours    string
    Carrier  string
}
```

### Database Schema

```sql
CREATE TABLE addresses (
    id VARCHAR(50) PRIMARY KEY,
    street VARCHAR(255) NOT NULL,
    city VARCHAR(255) NOT NULL,
    state VARCHAR(50) NOT NULL,
    postal_code VARCHAR(20) NOT NULL,
    country VARCHAR(100) NOT NULL,
    standardized_addr VARCHAR(500),
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    valid BOOLEAN DEFAULT false,
    deliverable BOOLEAN DEFAULT false,
    last_validated TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE locations (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(255),
    type VARCHAR(50),
    address_id VARCHAR(50) REFERENCES addresses(id),
    distance DECIMAL(10, 2),
    hours VARCHAR(255),
    carrier VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_addresses_postal_code ON addresses(postal_code);
CREATE INDEX idx_addresses_city_state ON addresses(city, state);
CREATE INDEX idx_locations_type ON locations(type);
```

## Validation Process

### Flowchart

```
Input Address
    ↓
1. Format Validation
   ├─ Check required fields
   ├─ Validate formats
   └─ Normalize case/spacing
    ↓
2. Reference Data Check
   ├─ Verify state/country codes
   ├─ Check postal code format
   └─ Look up in address database
    ↓
3. Geocoding
   ├─ Call geocoding service
   ├─ Get coordinates
   ├─ Get standardized address
    ↓
4. Deliverability Check
   ├─ Check if commercial/residential
   ├─ Check access restrictions
   ├─ Flag rural/remote addresses
    ↓
5. Return Result
   ├─ Store validated address
   ├─ Cache result
   └─ Publish address.validated
```

## Geocoding Integration

### APIs Supported

1. **Google Maps Geocoding API**
   ```
   GET https://maps.googleapis.com/maps/api/geocode/json
   ?address={address}&key={API_KEY}
   ```

2. **OpenStreetMap (Free)**
   ```
   GET https://nominatim.openstreetmap.org/search
   ?q={address}&format=json
   ```

3. **USPS Address Validation**
   ```
   POST https://secure.shippingapis.com/ShippingAPI.dll
   ?API=Verify&XML={address_xml}
   ```

### Caching Strategy

```
Geocode Request
    ↓
Check cache (7 days TTL)
    ├─ Hit: Return cached result
    └─ Miss: Call geocoding API
            ├─ Parse response
            ├─ Cache result
            └─ Return
```

## Error Handling

### Invalid Address

```json
{
  "valid": false,
  "error": "address not found",
  "suggestions": [
    "123 MAIN STREET, NEW YORK, NY",
    "123 MAIN STREET APT 1, NEW YORK, NY"
  ]
}
```

### Ambiguous Address

```json
{
  "valid": false,
  "error": "ambiguous address",
  "message": "Multiple matches found",
  "options": [
    {"street": "123 Main St", "city": "New York", "state": "NY"},
    {"street": "123 Main St", "city": "New York", "state": "NY"}
  ]
}
```

## Kafka Events

### address.validated

**Event**:
```json
{
  "event_type": "address.validated",
  "address_id": "ADDR-001",
  "valid": true,
  "standardized_address": "123 MAIN ST, NEW YORK, NY 10001",
  "latitude": 40.7128,
  "longitude": -74.0060,
  "timestamp": "2026-05-24T10:30:00Z"
}
```

## Configuration

**Environment Variables**:
```
PORT=8086
GEOCODING_API=google  # or osm, usps
GOOGLE_MAPS_API_KEY=AIzaSyD...
USPS_USER_ID=your-usps-id
CACHE_TTL=604800  # 7 days
```

## Performance

### Caching

```go
type CacheKey struct {
    Street     string
    City       string
    State      string
    PostalCode string
    Country    string
}

// Look up in cache first
if cached, ok := addressCache[cacheKey]; ok {
    return cached
}
```

### Batch Validation

```go
// Validate multiple addresses in parallel
func ValidateAddresses(addresses []Address) {
    for i := 0; i < len(addresses); i += batchSize {
        batch := addresses[i : i+batchSize]
        validateBatch(batch)  // Parallel processing
    }
}
```

## Monitoring

### Key Metrics

- Validations per day
- Success rate
- Average validation time
- Geocoding accuracy
- Cache hit rate

### Logs

```bash
# View validations
docker logs address-service | grep "validating"

# View geocoding results
docker logs address-service | grep "geocode"

# View errors
docker logs address-service | grep ERROR
```

## Troubleshooting

### Geocoding Failures

```bash
# Verify API key
docker exec address-service env | grep API_KEY

# Check service connectivity
docker exec address-service curl https://maps.googleapis.com

# View errors
docker logs address-service | grep "geocoding"
```

### Cache Issues

```bash
# Clear address cache
docker exec address-service redis-cli FLUSHDB

# Check cache stats
docker exec address-service redis-cli INFO stats
```

## Future Enhancements

1. **Real-time Validation**: Validate during form input
2. **Address Autocomplete**: Suggest addresses as user types
3. **Signature Capture**: Proof of delivery with photos
4. **Timezone Detection**: Get timezone from coordinates
5. **Holiday Calendar**: Know when addresses aren't deliverable
6. **Accessibility Info**: Wheelchair accessibility, building access
