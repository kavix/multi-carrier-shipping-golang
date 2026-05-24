package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shipping/tracking-service/internal/service"
	"github.com/shipping/shared/pkg/kafka"
	"github.com/shipping/shared/pkg/logger"
)

type ShipmentCreatedEvent struct {
	ShipmentID string `json:"shipment_id"`
	Carrier    string `json:"carrier"`
	Status     string `json:"status"`
}

type TrackingConsumer struct {
	service *service.TrackingService
}

func NewTrackingConsumer(service *service.TrackingService) *TrackingConsumer {
	return &TrackingConsumer{service: service}
}

func (c *TrackingConsumer) HandleShipmentCreated(ctx context.Context, key string, payload []byte) error {
	var event ShipmentCreatedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	logger.Info("initiating tracking for shipment", logger.String("shipment_id", event.ShipmentID))

	// Create initial tracking event
	_, err := c.service.AddTrackingEvent(ctx, event.ShipmentID, "", event.Carrier, "pending", "", "Shipment registered, awaiting pickup")
	if err != nil {
		return fmt.Errorf("add initial tracking: %w", err)
	}

	return nil
}

func (c *TrackingConsumer) Start(ctx context.Context, brokers []string) error {
	handler := func(ctx context.Context, key string, payload json.RawMessage) error {
		return c.HandleShipmentCreated(ctx, key, payload)
	}

	consumer := kafka.NewConsumer(brokers, kafka.TopicShipmentCreated, "tracking-service-group", handler)
	defer consumer.Close()

	logger.Info("tracking consumer started")
	return consumer.Start(ctx)
}
