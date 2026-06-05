package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/shipping/billing-service/internal/domain"
	"github.com/shipping/billing-service/internal/repository"
	"github.com/shipping/shared/pkg/kafka"
	"github.com/shipping/shared/pkg/logger"
	"github.com/shipping/shared/pkg/utils"
)

type BillingService struct {
	repo            *repository.BillingRepo
	paymentProducer *kafka.Producer
	invoiceProducer *kafka.Producer
	stripeClient    *StripeClient
	stripeSecretKey string
}

func NewBillingService(repo *repository.BillingRepo, paymentProducer, invoiceProducer *kafka.Producer, stripeSecretKey string) *BillingService {
	return &BillingService{
		repo:            repo,
		paymentProducer: paymentProducer,
		invoiceProducer: invoiceProducer,
		stripeClient:    NewStripeClient(stripeSecretKey),
		stripeSecretKey: stripeSecretKey,
	}
}

func (s *BillingService) CreateInvoice(ctx context.Context, shipmentID, userID string, amount float64, description string) (*domain.Invoice, error) {
	invoice := &domain.Invoice{
		ID:          utils.GenerateID(),
		ShipmentID:  shipmentID,
		UserID:      userID,
		Amount:      amount,
		Currency:    "USD",
		Status:      "pending",
		Description: description,
		CreatedAt:   time.Now(),
	}

	if err := s.repo.CreateInvoice(ctx, invoice); err != nil {
		return nil, fmt.Errorf("create invoice: %w", err)
	}

	// Publish event
	event := map[string]interface{}{
		"invoice_id":  invoice.ID,
		"shipment_id": shipmentID,
		"user_id":     userID,
		"amount":      amount,
		"status":      "pending",
		"event_type":  "invoice.generated",
	}
	if err := s.invoiceProducer.Publish(ctx, invoice.ID, event); err != nil {
		logger.Error("failed to publish invoice.generated", logger.String("err", err.Error()))
	}

	return invoice, nil
}

func (s *BillingService) ProcessPayment(ctx context.Context, invoiceID, method string) (string, string, error) {
	invoice, err := s.repo.GetInvoiceByID(ctx, invoiceID)
	if err != nil {
		return "", "", err
	}

	if invoice.Status != "pending" {
		return "", "", fmt.Errorf("invoice already %s", invoice.Status)
	}

	var sessionID string
	var checkoutURL string
	var payErr error

	if s.stripeSecretKey != "" && s.stripeSecretKey != "sk_test_your_key" {
		sessionID, checkoutURL, payErr = s.stripeClient.CreateCheckoutSession(ctx, invoice.ID, invoice.Amount, invoice.Currency)
		if payErr != nil {
			return "", "", fmt.Errorf("stripe checkout session: %w", payErr)
		}
	} else {
		logger.Info("Stripe secret key is placeholder or empty, using simulated checkout flow")
		sessionID = fmt.Sprintf("cs_sim_%s_%d", invoice.ID, time.Now().Unix())
		checkoutURL = fmt.Sprintf("http://localhost:5173/?session_id=%s&payment_status=success", sessionID)
	}

	return sessionID, checkoutURL, nil
}

func (s *BillingService) ConfirmPayment(ctx context.Context, sessionID string) (*domain.Payment, error) {
	var paymentStatus string
	var invoiceID string
	var err error

	if strings.HasPrefix(sessionID, "cs_sim_") {
		logger.Info("Confirming simulated payment", logger.String("session_id", sessionID))
		// Extract invoice_id from cs_sim_<invoice_id>_<timestamp>
		parts := strings.Split(sessionID, "_")
		if len(parts) >= 3 {
			invoiceID = parts[2]
		} else {
			return nil, fmt.Errorf("invalid simulated session ID format")
		}
		paymentStatus = "paid"
	} else {
		logger.Info("Retrieving checkout session from Stripe", logger.String("session_id", sessionID))
		paymentStatus, invoiceID, err = s.stripeClient.RetrieveCheckoutSession(ctx, sessionID)
		if err != nil {
			return nil, fmt.Errorf("stripe retrieve session: %w", err)
		}
	}

	invoice, err := s.repo.GetInvoiceByID(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("get invoice: %w", err)
	}

	if invoice.Status != "pending" {
		return nil, fmt.Errorf("invoice already %s", invoice.Status)
	}

	status := "completed"
	invoiceStatus := "paid"
	if paymentStatus != "paid" {
		status = "failed"
		invoiceStatus = "failed"
	}

	payment := &domain.Payment{
		ID:        utils.GenerateID(),
		InvoiceID: invoiceID,
		Amount:    invoice.Amount,
		Currency:  invoice.Currency,
		Status:    status,
		Method:    "stripe",
		StripeID:  sessionID,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreatePayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("record payment: %w", err)
	}

	if err := s.repo.UpdateInvoiceStatus(ctx, invoiceID, invoiceStatus, sessionID); err != nil {
		return nil, fmt.Errorf("update invoice: %w", err)
	}

	// Publish payment processed event
	event := map[string]interface{}{
		"payment_id":  payment.ID,
		"invoice_id":  invoiceID,
		"shipment_id": invoice.ShipmentID,
		"amount":      payment.Amount,
		"status":      status,
		"event_type":  "payment.processed",
	}
	if err := s.paymentProducer.Publish(ctx, payment.ID, event); err != nil {
		logger.Error("failed to publish payment.processed", logger.String("err", err.Error()))
	}

	return payment, nil
}

func (s *BillingService) GetInvoice(ctx context.Context, id string) (*domain.Invoice, error) {
	return s.repo.GetInvoiceByID(ctx, id)
}

func (s *BillingService) GetInvoiceByShipment(ctx context.Context, shipmentID string) (*domain.Invoice, error) {
	return s.repo.GetInvoiceByShipmentID(ctx, shipmentID)
}
