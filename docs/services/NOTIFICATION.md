# Notification Service

**Port**: None (Background worker)  
**Role**: Email/SMS Notifications  
**Kafka**: Consumer (7 topics)

## Overview

The Notification Service is a background worker that consumes Kafka events and sends notifications via email and SMS to users and admins.

## Responsibilities

1. **Email Notifications**
   - Shipment confirmation
   - Status updates
   - Delivery confirmation
   - Invoice notification

2. **SMS Notifications**
   - Quick status updates
   - Delivery alerts
   - Payment confirmations

3. **Notification Routing**
   - Send to customer email/phone
   - Send admin alerts
   - Handle opt-outs

## Subscribed Topics

### 1. shipment.created

**Event**:
```json
{
  "shipment_id": "SHIP-001",
  "user_id": "user-123",
  "tracking_number": "1234567890"
}
```

**Action**:
- Send confirmation email to user
- Include tracking number
- Include estimated delivery date

**Email Template**:
```
Subject: Your Shipment is Confirmed!

Hi {user_name},

Your shipment has been confirmed:
- Shipment ID: {shipment_id}
- Tracking Number: {tracking_number}
- Estimated Delivery: {estimated_delivery}

Track your package: {tracking_link}

Thank you for using our service!
```

### 2. shipment.updated

**Action**:
- Send email with updated details
- Include changes made

### 3. shipment.status.changed

**Action**:
- Send status update email
- Send SMS alert for urgent statuses

**Status→Action Mapping**:
```
pending → No notification
in_transit → Email + SMS
out_for_delivery → SMS (high priority)
delivered → Email + SMS
failed → Email (alert)
cancelled → Email
```

### 4. label.generated

**Action**:
- Send shipping label email
- Include label PDF attachment

### 5. tracking.updated

**Action**:
- Send tracking update email
- Include new location
- Only for significant updates (not every event)

### 6. payment.processed

**Action**:
- Send payment confirmation email
- Include receipt
- Include invoice

### 7. invoice.generated

**Action**:
- Send invoice email
- Include PDF invoice

### 8. return.created

**Action**:
- Send return confirmation
- Include return tracking

### 9. return.status.changed

**Action**:
- Send status update
- Include refund status

## Architecture

```
Kafka Topics
    ↓
Consumer Group (notification-service)
    ├─ shipment.created topic
    ├─ shipment.updated topic
    ├─ shipment.status.changed topic
    ├─ label.generated topic
    ├─ tracking.updated topic
    ├─ payment.processed topic
    ├─ invoice.generated topic
    ├─ return.created topic
    └─ return.status.changed topic
    ↓
Process Event
    ├─ Determine notification type
    ├─ Get user preferences
    ├─ Render template
    └─ Send notification
        ├─ Email (SMTP)
        └─ SMS (Twilio)
```

## Data Model

### Notification Entity

```go
type Notification struct {
    ID        string
    UserID    string
    Type      string    // email, sms
    Channel   string    // shipment, billing, tracking
    Subject   string
    Body      string
    Recipient string    // email or phone
    Status    string    // pending, sent, failed
    Attempts  int
    LastError string
    CreatedAt time.Time
    SentAt    time.Time
}

type UserPreference struct {
    UserID          string
    EmailNotifications bool
    SMSNotifications   bool
    OptOutCategories   []string  // ["marketing", "promotional"]
}
```

## Email Templates

### 1. Shipment Confirmation

```html
<h2>Shipment Confirmed</h2>
<p>Your shipment has been confirmed:</p>
<table>
  <tr><td>Shipment ID:</td><td>{shipment_id}</td></tr>
  <tr><td>Tracking Number:</td><td>{tracking_number}</td></tr>
  <tr><td>Carrier:</td><td>{carrier}</td></tr>
  <tr><td>Weight:</td><td>{weight}kg</td></tr>
  <tr><td>Estimated Delivery:</td><td>{estimated_delivery}</td></tr>
</table>
<a href="{tracking_link}">Track Your Package</a>
```

### 2. Status Update

