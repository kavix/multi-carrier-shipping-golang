package service

import (
	"context"
	"fmt"
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

func (s *BillingService) ProcessPayment(ctx context.Context, invoiceID, method string) (*domain.Payment, error) {
	invoice, err := s.repo.GetInvoiceByID(ctx, invoiceID)
	if err != nil {
		return nil, err
	}

	if invoice.Status != "pending" {
		return nil, fmt.Errorf("invoice already %s", invoice.Status)
	}

	var stripeID string
	var payErr error

	if s.stripeSecretKey != "" && s.stripeSecretKey != "sk_test_your_key" {
		stripeID, payErr = s.stripeClient.Charge(ctx, invoice.Amount, invoice.Currency)
		if payErr != nil {
			return nil, fmt.Errorf("stripe charge: %w", payErr)
		}
	} else {
		logger.Info("Stripe secret key is placeholder or empty, using simulated payment gateway")
		stripeID = fmt.Sprintf("pi_sim_%s", utils.GenerateID())
	}

	payment := &domain.Payment{
		ID:        utils.GenerateID(),
		InvoiceID: invoiceID,
		Amount:    invoice.Amount,
		Currency:  invoice.Currency,
		Status:    "completed",
		Method:    method,
		StripeID:  stripeID,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreatePayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("record payment: %w", err)
	}

	if err := s.repo.UpdateInvoiceStatus(ctx, invoiceID, "paid"); err != nil {
		return nil, fmt.Errorf("update invoice: %w", err)
	}

	// Publish payment processed event
	event := map[string]interface{}{
		"payment_id":  payment.ID,
		"invoice_id":  invoiceID,
		"shipment_id": invoice.ShipmentID,
		"amount":      payment.Amount,
		"status":      "completed",
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
