# Kafka Event Flows

## Overview

Apache Kafka enables asynchronous communication between microservices. Services publish events to Kafka topics, and other services consume those events without direct coupling.

## Event Topics

### 1. shipment.created
**Producer**: Shipment Service  
**Consumers**: Tracking Service, Notification Service, Billing Service

**Trigger**: When a new shipment is created via `POST /shipments`

**Event Structure**:
```json
{
  "event_type": "shipment.created",
  "timestamp": "2026-05-24T10:30:00Z",
  "shipment_id": "SHIP-001",
  "user_id": "user-123",
  "sender_name": "John Doe",
  "sender_address": "123 Main St, New York, NY",
  "receiver_name": "Jane Smith",
  "receiver_address": "456 Oak Ave, Los Angeles, CA",
  "weight": 2.5,
  "carrier": "dhl",
  "tracking_number": "1234567890"
}
```

**Consumer Actions**:

```
┌─────────────────────┐
│ shipment.created    │
└────────┬────────────┘
         │
    ┌────┴──────────────┬─────────────────┬──────────────────┐
    │                   │                 │                  │
    ▼                   ▼                 ▼                  ▼
┌─────────────┐  ┌──────────────┐  ┌──────────────┐  ┌─────────────┐
│  Tracking   │  │ Notification │  │   Billing    │  │   Other     │
│  Service    │  │   Service    │  │   Service    │  │  Services   │
├─────────────┤  ├──────────────┤  ├──────────────┤  └─────────────┘
│ - Creates   │  │ - Sends      │  │ - Creates    │
│   initial   │  │   shipment   │  │   invoice    │
│   tracking  │  │   created    │  │ - Calculates │
│   record    │  │   confirmation│  │   charges    │
│ - Sets      │  │   email      │  │ - Sets       │
│   status    │  │              │  │   billing    │
│   "created" │  │              │  │   status     │
└─────────────┘  └──────────────┘  └──────────────┘
```

**Message Flow**:
```
Shipment Service:
  1. Receive POST /shipments request
  2. Validate input
  3. Create shipment in database
  4. Generate tracking number
  5. Publish shipment.created to Kafka
  ✓ Return 201 Created to client

Tracking Service:
  1. Consume shipment.created event
  2. Create TrackingEvent: "created"
  3. Create TrackingHistory record
  4. Store in database

Notification Service:
  1. Consume shipment.created event
  2. Generate email template
  3. Send confirmation email to user
  4. Log notification

Billing Service:
  1. Consume shipment.created event
  2. Calculate charges based on weight/carrier
  3. Create Invoice record
  4. Update billing status
```

---

### 2. shipment.updated
**Producer**: Shipment Service  
**Consumers**: Notification Service

**Trigger**: When shipment details are updated via `PUT /shipments/{id}`

**Event Structure**:
```json
{
  "event_type": "shipment.updated",
  "timestamp": "2026-05-24T10:35:00Z",
  "shipment_id": "SHIP-001",
  "changed_fields": ["receiver_address"],
  "old_receiver_address": "456 Oak Ave, Los Angeles, CA 90001",
  "new_receiver_address": "789 Pine St, Los Angeles, CA 90002"
}
```

**Consumer Actions**:
- **Notification Service**: Send notification to user about shipment modification

---

### 3. shipment.status.changed
**Producer**: Shipment Service  
**Consumers**: Notification Service, Billing Service

**Trigger**: When shipment status is updated via `PATCH /shipments/{id}/status`

**Event Structure**:
```json
{
  "event_type": "shipment.status.changed",
  "timestamp": "2026-05-24T10:45:00Z",
  "shipment_id": "SHIP-001",
  "old_status": "pending",
  "new_status": "in_transit",
  "reason": "picked_up_by_carrier"
}
```

**Consumer Actions**:
```
Notification Service:
  - Send status update email
  - "Your package is now in transit"

Billing Service:
  - If status = "delivered": Mark invoice as delivered
  - If status = "failed": Calculate recovery charges
```

