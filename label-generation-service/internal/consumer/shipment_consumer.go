package consumer

import (
	"context"
	"encoding/json"

	"github.com/shipping/label-generation-service/internal/service"
	"github.com/shipping/shared/pkg/kafka"
	"github.com/shipping/shared/pkg/logger"
)

type ShipmentConsumer struct {
	consumer *kafka.Consumer
	svc      *service.LabelService
}

func NewShipmentConsumer(brokers []string, svc *service.LabelService) *ShipmentConsumer {
	sc := &ShipmentConsumer{svc: svc}
	c := kafka.NewConsumer(brokers, kafka.TopicShipmentCreated, "label-generation-group", sc.handle)
	sc.consumer = c
	return sc
}

func (c *ShipmentConsumer) Start(ctx context.Context) {
	logger.Info("starting shipment consumer for label generation")
	if err := c.consumer.Start(ctx); err != nil {
		logger.Error("shipment consumer stopped", logger.String("err", err.Error()))
	}
}

func (c *ShipmentConsumer) handle(ctx context.Context, key string, payload json.RawMessage) error {
	var event map[string]interface{}
	if err := json.Unmarshal(payload, &event); err != nil {
		logger.Error("failed to unmarshal shipment event", logger.String("err", err.Error()))
		return nil // Don't retry on malformed JSON
	}

	shipmentID, ok := event["shipment_id"].(string)
	if !ok {
		return nil
	}

	logger.Info("shipment created event received, generating label", logger.String("shipment_id", shipmentID))

	// Automatically generate label
	_, err := c.svc.GenerateLabel(ctx, event)
	if err != nil {
		logger.Error("failed to auto-generate label", logger.String("shipment_id", shipmentID), logger.String("err", err.Error()))
		return err
	}

	return nil
}
