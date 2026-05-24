# API Guide

**Base URL**: `http://localhost:8080` (or your API Gateway host)

**Authentication**: All requests require authorization header
```
Authorization: Bearer <token>
```

## Shipment Management

### Create Shipment
Create a new shipment with sender and receiver details.

**Endpoint**: `POST /shipments`

**Request Body**:
```json
{
  "sender_name": "John Doe",
  "sender_address": "123 Main St, New York, NY 10001",
  "sender_phone": "+1-555-0100",
  "sender_email": "john@example.com",
  "receiver_name": "Jane Smith",
  "receiver_address": "456 Oak Ave, Los Angeles, CA 90001",
  "receiver_phone": "+1-555-0200",
  "receiver_email": "jane@example.com",
  "weight": 2.5,
  "dimensions": "10x10x10",
  "description": "Electronics package",
  "carrier": "dhl",
  "service_type": "express"
}
```

**Response** (201 Created):
```json
{
  "id": "SHIP-001",
  "user_id": "user-123",
  "sender_name": "John Doe",
  "receiver_name": "Jane Smith",
  "status": "pending",
  "weight": 2.5,
  "carrier": "dhl",
  "tracking_number": "1234567890",
  "created_at": "2026-05-24T10:30:00Z",
  "updated_at": "2026-05-24T10:30:00Z"
}
```

**Triggers**:
- Publishes `shipment.created` event to Kafka
- Tracking service creates initial tracking record
- Notification service sends confirmation email
- Billing service creates invoice

---

### Get Shipment
Retrieve a specific shipment by ID.

**Endpoint**: `GET /shipments/{id}`

**Response** (200 OK):
```json
{
  "id": "SHIP-001",
  "user_id": "user-123",
  "sender_name": "John Doe",
  "receiver_name": "Jane Smith",
  "status": "in_transit",
  "weight": 2.5,
  "carrier": "dhl",
  "tracking_number": "1234567890",
  "created_at": "2026-05-24T10:30:00Z",
  "updated_at": "2026-05-24T11:00:00Z"
}
```

---

### List User Shipments
Get all shipments for the authenticated user.

**Endpoint**: `GET /shipments`

**Query Parameters**:
- `status` (optional): Filter by status (pending, in_transit, delivered)
- `carrier` (optional): Filter by carrier (dhl, fedex, ups)
- `limit` (optional, default: 20): Number of results
- `offset` (optional, default: 0): Pagination offset

**Response** (200 OK):
```json
[
  {
    "id": "SHIP-001",
    "status": "in_transit",
    "carrier": "dhl",
    "tracking_number": "1234567890",
    "created_at": "2026-05-24T10:30:00Z"
  },
  {
    "id": "SHIP-002",
    "status": "delivered",
    "carrier": "fedex",
    "tracking_number": "9876543210",
    "created_at": "2026-05-23T14:20:00Z"
  }
]
```

---

### Update Shipment
Modify shipment details (only pending shipments).

**Endpoint**: `PUT /shipments/{id}`

**Request Body**:
```json
{
  "receiver_address": "789 Pine St, Los Angeles, CA 90002"
}
```

**Response** (200 OK):
```json
{
  "id": "SHIP-001",
  "status": "pending",
  "receiver_address": "789 Pine St, Los Angeles, CA 90002",
  "updated_at": "2026-05-24T10:35:00Z"
}
```

---

### Update Shipment Status
Change shipment status (admin/system operation).

**Endpoint**: `PATCH /shipments/{id}/status`

**Request Body**:
```json
{
  "status": "in_transit"
}
```

**Status Values**: `pending`, `in_transit`, `delivered`, `failed`, `cancelled`

**Response** (200 OK):
```json
{
  "message": "status updated",
  "id": "SHIP-001",
  "status": "in_transit"
}
```

**Triggers**:
- Publishes `shipment.status.changed` event
- Notification service sends status update email

---

### Delete Shipment
Delete a pending shipment.

**Endpoint**: `DELETE /shipments/{id}`

**Response** (200 OK):
```json
{
  "message": "shipment deleted",
  "id": "SHIP-001"
}
```

---

## Carrier Integration

### Register Carrier
Register a new carrier with API credentials.

**Endpoint**: `POST /carriers`

**Request Body**:
```json
{
  "name": "DHL Express",
  "code": "dhl",
  "api_key": "your-dhl-api-key",
  "api_secret": "your-dhl-api-secret",
  "base_url": "https://api.dhl.com/v1"
}
```