**Status Transitions**:
```
pending → in_transit → out_for_delivery → delivered
   ↓
cancelled (can happen from pending)
   ↓
failed (can happen from in_transit)
```

---

### 4. rates.compared
**Producer**: Rate Comparison Service  
**Consumers**: None (audit/logging)

**Trigger**: When rates are compared via `POST /rates/compare`

**Event Structure**:
```json
{
  "event_type": "rates.compared",
  "timestamp": "2026-05-24T10:50:00Z",
  "comparison_id": "COMP-001",
  "shipment_id": "SHIP-001",
  "from": "New York, NY",
  "to": "Los Angeles, CA",
  "weight": 2.5,
  "options": [
    {
      "carrier": "fedex",
      "rate": 32.50,
      "delivery_days": 5
    },
    {
      "carrier": "dhl",
      "rate": 45.99,
      "delivery_days": 2
    }
  ]
}
```

**Purpose**: Audit trail for rate comparisons and pricing history

---

### 5. label.generated
**Producer**: Label Generation Service  
**Consumers**: Notification Service

**Trigger**: When a label is generated via `POST /labels`

**Event Structure**:
```json
{
  "event_type": "label.generated",
  "timestamp": "2026-05-24T11:00:00Z",
  "label_id": "LABEL-001",
  "shipment_id": "SHIP-001",
  "carrier": "dhl",
  "tracking_number": "1234567890",
  "file_path": "/labels/LABEL-001.pdf"
}
```

**Consumer Actions**:
- **Notification Service**: Send label download link to user

---

### 6. tracking.updated
**Producer**: Tracking Service  
**Consumers**: Notification Service

**Trigger**: When tracking information is updated from carrier

**Event Structure**:
```json
{
  "event_type": "tracking.updated",
  "timestamp": "2026-05-24T18:00:00Z",
  "shipment_id": "SHIP-001",
  "tracking_number": "1234567890",
  "carrier": "dhl",
  "status": "in_transit",
  "location": "Chicago, IL",
  "description": "Package in transit to destination",
  "event_timestamp": "2026-05-24T17:45:00Z"
}
```

**Consumer Actions**:
- **Notification Service**: Send real-time tracking update to user

---

### 7. address.validated
**Producer**: Address Validation Service  
**Consumers**: None (audit/logging)

**Trigger**: When an address is validated via `POST /addresses/validate`

**Event Structure**:
```json
{
  "event_type": "address.validated",
  "timestamp": "2026-05-24T10:50:00Z",
  "raw_address": "123 Main St, New York, NY",
  "validated_address": {
    "street": "123 Main Street",
    "city": "New York",
    "state": "NY",
    "postal_code": "10001",
    "latitude": 40.7128,
    "longitude": -74.0060,
    "is_valid": true
  }
}
```

**Purpose**: Audit trail for address validations and geocoding history

---

### 8. payment.processed
**Producer**: Billing Service  
**Consumers**: Notification Service

**Trigger**: When a payment is processed via `POST /billing/payments`

**Event Structure**:
```json
{
  "event_type": "payment.processed",
  "timestamp": "2026-05-24T11:05:00Z",
  "payment_id": "PAY-001",
  "invoice_id": "INV-001",
  "shipment_id": "SHIP-001",
  "amount": 45.99,
  "currency": "USD",
  "method": "stripe",
  "transaction_id": "txn_123456",
  "status": "completed"
}
```

**Consumer Actions**:
- **Notification Service**: Send payment confirmation email with receipt

---

### 9. invoice.generated
**Producer**: Billing Service  
**Consumers**: None (audit/logging)

**Trigger**: When an invoice is generated

**Event Structure**:
```json
{
  "event_type": "invoice.generated",
  "timestamp": "2026-05-24T11:00:00Z",
  "invoice_id": "INV-001",
  "shipment_id": "SHIP-001",
  "amount": 45.99,
  "currency": "USD",
  "due_date": "2026-06-24",
  "items": [
    {
      "description": "Shipping Service",
      "quantity": 1,
      "unit_price": 45.99,
      "total": 45.99
    }
  ]
}
```

