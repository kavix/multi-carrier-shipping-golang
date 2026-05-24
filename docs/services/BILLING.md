# Billing Service

**Port**: 8087  
**Database**: PostgreSQL (billing)  
**Role**: Payment Processing and Invoicing  
**Kafka**: Producer (payment.processed, invoice.generated)

## Overview

The Billing Service handles all payment processing and invoice management. It calculates charges based on shipment weight and carrier, processes payments through Stripe, and generates invoices.

## Responsibilities

1. **Invoice Generation**
   - Create invoices after shipment creation
   - Calculate charges (base + weight surcharge)
   - Apply discounts and promotions

2. **Payment Processing**
   - Process payments via Stripe
   - Validate payment methods
   - Handle payment failures

3. **Invoice Tracking**
   - Store payment history
   - Track invoice status
   - Handle refunds

## API Endpoints

### POST /billing/invoices - Create Invoice

**Request**:
```json
{
  "shipment_id": "SHIP-001",
  "user_id": "user-123",
  "amount": 45.99,
  "carrier": "fedex"
}
```

**Response**:
```json
{
  "invoice_id": "INV-001",
  "shipment_id": "SHIP-001",
  "amount": 45.99,
  "status": "pending",
  "created_at": "2026-05-24T10:30:00Z",
  "due_date": "2026-06-24T10:30:00Z"
}
```

### POST /billing/payments - Process Payment

**Request**:
```json
{
  "invoice_id": "INV-001",
  "payment_method": "stripe",
  "amount": 45.99,
  "currency": "USD",
  "stripe_token": "tok_visa"
}
```

**Response**:
```json
{
  "payment_id": "PAY-001",
  "invoice_id": "INV-001",
  "amount": 45.99,
  "status": "completed",
  "transaction_id": "ch_1Iv5BsIl4KpAR1Y3qXL0vQBN",
  "timestamp": "2026-05-24T10:35:00Z"
}
```

### GET /billing/invoices/:id - Get Invoice

**Response**:
```json
{
  "invoice_id": "INV-001",
  "shipment_id": "SHIP-001",
  "user_id": "user-123",
  "amount": 45.99,
  "status": "paid",
  "created_at": "2026-05-24T10:30:00Z",
  "paid_at": "2026-05-24T10:35:00Z",
  "payment_id": "PAY-001",
  "line_items": [
    {
      "description": "FedEx Ground Shipping",
      "quantity": 1,
      "unit_price": 35.99,
      "total": 35.99
    },
    {
      "description": "Weight Surcharge (2.5kg)",
      "quantity": 1,
      "unit_price": 10.00,
      "total": 10.00
    }
  ]
}
```

### POST /billing/payments/:id/refund - Refund Payment

**Request**:
```json
{
  "reason": "customer_request"
}
```

## Data Model

### Invoice Entity

```go
type Invoice struct {
    ID            string
    ShipmentID    string
    UserID        string
    Amount        float64
    Currency      string
    Status        string    // pending, paid, overdue, cancelled
    PaymentID     string    // Links to payment
    CreatedAt     time.Time
    DueDate       time.Time
    PaidAt        time.Time
}

type Payment struct {
    ID             string
    InvoiceID      string
    Amount         float64
    Currency       string
    Method         string    // stripe, paypal, bank_transfer
    TransactionID  string    // External transaction ID
    Status         string    // completed, pending, failed, refunded
    StripeChargeID string
    CreatedAt      time.Time
}

type LineItem struct {
    ID          string
    InvoiceID   string
    Description string
    Quantity    int
    UnitPrice   float64
    Total       float64
}
```

### Database Schema

```sql
CREATE TABLE invoices (
    id VARCHAR(50) PRIMARY KEY,
    shipment_id VARCHAR(50) NOT NULL,
    user_id VARCHAR(50) NOT NULL,
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    payment_id VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW(),
    due_date TIMESTAMP,
    paid_at TIMESTAMP
);

CREATE TABLE payments (
    id VARCHAR(50) PRIMARY KEY,
    invoice_id VARCHAR(50) REFERENCES invoices(id),
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    method VARCHAR(50),
    transaction_id VARCHAR(100),
    stripe_charge_id VARCHAR(100),
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE line_items (
    id VARCHAR(50) PRIMARY KEY,
    invoice_id VARCHAR(50) REFERENCES invoices(id),
    description VARCHAR(255),
    quantity INT DEFAULT 1,
    unit_price DECIMAL(10, 2),
    total DECIMAL(10, 2)
);

CREATE INDEX idx_invoices_user_id ON invoices(user_id);
CREATE INDEX idx_invoices_status ON invoices(status);
CREATE INDEX idx_payments_invoice_id ON payments(invoice_id);
```

