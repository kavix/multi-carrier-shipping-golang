# Return Service

**Port**: 8088  
**Database**: PostgreSQL (returns)  
**Role**: Return Management and Refunds  
**Kafka**: Producer (return.created, return.status.changed)

## Overview

The Return Service manages product returns, refund processing, and reverse logistics. It handles the complete lifecycle of returns from creation through final disposition.

## Responsibilities

1. **Return Management**
   - Create return requests
   - Track return status
   - Manage return inventory

2. **Refund Processing**
   - Calculate refund amounts
   - Process refunds to original payment method
   - Track refund status

3. **Reverse Logistics**
   - Generate return shipping labels
   - Track return shipments
   - Receive returned items

## API Endpoints

### POST /returns - Create Return Request

**Request**:
```json
{
  "shipment_id": "SHIP-001",
  "user_id": "user-123",
  "reason": "product_defective",
  "description": "Product arrived damaged",
  "return_method": "mail"  // or drop-off
}
```

**Response**: 201 Created
```json
{
  "return_id": "RET-001",
  "shipment_id": "SHIP-001",
  "status": "approved",
  "return_tracking": "RET1234567890",
  "return_label_url": "/labels/RET-001/download",
  "approved_at": "2026-05-24T10:30:00Z"
}
```

**Approval Logic**:
```
1. Check original shipment
2. Verify within return window (30 days)
3. Validate return reason
4. Auto-approve for defects/wrong items
5. Send to manual review for other reasons
```

### GET /returns/:id - Get Return Details

**Response**:
```json
{
  "return_id": "RET-001",
  "shipment_id": "SHIP-001",
  "user_id": "user-123",
  "reason": "product_defective",
  "status": "shipped_back",
  "return_tracking": "RET1234567890",
  "original_amount": 45.99,
  "refund_amount": 45.99,
  "refund_status": "pending",
  "created_at": "2026-05-24T10:30:00Z",
  "returned_at": "2026-05-27T14:00:00Z",
  "timeline": [
    {
      "timestamp": "2026-05-24T10:30:00Z",
      "status": "created",
      "notes": "Return request created"
    },
    {
      "timestamp": "2026-05-24T11:00:00Z",
      "status": "approved",
      "notes": "Return approved automatically"
    }
  ]
}
```

### POST /returns/:id/approve - Approve Return

**Request**:
```json
{
  "notes": "Return approved by admin"
}
```

### POST /returns/:id/refund - Process Refund

**Request**:
```json
{
  "refund_type": "full"  // or partial
}
```

**Response**:
```json
{
  "refund_id": "REF-001",
  "return_id": "RET-001",
  "amount": 45.99,
  "status": "processed",
  "transaction_id": "re_1Iv5BsIl4KpAR1Y3qXL0vQBN",
  "processed_at": "2026-05-27T14:30:00Z"
}
```

### GET /returns - List Returns

**Query Parameters**:
- `status` (optional): Filter by status
- `user_id` (optional): Filter by user
- `limit` (optional): Pagination limit

## Data Model

### Return Entity

```go
type Return struct {
    ID              string
    ShipmentID      string
    UserID          string
    Reason          string      // defective, wrong_item, not_needed, etc.
    Description     string
    Status          string      // created, approved, rejected, received, refunded
    ReturnMethod    string      // mail, drop-off
    ReturnTracking  string
    OriginalAmount  float64
    RefundAmount    float64
    RefundStatus    string      // pending, processed, failed
    CreatedAt       time.Time
    ApprovedAt      time.Time
    ReturnedAt      time.Time
    RefundedAt      time.Time
}

type RefundTransaction struct {
    ID          string
    ReturnID    string
    Amount      float64
    Method      string      // stripe_refund, bank_transfer
    ExternalID  string      // Stripe refund ID
    Status      string
    CreatedAt   time.Time
}

type ReturnTimeline struct {
    ID        string
    ReturnID  string
    Status    string
    Notes     string
    CreatedAt time.Time
}
```

### Database Schema

```sql
CREATE TABLE returns (
    id VARCHAR(50) PRIMARY KEY,
    shipment_id VARCHAR(50) NOT NULL,
    user_id VARCHAR(50) NOT NULL,
    reason VARCHAR(100) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'created',
    return_method VARCHAR(50),
    return_tracking VARCHAR(100),
    original_amount DECIMAL(10, 2),
    refund_amount DECIMAL(10, 2),
    refund_status VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW(),
    approved_at TIMESTAMP,
    returned_at TIMESTAMP,
    refunded_at TIMESTAMP
);

CREATE TABLE refund_transactions (
    id VARCHAR(50) PRIMARY KEY,
    return_id VARCHAR(50) REFERENCES returns(id),
    amount DECIMAL(10, 2),
    method VARCHAR(50),
    external_id VARCHAR(100),
    status VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE return_timeline (
    id VARCHAR(50) PRIMARY KEY,
    return_id VARCHAR(50) REFERENCES returns(id),
    status VARCHAR(50),
    notes TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_returns_user_id ON returns(user_id);
CREATE INDEX idx_returns_status ON returns(status);
CREATE INDEX idx_returns_shipment_id ON returns(shipment_id);
CREATE INDEX idx_refund_return_id ON refund_transactions(return_id);
```