**Purpose**: Audit trail for invoice generation and billing history

---

### 10. return.created
**Producer**: Return Service  
**Consumers**: Notification Service

**Trigger**: When a return is requested via `POST /returns`

**Event Structure**:
```json
{
  "event_type": "return.created",
  "timestamp": "2026-05-24T12:00:00Z",
  "return_id": "RET-001",
  "shipment_id": "SHIP-001",
  "user_id": "user-123",
  "reason": "Product damaged",
  "requested_amount": 45.99
}
```

**Consumer Actions**:
- **Notification Service**: Send return confirmation to user

---

### 11. return.status.changed
**Producer**: Return Service  
**Consumers**: Notification Service

**Trigger**: When return status changes (`approved`, `rejected`, `completed`)

**Event Structure**:
```json
{
  "event_type": "return.status.changed",
  "timestamp": "2026-05-24T12:30:00Z",
  "return_id": "RET-001",
  "shipment_id": "SHIP-001",
  "old_status": "pending",
  "new_status": "approved",
  "return_tracking_number": "9876543210",
  "return_carrier": "dhl"
}
```

**Consumer Actions**:
- **Notification Service**: Send status update with return label/shipping instructions

---

## Complete User Journey (Event Flow)

### Scenario: Customer Creates Shipment, Receives Updates, Returns Item

```
┌─────────────────────────────────────────────────────────────────┐
│ STEP 1: Customer Creates Shipment                               │
└─────────────────────────────────────────────────────────────────┘
       │
       ▼
   POST /shipments (API Gateway)
       │
       ▼
   Shipment Service
       ├─ Validate input
       ├─ Create shipment in DB
       ├─ Generate tracking number
       └─ Publish: shipment.created ──────┐
                                           │
       ┌───────────────────────────────────┼────────────────────┐
       │                                   │                    │
       ▼                                   ▼                    ▼
   Tracking Service             Notification Service      Billing Service
   ├─ Create tracking record    ├─ Send email             ├─ Calculate charges
   ├─ Set status: "created"     └─ Confirm shipment       ├─ Create invoice
   └─ Store in DB                                         └─ Store in DB

┌─────────────────────────────────────────────────────────────────┐
│ STEP 2: Carrier Picks Up Package                                 │
└─────────────────────────────────────────────────────────────────┘
   External System (Carrier API) → Updates tracking
       │
       ▼
   Tracking Service polls carrier for updates
       ├─ Receives pickup confirmation
       ├─ Add tracking event
       └─ Publish: tracking.updated
                        │
                        ▼
                Notification Service
                ├─ Send email: "Package picked up"
                └─ Send SMS: Status update

┌─────────────────────────────────────────────────────────────────┐
│ STEP 3: Package In Transit                                        │
└─────────────────────────────────────────────────────────────────┘
   Tracking Service (polls carrier periodically)
       │
       ├─ Receives: in_transit, Chicago, IL
       ├─ Publish: tracking.updated
       │
       ▼
   Notification Service
       ├─ Send email: "Package in transit"
       ├─ Send SMS: Real-time update
       └─ Update website tracker

┌─────────────────────────────────────────────────────────────────┐
│ STEP 4: Package Delivered                                         │
└─────────────────────────────────────────────────────────────────┘
   Tracking Service
       ├─ Receives: delivered status
       ├─ Update shipment status in DB
       └─ Publish: shipment.status.changed
                        │
           ┌────────────┼────────────┐
           ▼            ▼            ▼
     Notification    Billing      Other
     Service         Service      Services
     
     Send email:   Mark invoice
     "Delivered"   as complete

┌─────────────────────────────────────────────────────────────────┐
│ STEP 5: Customer Requests Return                                 │
└─────────────────────────────────────────────────────────────────┘
   POST /returns (API Gateway)
       │
       ▼
   Return Service
       ├─ Create return request
       ├─ Set status: "pending"
       └─ Publish: return.created
                        │
                        ▼
                Notification Service
                └─ Send email: "Return received"

┌─────────────────────────────────────────────────────────────────┐
│ STEP 6: Admin Approves Return                                     │
└─────────────────────────────────────────────────────────────────┘
   POST /returns/{id}/approve
       │
       ▼
   Return Service
       ├─ Update status: "approved"
       ├─ Generate return label
       ├─ Get return tracking number
       └─ Publish: return.status.changed
                        │
                        ▼
                Notification Service
                ├─ Send email: "Return approved"
                ├─ Attach return label PDF
                └─ Provide return shipping address

┌─────────────────────────────────────────────────────────────────┐
│ STEP 7: Process Refund                                            │
└─────────────────────────────────────────────────────────────────┘
   POST /returns/{id}/refund
       │
       ▼
   Return Service/Billing Service
       ├─ Process refund via Stripe
       ├─ Update return status: "refunded"
       └─ Publish: payment.processed (refund)
                        │
                        ▼
                Notification Service
                ├─ Send email: "Refund processed"
                ├─ Provide receipt
                └─ Thank you message
```

