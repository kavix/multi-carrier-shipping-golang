package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/shipping/notification-service/internal/service"
	"github.com/shipping/shared/pkg/kafka"
	"github.com/shipping/shared/pkg/logger"
)

type ShipmentEvent struct {
	ShipmentID    string `json:"shipment_id"`
	UserID        string `json:"user_id"`
	Carrier       string `json:"carrier"`
	Status        string `json:"status"`
	Sender        string `json:"sender"`
	Receiver      string `json:"receiver"`
	SenderEmail   string `json:"sender_email"`
	ReceiverEmail string `json:"receiver_email"`
	EventType     string `json:"event_type"`
}

type TrackingEvent struct {
	ShipmentID     string `json:"shipment_id"`
	TrackingNumber string `json:"tracking_number"`
	Carrier        string `json:"carrier"`
	Status         string `json:"status"`
	Location       string `json:"location"`
	EventType      string `json:"event_type"`
}

type PaymentEvent struct {
	PaymentID  string  `json:"payment_id"`
	InvoiceID  string  `json:"invoice_id"`
	ShipmentID string  `json:"shipment_id"`
	Amount     float64 `json:"amount"`
	Status     string  `json:"status"`
	UserID     string  `json:"user_id"`
	EventType  string  `json:"event_type"`
}

type NotificationConsumer struct {
	service *service.NotificationService
}

func NewNotificationConsumer(service *service.NotificationService) *NotificationConsumer {
	return &NotificationConsumer{service: service}
}