**Response** (201 Created):
```json
{
  "id": "CARRIER-001",
  "name": "DHL Express",
  "code": "dhl",
  "active": true,
  "created_at": "2026-05-24T10:00:00Z"
}
```

---

### Get Carrier Rates
Fetch real-time rates from all configured carriers.

**Endpoint**: `GET /carriers/rates`

**Query Parameters**:
- `from` (required): Origin address or postal code
- `to` (required): Destination address or postal code
- `weight` (required): Package weight in kg

**Example**: `/carriers/rates?from=New+York&to=Los+Angeles&weight=2.5`

**Response** (200 OK):
```json
{
  "rates": [
    {
      "carrier": "dhl",
      "carrier_name": "DHL Express",
      "service": "express",
      "rate": 45.99,
      "currency": "USD",
      "delivery_days": 2,
      "available": true
    },
    {
      "carrier": "fedex",
      "carrier_name": "FedEx Standard",
      "service": "ground",
      "rate": 32.50,
      "currency": "USD",
      "delivery_days": 5,
      "available": true
    },
    {
      "carrier": "ups",
      "carrier_name": "UPS Express",
      "service": "overnight",
      "rate": 78.99,
      "currency": "USD",
      "delivery_days": 1,
      "available": true
    }
  ]
}
```

---

### Get Tracking Info
Retrieve tracking information from carrier.

**Endpoint**: `GET /carriers/tracking`

**Query Parameters**:
- `carrier` (required): Carrier code (dhl, fedex, ups)
- `tracking_number` (required): Tracking number

**Example**: `/carriers/tracking?carrier=dhl&tracking_number=1234567890`

