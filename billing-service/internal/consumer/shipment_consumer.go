package consumer

import (
	"context"
	"encoding/json"

	"github.com/shipping/billing-service/internal/service"
	"github.com/shipping/shared/pkg/kafka"
	"github.com/shipping/shared/pkg/logger"
)

type ShipmentConsumer struct {
	consumer *kafka.Consumer
	svc      *service.BillingService
}

func NewShipmentConsumer(brokers []string, svc *service.BillingService) *ShipmentConsumer {
	sc := &ShipmentConsumer{svc: svc}
	c := kafka.NewConsumer(brokers, kafka.TopicShipmentCreated, "billing-service-group", sc.handle)
	sc.consumer = c
	return sc
}

func (c *ShipmentConsumer) Start(ctx context.Context) {
	logger.Info("starting shipment consumer for billing")
	if err := c.consumer.Start(ctx); err != nil {
		logger.Error("shipment consumer stopped", logger.String("err", err.Error()))
	}
}

func (c *ShipmentConsumer) handle(ctx context.Context, key string, payload json.RawMessage) error {
	var event map[string]interface{}
	if err := json.Unmarshal(payload, &event); err != nil {
		logger.Error("failed to unmarshal shipment event", logger.String("err", err.Error()))
		return nil
	}

	shipmentID, ok := event["shipment_id"].(string)
	if !ok {
		return nil
	}
	userID, _ := event["user_id"].(string)
	weight, _ := event["weight"].(float64)
	carrier, _ := event["carrier"].(string)

	logger.Info("shipment created event received, generating invoice", 
		logger.String("shipment_id", shipmentID),
		logger.Float64("weight", weight))

	// Simple cost calculation: $5 base + $2 per kg
	amount := 5.0 + (weight * 2.0)
	description := "Shipping charges for " + carrier

	_, err := c.svc.CreateInvoice(ctx, shipmentID, userID, amount, description)
	if err != nil {
		logger.Error("failed to create invoice", 
			logger.String("shipment_id", shipmentID), 
			logger.String("err", err.Error()))
		return err
	}

	return nil
}