## Pricing Model

### Base Price

```
Base Price = Carrier Rate
  ├─ DHL: $45/kg base
  ├─ FedEx: $35/kg base
  └─ UPS: $40/kg base
```

### Surcharges

```
Weight Surcharge = (Weight - 1) * $2.50
  - Example: 2.5kg = (2.5 - 1) * 2.50 = $3.75

Distance Surcharge = $5 for international

Service Fee = Amount * 0.1 (10%)
```

### Total Calculation

```go
lineItems := []LineItem{
    {Description: "FedEx Ground", UnitPrice: baseRate},
    {Description: "Weight Surcharge", UnitPrice: weightSurcharge},
    {Description: "Distance Surcharge", UnitPrice: distanceSurcharge},
}

subtotal := sum(lineItems)
serviceFee := subtotal * 0.1
tax := (subtotal + serviceFee) * taxRate
totalAmount := subtotal + serviceFee + tax
```

## Payment Integration

### Stripe Integration

```go
import "github.com/stripe/stripe-go/v72"
import "github.com/stripe/stripe-go/v72/charge"

// Create charge
params := &stripe.ChargeParams{
    Amount:      stripe.Int64(int64(amount * 100)),  // Amount in cents
    Currency:    stripe.String(string(stripe.CurrencyUSD)),
    Source:      stripe.String(stripeToken),
}

charge, err := charge.New(params)
if err != nil {
    return err
}

// Store transaction ID
payment.StripeChargeID = charge.ID
payment.TransactionID = charge.ID
```

### Error Handling

```
Payment Failure:
    ├─ Insufficient funds → Retry after 3 days
    ├─ Card expired → Notify user to update card
    ├─ Network error → Retry exponentially
    └─ Fraud detected → Flag for review
```

## Kafka Events

### payment.processed

**Event**:
```json
{
  "event_type": "payment.processed",
  "payment_id": "PAY-001",
  "invoice_id": "INV-001",
  "shipment_id": "SHIP-001",
  "amount": 45.99,
  "status": "completed",
  "timestamp": "2026-05-24T10:35:00Z"
}
```

### invoice.generated

**Event**:
```json
{
  "event_type": "invoice.generated",
  "invoice_id": "INV-001",
  "shipment_id": "SHIP-001",
  "amount": 45.99,
  "timestamp": "2026-05-24T10:30:00Z"
}
```

## Configuration

**Environment Variables**:
```
PORT=8087
STRIPE_SECRET_KEY=sk_live_...
STRIPE_PUBLIC_KEY=pk_live_...

# Pricing
BASE_RATE_MULTIPLIER=1.0
WEIGHT_SURCHARGE_PER_KG=2.50
SERVICE_FEE_PERCENTAGE=10
TAX_RATE=0.08
```

## Monitoring

### Key Metrics

- Invoices created per day
- Payment success rate
- Average payment amount
- Failed payments
- Refund rate

### Logs

```bash
# View payment processing
docker logs billing-service | grep "payment"

# View errors
docker logs billing-service | grep ERROR

# View charges
docker logs billing-service | grep "charge"
```

## Security

### PCI Compliance

- Never log full credit card numbers
- Use Stripe tokens for payments
- HTTPS only for payment endpoints
- Validate amounts server-side

### Fraud Prevention

```go
// Check for suspicious patterns
if amount > 10000 {
    flagForReview()
}

// Limit concurrent charges per user
concurrent := countConcurrentCharges(userID)
if concurrent > 5 {
    reject()
}
```

## Troubleshooting

### Payment Failures

```bash
# Check Stripe connectivity
docker exec billing-service curl -H "Authorization: Bearer sk_test_..." https://api.stripe.com/v1/charges

# Check logs for Stripe errors
docker logs billing-service | grep "Stripe"

# Verify Stripe keys
docker exec billing-service env | grep STRIPE
```

### Invoice Not Generated

```bash
# Check Kafka event received
docker logs billing-service | grep "payment.processed"

# Check database connection
docker exec billing-service curl postgres://localhost:5437
```

## Future Enhancements

1. **Recurring Billing**: Subscription payments
2. **Invoicing**: PDF invoice generation
3. **Accounting**: Integration with QuickBooks/Xero
4. **Analytics**: Revenue tracking and reporting
5. **Discounts**: Coupon and promotion system
6. **Multiple Currencies**: Support for different currencies
7. **Payment Plans**: Installment payments