**Response** (200 OK):
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
    },
    {
      "timestamp": "2026-05-24T18:00:00Z",
      "status": "in_transit",
      "location": "Chicago, IL",
      "description": "Package in transit"
    }
  ]
}
```

---

### Get Pickup Locations
Find pickup points for a carrier near an address.

**Endpoint**: `GET /carriers/pickup-locations`

**Query Parameters**:
- `carrier` (required): Carrier code
- `address` (required): Address/zip code
- `limit` (optional, default: 10): Max results

**Example**: `/carriers/pickup-locations?carrier=dhl&address=New+York&limit=5`

**Response** (200 OK):
```json
{
  "locations": [
    {
      "id": "LOC-001",
      "name": "DHL Service Point - Manhattan",
      "address": "123 Broadway, New York, NY 10001",
      "distance_km": 0.5,
      "phone": "+1-555-1234",
      "hours": "Mon-Fri 9AM-7PM, Sat 10AM-5PM"
    },
    {
      "id": "LOC-002",
      "name": "DHL Service Point - Midtown",
      "address": "456 Park Ave, New York, NY 10022",
      "distance_km": 1.2,
      "phone": "+1-555-5678",
      "hours": "Mon-Fri 8AM-9PM, Sat 10AM-6PM"
    }
  ]
}
```

---

### Get Drop-off Locations
Find drop-off points for a carrier near an address.

**Endpoint**: `GET /carriers/drop-locations`

**Query Parameters**: Same as pickup-locations

**Response**: Same format as pickup-locations

---

## Rate Comparison

### Compare Rates
Compare rates across all carriers for a shipment.

**Endpoint**: `POST /rates/compare`

**Request Body**:
```json
{
  "shipment_id": "SHIP-001",
  "from": "New York, NY 10001",
  "to": "Los Angeles, CA 90001",
  "weight": 2.5
}
```

**Response** (200 OK):
```json
{
  "comparison_id": "COMP-001",
  "shipment_id": "SHIP-001",
  "from": "New York, NY 10001",
  "to": "Los Angeles, CA 90001",
  "weight": 2.5,
  "options": [
    {
      "carrier": "fedex",
      "carrier_name": "FedEx Ground",
      "rate": 32.50,
      "delivery_days": 5,
      "rank": 1
    },
    {
      "carrier": "ups",
      "carrier_name": "UPS Ground",
      "rate": 35.99,
      "delivery_days": 6,
      "rank": 2
    },
    {
      "carrier": "dhl",
      "carrier_name": "DHL Express",
      "rate": 45.99,
      "delivery_days": 2,
      "rank": 3
    }
  ],
  "created_at": "2026-05-24T10:45:00Z"
}
```

**Triggers**:
- Publishes `rates.compared` event to Kafka

---

### Get Rate Comparison
Retrieve a previous rate comparison.

**Endpoint**: `GET /rates/comparison`

**Query Parameters**:
- `shipment_id` (required): Shipment ID

**Response** (200 OK):
```json
{
  "comparison_id": "COMP-001",
  "shipment_id": "SHIP-001",
  "options": [ ... ]
}
```

---

## Label Generation

### Generate Label
Generate shipping label for a shipment.

**Endpoint**: `POST /labels`

**Request Body**:
```json
{
  "shipment_id": "SHIP-001",
  "carrier": "dhl",
  "format": "pdf"
}
```

**Response** (201 Created):
```json
{
  "id": "LABEL-001",
  "shipment_id": "SHIP-001",
  "carrier": "dhl",
  "tracking_number": "1234567890",
  "format": "pdf",
  "file_path": "/labels/LABEL-001.pdf",
  "created_at": "2026-05-24T11:00:00Z"
}
```

**Triggers**:
- Publishes `label.generated` event to Kafka

---

### Get Label
Retrieve label details.

**Endpoint**: `GET /labels/{id}`

**Response** (200 OK):
```json
{
  "id": "LABEL-001",
  "shipment_id": "SHIP-001",
  "carrier": "dhl",
  "tracking_number": "1234567890",
  "status": "ready",
  "created_at": "2026-05-24T11:00:00Z"
}
```

---

### Download Label
Download label PDF file.

**Endpoint**: `GET /labels/{id}/download`

**Response** (200 OK):
```
[PDF Binary Data]
Content-Type: application/pdf
Content-Disposition: attachment; filename="LABEL-001.pdf"
```

---

## Tracking

### Get Tracking History
Get complete tracking history for a shipment.

**Endpoint**: `GET /tracking/{shipment_id}`

**Response** (200 OK):
```json
{
  "shipment_id": "SHIP-001",
  "tracking_number": "1234567890",
  "carrier": "dhl",
  "current_status": "in_transit",
  "events": [
    {
      "id": "EVT-001",
      "timestamp": "2026-05-24T10:30:00Z",
      "status": "created",
      "location": "New York, NY",
      "description": "Shipment created"
    },
    {
      "id": "EVT-002",
      "timestamp": "2026-05-24T15:30:00Z",
      "status": "picked_up",
      "location": "New York, NY",
      "description": "Package picked up from sender"
    },
    {
      "id": "EVT-003",
      "timestamp": "2026-05-24T18:00:00Z",
      "status": "in_transit",
      "location": "Chicago, IL",
      "description": "In transit to destination"
    }
  ]
}
```

---

## Address Validation

### Validate Address
Validate and standardize an address.

**Endpoint**: `POST /addresses/validate`

**Request Body**:
```json
{
  "address": "123 Main St, New York, NY"
}
```

**Response** (200 OK):
```json
{
  "id": "ADDR-001",
  "raw_address": "123 Main St, New York, NY",
  "street": "123 Main Street",
  "city": "New York",
  "state": "NY",
  "postal_code": "10001",
  "country": "USA",
  "latitude": 40.7128,
  "longitude": -74.0060,
  "is_valid": true,
  "validated_at": "2026-05-24T10:50:00Z"
}
```

**Triggers**:
- Publishes `address.validated` event to Kafka

---

### Get Pickup Locations
Find pickup points for address.

**Endpoint**: `GET /addresses/pickup-locations`

**Query Parameters**:
- `address` (required): Address to search from
- `carrier` (required): Carrier code
- `limit` (optional, default: 10): Max results

**Response** (200 OK):
```json
{
  "locations": [
    {
      "id": "LOC-001",
      "name": "DHL Service Point",
      "address": "123 Broadway, New York, NY 10001",
      "city": "New York",
      "latitude": 40.7150,
      "longitude": -74.0070,
      "distance_km": 0.5,
      "type": "pickup"
    }
  ]
}
```

---

### Get Drop-off Locations
Find drop-off points for address.

**Endpoint**: `GET /addresses/drop-locations`

**Query Parameters**: Same as pickup-locations

**Response**: Same format as pickup-locations

---

## Billing

### Create Invoice
Generate invoice for a shipment.

**Endpoint**: `POST /billing/invoices`

**Request Body**:
```json
{
  "shipment_id": "SHIP-001",
  "amount": 45.99,
  "currency": "USD",
  "due_date": "2026-06-24"
}
```

**Response** (201 Created):
```json
{
  "id": "INV-001",
  "shipment_id": "SHIP-001",
  "amount": 45.99,
  "currency": "USD",
  "status": "pending",
  "due_date": "2026-06-24",
  "created_at": "2026-05-24T11:00:00Z"
}
```

---

### Process Payment
Process payment for an invoice.

**Endpoint**: `POST /billing/payments`

**Request Body**:
```json
{
  "invoice_id": "INV-001",
  "method": "stripe"
}
```

**Response** (200 OK):
```json
{
  "id": "PAY-001",
  "invoice_id": "INV-001",
  "amount": 45.99,
  "currency": "USD",
  "status": "completed",
  "method": "stripe",
  "transaction_id": "txn_123456",
  "processed_at": "2026-05-24T11:05:00Z"
}
```

**Triggers**:
- Publishes `payment.processed` event to Kafka
- Notification service sends payment confirmation

---

### Get Invoice
Retrieve invoice details.

**Endpoint**: `GET /billing/invoices/{id}`

**Response** (200 OK):
```json
{
  "id": "INV-001",
  "shipment_id": "SHIP-001",
  "amount": 45.99,
  "status": "paid",
  "created_at": "2026-05-24T11:00:00Z"
}
```

---

### Get Invoice by Shipment
Get invoice for a specific shipment.

**Endpoint**: `GET /billing/invoices`

**Query Parameters**:
- `shipment_id` (required): Shipment ID

**Response** (200 OK):
```json
{
  "id": "INV-001",
  "shipment_id": "SHIP-001",
  "amount": 45.99,
  "status": "paid"
}
```

---

## Returns

### Request Return
Initiate a return for a delivered shipment.

**Endpoint**: `POST /returns`

**Request Body**:
```json
{
  "shipment_id": "SHIP-001",
  "reason": "Product damaged"
}
```

**Response** (201 Created):
```json
{
  "id": "RET-001",
  "shipment_id": "SHIP-001",
  "status": "pending",
  "reason": "Product damaged",
  "created_at": "2026-05-24T12:00:00Z"
}
```

**Triggers**:
- Publishes `return.created` event to Kafka

---

### Approve Return
Approve a return request.

**Endpoint**: `POST /returns/{id}/approve`

**Request Body**:
```json
{
  "carrier": "dhl"
}
```

**Response** (200 OK):
```json
{
  "id": "RET-001",
  "status": "approved",
  "return_tracking_number": "9876543210",
  "approved_at": "2026-05-24T12:30:00Z"
}
```

**Triggers**:
- Publishes `return.status.changed` event
- Generates return label

---

### Process Refund
Process refund for approved return.

**Endpoint**: `POST /returns/{id}/refund`

**Request Body**:
```json
{
  "amount": 45.99
}
```

**Response** (200 OK):
```json
{
  "message": "refund processed",
  "id": "RET-001",
  "refund_amount": 45.99,
  "processed_at": "2026-05-24T13:00:00Z"
}
```

---

### Get Return
Retrieve return details.

**Endpoint**: `GET /returns/{id}`

**Response** (200 OK):
```json
{
  "id": "RET-001",
  "shipment_id": "SHIP-001",
  "status": "approved",
  "reason": "Product damaged",
  "refund_amount": 45.99,
  "created_at": "2026-05-24T12:00:00Z"
}
```

---

### List Returns
Get all returns for a shipment.

**Endpoint**: `GET /returns`

**Query Parameters**:
- `shipment_id` (required): Shipment ID

**Response** (200 OK):
```json
[
  {
    "id": "RET-001",
    "shipment_id": "SHIP-001",
    "status": "approved",
    "created_at": "2026-05-24T12:00:00Z"
  }
]
```

---

## Error Responses

### 400 Bad Request
```json
{
  "error": "Invalid input: weight must be greater than 0"
}
```

### 401 Unauthorized
```json
{
  "error": "missing authorization"
}
```

### 404 Not Found
```json
{
  "error": "shipment not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "internal server error"
}
```

---

## Response Codes

| Code | Meaning |
|------|---------|
| 200 | OK - Success |
| 201 | Created - Resource created |
| 400 | Bad Request - Invalid input |
| 401 | Unauthorized - Auth required |
| 404 | Not Found - Resource not found |
| 500 | Internal Error - Server error |
| 503 | Service Unavailable - Downstream error |
