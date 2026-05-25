package consumer

import (
	"context"
	"encoding/json"

	"github.com/shipping/shared/pkg/kafka"
	"github.com/shipping/shared/pkg/logger"
	"github.com/shipping/shipment-service/internal/repository"
)

type LabelConsumer struct {
	consumer *kafka.Consumer
	repo     *repository.ShipmentRepo
}

func NewLabelConsumer(brokers []string, repo *repository.ShipmentRepo) *LabelConsumer {
	lc := &LabelConsumer{repo: repo}
	c := kafka.NewConsumer(brokers, kafka.TopicLabelGenerated, "shipment-label-group", lc.handle)
	lc.consumer = c
	return lc
}

func (c *LabelConsumer) Start(ctx context.Context) {
	logger.Info("starting label consumer for shipment update")
	if err := c.consumer.Start(ctx); err != nil {
		logger.Error("label consumer stopped", logger.String("err", err.Error()))
	}
}

func (c *LabelConsumer) handle(ctx context.Context, key string, payload json.RawMessage) error {
	var event map[string]interface{}
	if err := json.Unmarshal(payload, &event); err != nil {
		logger.Error("failed to unmarshal label event", logger.String("err", err.Error()))
		return nil
	}

	shipmentID, ok := event["shipment_id"].(string)
	if !ok {
		return nil
	}

	labelID, _ := event["label_id"].(string)
	labelURL, _ := event["label_url"].(string)
	trackingNumber, _ := event["tracking_number"].(string)

	logger.Info("label generated event received, updating shipment",
		logger.String("shipment_id", shipmentID),
		logger.String("label_id", labelID),
		logger.String("label_url", labelURL),
		logger.String("tracking_number", trackingNumber))

	if err := c.repo.UpdateLabelInfo(ctx, shipmentID, labelID, labelURL, trackingNumber); err != nil {
		logger.Error("failed to update shipment label info",
			logger.String("shipment_id", shipmentID),
			logger.String("err", err.Error()))
		return err
	}

	return nil
}