```html
<h2>Your Shipment Status Changed</h2>
<p>Shipment {shipment_id} is now: <strong>{new_status}</strong></p>
<p>Current Location: {location}</p>
<p>Last Update: {timestamp}</p>
<a href="{tracking_link}">View Full Tracking</a>
```

### 3. Delivery Confirmation

```html
<h2>Package Delivered!</h2>
<p>Your shipment {shipment_id} has been delivered!</p>
<p>Delivered at: {timestamp}</p>
<p>Delivered to: {location}</p>
```

## SMS Templates

### Status Update SMS

```
Your package {tracking_number} is {status}. 
Location: {location}. 
Track: {tracking_url}
```

### Delivery SMS

```
Your package has been delivered! 
Shipment: {shipment_id}
Tracking: {tracking_url}
```

## Configuration

**Environment Variables**:
```
KAFKA_BROKERS=kafka:29092
KAFKA_GROUP_ID=notification-service

# Email (SMTP)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=noreply@shipping.com
SMTP_PASS=app-password

# SMS (Twilio)
TWILIO_ACCOUNT_SID=AC...
TWILIO_AUTH_TOKEN=...
TWILIO_PHONE_NUMBER=+1234567890

# Notification Settings
RETRY_ATTEMPTS=3
RETRY_DELAY=300  # 5 minutes
BATCH_SIZE=100
```

## Error Handling

### Email Delivery Failure

```go
// Retry with exponential backoff
retryDelays := []int{60, 300, 900}  // 1m, 5m, 15m

for attempt := 0; attempt < 3; attempt++ {
    err := sendEmail()
    if err == nil {
        return nil
    }
    
    if attempt < 2 {
        time.Sleep(time.Duration(retryDelays[attempt]) * time.Second)
    }
}

// After 3 failures, log and continue
log.Error("Failed to send email after 3 attempts")
```

### SMS Delivery Failure

```go
// Similar retry logic for SMS
// But separate failure tracking (SMS failures are different from email)
```

## Monitoring

### Key Metrics

- Email sent per day
- SMS sent per day
- Delivery success rate
- Average delivery time
- Failed notifications

### Logs

```bash
# View all notifications
docker logs notification-service | grep "sending"

# View email sends
docker logs notification-service | grep "email"

# View SMS sends
docker logs notification-service | grep "sms"

# View failures
docker logs notification-service | grep ERROR
```

## Performance

### Batch Processing

```go
// Process multiple events in batches
const batchSize = 100
var batch []Event

for event := range eventChan {
    batch = append(batch, event)
    
    if len(batch) >= batchSize {
        processBatch(batch)
        batch = []Event{}
    }
}
```

### Rate Limiting

```go
// Limit notifications per user
limiter := rate.NewLimiter(rate.Limit(10), 1)  // 10 per second

if !limiter.Allow() {
    // Queue for later
    queueNotification(notification)
}
```

## Testing

### Mock Email Service

```go
type MockEmailService struct {
    SentEmails []Email
}

func (m *MockEmailService) Send(email Email) error {
    m.SentEmails = append(m.SentEmails, email)
    return nil
}
```

### Testing

```bash
# Test email delivery
docker exec notification-service go test ./...

# Test Kafka consumption
docker logs notification-service | grep "consuming"
```

## Troubleshooting

### Emails Not Sending

```bash
# Check SMTP credentials
docker exec notification-service env | grep SMTP

# Check Kafka connection
docker exec notification-service curl kafka:29092

# View error logs
docker logs notification-service | grep ERROR
```

### SMS Not Sending

```bash
# Verify Twilio credentials
docker exec notification-service env | grep TWILIO

# Check logs for SMS errors
docker logs notification-service | grep "SMS\|sms"
```

## Future Enhancements

1. **In-App Notifications**: Push notifications to mobile app
2. **Notification Center**: User dashboard for notifications
3. **Notification Preferences**: Granular control per event type
4. **Unsubscribe**: Easy opt-out mechanism
5. **Digest Emails**: Daily digest instead of individual notifications
6. **Personalization**: Custom greeting, preferred language
7. **Rich Notifications**: Include images, maps, buttons