---

## Kafka Configuration

### Topic Setup

Each topic has partition configuration for parallelism:

```yaml
Topics:
  - shipment.created: 3 partitions (handle burst shipment creation)
  - shipment.updated: 1 partition (order matters)
  - shipment.status.changed: 3 partitions
  - rates.compared: 1 partition (audit trail)
  - label.generated: 1 partition
  - tracking.updated: 3 partitions (high volume)
  - address.validated: 1 partition (audit)
  - payment.processed: 3 partitions
  - invoice.generated: 1 partition (audit)
  - return.created: 2 partitions
  - return.status.changed: 2 partitions
```

### Consumer Groups

```
Consumer Group: notification-service
  - Subscribed Topics:
    - shipment.created
    - shipment.updated
    - shipment.status.changed
    - label.generated
    - tracking.updated
    - payment.processed
    - return.created
    - return.status.changed

Consumer Group: tracking-service
  - Subscribed Topics:
    - shipment.created

Consumer Group: billing-service
  - Subscribed Topics:
    - shipment.created
    - shipment.status.changed
```

---

## Error Handling in Events

### Event Delivery Guarantees

1. **At-Least-Once**: Events delivered at least once (may have duplicates)
   - Consumer should be idempotent
   - Use event ID to deduplicate

2. **Dead Letter Topic**: Failed events moved to DLQ
   ```
   original-topic → Consumer processes → Error
                                         ├─ Retry 3 times
                                         └─ Move to: original-topic.dlq
   ```

3. **Retry Logic**
   ```go
   for attempt := 0; attempt < 3; attempt++ {
     err := processEvent(event)
     if err == nil {
       return nil
     }
     time.Sleep(time.Duration(math.Pow(2, float64(attempt))) * time.Second)
   }
   // On final failure: send to DLQ
   ```

---

## Monitoring Kafka

### Health Checks

```
Kafka Broker: http://kafka:9092
Zookeeper: zookeeper:2181

Check Topics:
  docker exec kafka kafka-topics --list --bootstrap-server kafka:9092

Check Consumer Groups:
  docker exec kafka kafka-consumer-groups --list --bootstrap-server kafka:9092

Check Consumer Lag:
  docker exec kafka kafka-consumer-groups --describe \
    --group notification-service \
    --bootstrap-server kafka:9092
```

---

## Best Practices

1. **Event Structure**: Always include timestamp, ID, version
2. **Idempotency**: Consumer should handle duplicate events
3. **Event Size**: Keep events small (< 1MB typically)
4. **Versioning**: Version events for schema evolution
5. **Monitoring**: Track consumer lag and processing time
6. **Ordering**: Use partition key for ordering guarantee
7. **Dead Letters**: Monitor and replay failed events
