package consumer

import (
	"context"
	"encoding/json"

	"github.com/shipping/shared/pkg/kafka"
	"github.com/shipping/shared/pkg/logger"
	"github.com/shipping/shipment-service/internal/repository"
)

type InvoiceConsumer struct {
	consumer *kafka.Consumer
	repo     *repository.ShipmentRepo
}

func NewInvoiceConsumer(brokers []string, repo *repository.ShipmentRepo) *InvoiceConsumer {
	ic := &InvoiceConsumer{repo: repo}
	c := kafka.NewConsumer(brokers, kafka.TopicInvoiceGenerated, "shipment-invoice-group", ic.handle)
	ic.consumer = c
	return ic
}

func (c *InvoiceConsumer) Start(ctx context.Context) {
	logger.Info("starting invoice consumer for shipment update")
	if err := c.consumer.Start(ctx); err != nil {
		logger.Error("invoice consumer stopped", logger.String("err", err.Error()))
	}
}

func (c *InvoiceConsumer) handle(ctx context.Context, key string, payload json.RawMessage) error {
	var event map[string]interface{}
	if err := json.Unmarshal(payload, &event); err != nil {
		logger.Error("failed to unmarshal invoice event", logger.String("err", err.Error()))
		return nil
	}

	shipmentID, ok := event["shipment_id"].(string)
	if !ok {
		return nil
	}

	amount, _ := event["amount"].(float64)

	logger.Info("invoice generated event received, updating shipment cost",
		logger.String("shipment_id", shipmentID),
		logger.Float64("amount", amount))

	if err := c.repo.UpdateCost(ctx, shipmentID, amount); err != nil {
		logger.Error("failed to update shipment cost",
			logger.String("shipment_id", shipmentID),
			logger.String("err", err.Error()))
		return err
	}

	return nil
}