func (c *NotificationConsumer) HandleShipmentCreated(ctx context.Context, key string, payload []byte) error {
	var event ShipmentEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("unmarshal shipment event: %w", err)
	}

	logger.Info("notification: shipment created", logger.String("shipment_id", event.ShipmentID))

	// Send shipment confirmation email
	subject := "Shipment Created - " + event.ShipmentID
	body := fmt.Sprintf(`
Hello,

Your shipment has been created successfully!

Shipment ID: %s
Carrier: %s
Status: %s
From: %s
To: %s

You will receive tracking updates as your package moves.

Thank you for using our service!
`, event.ShipmentID, event.Carrier, event.Status, event.Sender, event.Receiver)

	var errs []error
	if event.SenderEmail != "" {
		if err := c.service.SendEmail(event.SenderEmail, subject, body); err != nil {
			errs = append(errs, fmt.Errorf("send created email to sender (%s): %w", event.SenderEmail, err))
		}
	} else {
		if err := c.service.SendEmail(event.UserID+"@example.com", subject, body); err != nil {
			errs = append(errs, fmt.Errorf("send created email to fallback sender: %w", err))
		}
	}

	if event.ReceiverEmail != "" {
		if err := c.service.SendEmail(event.ReceiverEmail, subject, body); err != nil {
			errs = append(errs, fmt.Errorf("send created email to receiver (%s): %w", event.ReceiverEmail, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("created notification errors: %v", errs)
	}
	return nil
}

func (c *NotificationConsumer) HandleShipmentStatusChanged(ctx context.Context, key string, payload []byte) error {
	var event ShipmentEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("unmarshal status event: %w", err)
	}

	logger.Info("notification: status changed",
		logger.String("shipment_id", event.ShipmentID),
		logger.String("status", event.Status))

	subject := fmt.Sprintf("Shipment Update - %s is now %s", event.ShipmentID, event.Status)
	body := fmt.Sprintf(`
Hello,

Your shipment status has been updated!

Shipment ID: %s
New Status: %s
Carrier: %s

Track your shipment at: https://tracking.example.com/%s

Thank you!
`, event.ShipmentID, event.Status, event.Carrier, event.ShipmentID)

	var errs []error

	// Send email for "return" status with special message
	if strings.EqualFold(event.Status, "return") {
		logger.Info("sending return notification email")
		subject = "Shipment Return Initiated - " + event.ShipmentID
		body = fmt.Sprintf(`
Hello,

Your shipment with ID %s has initiated a return.

Details:
Shipment ID: %s
Status: %s
Reason: (if provided by return service)

We will keep you updated on the return process.

Thank you for using our service!
		`, event.ShipmentID, event.ShipmentID, event.Status)
	}

	if event.SenderEmail != "" {
		if err := c.service.SendEmail(event.SenderEmail, subject, body); err != nil {
			errs = append(errs, fmt.Errorf("send status email to sender (%s): %w", event.SenderEmail, err))
		}
	} else {
		if err := c.service.SendEmail(event.UserID+"@example.com", subject, body); err != nil {
			errs = append(errs, fmt.Errorf("send status email to fallback sender: %w", err))
		}
	}

	if event.ReceiverEmail != "" {
		if err := c.service.SendEmail(event.ReceiverEmail, subject, body); err != nil {
			errs = append(errs, fmt.Errorf("send status email to receiver (%s): %w", event.ReceiverEmail, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("status notification errors: %v", errs)
	}
	return nil
}

func (c *NotificationConsumer) HandleTrackingUpdated(ctx context.Context, key string, payload []byte) error {
	var event TrackingEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("unmarshal tracking event: %w", err)
	}

	logger.Info("notification: tracking updated",
		logger.String("shipment_id", event.ShipmentID),
		logger.String("status", event.Status))

	subject := fmt.Sprintf("Tracking Update - %s", event.ShipmentID)
	body := fmt.Sprintf(`
Hello,

Your package is on the move!

Shipment ID: %s
Tracking Number: %s
Carrier: %s
Current Status: %s
Location: %s

Track live: https://tracking.example.com/%s
`, event.ShipmentID, event.TrackingNumber, event.Carrier, event.Status, event.Location, event.ShipmentID)

	return c.service.SendEmail("user@example.com", subject, body)
}

func (c *NotificationConsumer) HandlePaymentProcessed(ctx context.Context, key string, payload []byte) error {
	var event PaymentEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("unmarshal payment event: %w", err)
	}

	logger.Info("notification: payment processed",
		logger.String("payment_id", event.PaymentID),
		logger.String("status", event.Status))

	if event.Status != "completed" {
		logger.Info("payment was not completed, skipping notification email", logger.String("status", event.Status))
		return nil
	}

	subject := "Payment Succeeded - Receipt for Invoice " + event.InvoiceID
	body := fmt.Sprintf(`Hello,

We have successfully processed your payment! Thank you for your business.

Receipt Details:
Payment ID: %s
Invoice ID: %s
Shipment ID: %s
Amount Paid: $%.2f USD
Status: Completed

You can track your shipment status on the dashboard.

Best regards,
Shipping Team
`, event.PaymentID, event.InvoiceID, event.ShipmentID, event.Amount)

	to := event.UserID + "@example.com"
	return c.service.SendEmail(to, subject, body)
}

func (c *NotificationConsumer) Start(ctx context.Context, brokers []string) error {
	// Start multiple consumers in goroutines

	// Consumer 1: shipment.created
	go func() {
		handler := func(ctx context.Context, key string, payload json.RawMessage) error {
			return c.HandleShipmentCreated(ctx, key, payload)
		}
		consumer := kafka.NewConsumer(brokers, kafka.TopicShipmentCreated, "notification-shipment-group", handler)
		defer consumer.Close()
		logger.Info("notification consumer started for shipment.created")
		if err := consumer.Start(ctx); err != nil {
			logger.Error("shipment consumer error", logger.String("err", err.Error()))
		}
	}()

	// Consumer 2: shipment.status.changed
	go func() {
		handler := func(ctx context.Context, key string, payload json.RawMessage) error {
			return c.HandleShipmentStatusChanged(ctx, key, payload)
		}
		consumer := kafka.NewConsumer(brokers, kafka.TopicShipmentStatusChanged, "notification-status-group", handler)
		defer consumer.Close()
		logger.Info("notification consumer started for shipment.status.changed")
		if err := consumer.Start(ctx); err != nil {
			logger.Error("status consumer error", logger.String("err", err.Error()))
		}
	}()

	// Consumer 3: payment.processed
	go func() {
		handler := func(ctx context.Context, key string, payload json.RawMessage) error {
			return c.HandlePaymentProcessed(ctx, key, payload)
		}
		consumer := kafka.NewConsumer(brokers, kafka.TopicPaymentProcessed, "notification-payment-group", handler)
		defer consumer.Close()
		logger.Info("notification consumer started for payment.processed")
		if err := consumer.Start(ctx); err != nil {
			logger.Error("payment consumer error", logger.String("err", err.Error()))
		}
	}()

	// Consumer 4: tracking.updated (blocking - keeps main alive)
	handler := func(ctx context.Context, key string, payload json.RawMessage) error {
		return c.HandleTrackingUpdated(ctx, key, payload)
	}
	consumer := kafka.NewConsumer(brokers, kafka.TopicTrackingUpdated, "notification-tracking-group", handler)
	defer consumer.Close()
	logger.Info("notification consumer started for tracking.updated")
	return consumer.Start(ctx)
}