## Return Workflow

### Complete Workflow

```
1. User Initiates Return
   ├─ Submits reason
   ├─ Chooses method (mail/drop-off)
   └─ POST /returns

2. Automatic Approval Check
   ├─ Check return window (30 days)
   ├─ Auto-approve for: defective, wrong_item, damaged
   └─ Flag manual review for: not_needed, changed_mind

3. Return Label Generation
   ├─ Call Label Service
   ├─ Generate return shipping label
   ├─ Email label to user

4. Return Shipment
   ├─ User ships item back
   ├─ Track via return tracking number
   ├─ Publish return.status.changed → in_transit

5. Receipt at Warehouse
   ├─ Scan returned item
   ├─ Verify condition
   ├─ Update status → received
   ├─ Publish return.status.changed

6. Refund Processing
   ├─ Calculate refund (full or partial)
   ├─ Process via original payment method
   ├─ Publish return.status.changed → refunded

7. Notification
   └─ Send email confirmation to user
```

## Return Reasons

### Auto-Approved

```
- product_defective: Product has defects
- wrong_item: Wrong item shipped
- damaged_in_shipping: Item damaged in transit
- missing_items: Items missing from package
```

### Manual Review Required

```
- changed_mind: Customer changed mind
- not_needed: Customer no longer needs
- better_price_found: Found better price elsewhere
- other: Custom reason
```

## Refund Processing

### Stripe Refunds

```go
import "github.com/stripe/stripe-go/v72/refund"

params := &stripe.RefundParams{
    Charge: stripe.String(chargeID),
    Amount: stripe.Int64(int64(amount * 100)),  // in cents
    Reason: stripe.String(stripe.RefundReasonRequestedByCustomer),
}

refund, err := refund.New(params)
```

### Refund Status

```
pending → processed
  └─ If fails: pending → failed
```

## Kafka Events

### return.created

**Event**:
```json
{
  "event_type": "return.created",
  "return_id": "RET-001",
  "shipment_id": "SHIP-001",
  "user_id": "user-123",
  "reason": "product_defective",
  "timestamp": "2026-05-24T10:30:00Z"
}
```

### return.status.changed

**Event**:
```json
{
  "event_type": "return.status.changed",
  "return_id": "RET-001",
  "old_status": "created",
  "new_status": "approved",
  "timestamp": "2026-05-24T11:00:00Z"
}
```

### Example: Refund Status

```json
{
  "event_type": "return.status.changed",
  "return_id": "RET-001",
  "old_status": "received",
  "new_status": "refunded",
  "refund_amount": 45.99,
  "timestamp": "2026-05-27T15:00:00Z"
}
```

## Configuration

**Environment Variables**:
```
PORT=8088
RETURN_WINDOW_DAYS=30
STRIPE_SECRET_KEY=sk_live_...

# Return reasons
AUTO_APPROVE_REASONS=defective,wrong_item,damaged,missing

# Refund settings
REFUND_METHOD=stripe  # or bank_transfer
```

## Performance

### Batch Refunds

```go
// Process multiple refunds in batch
func ProcessBatchRefunds(returns []Return) {
    for _, ret := range returns {
        go processRefund(ret)
    }
}
```

## Monitoring

### Key Metrics

- Returns created per day
- Return approval rate
- Refund success rate
- Average refund time
- Return reasons distribution

### Logs

```bash
# View return creations
docker logs return-service | grep "return.created"

# View refund processing
docker logs return-service | grep "refund"

# View errors
docker logs return-service | grep ERROR
```

## Troubleshooting

### Refund Not Processing

```bash
# Check Stripe connectivity
docker exec return-service curl -H "Authorization: Bearer sk_..." https://api.stripe.com

# Check Stripe keys
docker exec return-service env | grep STRIPE

# View error details
docker logs return-service | grep "refund failed"
```

### Return Label Not Generated

```bash
# Verify Label Service connectivity
docker exec return-service curl http://label-service:8084/health

# Check logs
docker logs return-service | grep "label"
```

## Future Enhancements

1. **Partial Returns**: Return subset of items from order
2. **Exchanges**: Direct item exchanges without refund
3. **Return Condition**: Categorize returned items (new, used, damaged)
4. **Restocking Fees**: Apply fees for certain conditions
5. **RMA Management**: Generate RMA numbers for tracking
6. **Refurbishment Workflow**: Route returned items for refurbishment
7. **Insurance Claims**: Integration with shipping insurance
8. **Analytics**: Return rate trends and analysis
