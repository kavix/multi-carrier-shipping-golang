package consumer

import (
	"context"
	"encoding/json"

	"github.com/shipping/address-validation-service/internal/service"
	"github.com/shipping/shared/pkg/kafka"
	"github.com/shipping/shared/pkg/logger"
)

type ShipmentConsumer struct {
	consumer *kafka.Consumer
	producer *kafka.Producer
	svc      *service.AddressService
}

func NewShipmentConsumer(brokers []string, svc *service.AddressService, producer *kafka.Producer) *ShipmentConsumer {
	sc := &ShipmentConsumer{svc: svc, producer: producer}
	c := kafka.NewConsumer(brokers, kafka.TopicShipmentCreated, "address-validation-group", sc.handle)
	sc.consumer = c
	return sc
}

func (c *ShipmentConsumer) Start(ctx context.Context) {
	logger.Info("starting shipment consumer for address validation")
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

	senderAddr, _ := event["sender"].(string)
	receiverAddr, _ := event["receiver"].(string)

	logger.Info("validating addresses for shipment", logger.String("shipment_id", shipmentID))

	// Validate Sender
	senderValidated, err := c.svc.ValidateAddress(ctx, senderAddr)
	if err != nil {
		logger.Error("sender validation failed", logger.String("shipment_id", shipmentID), logger.String("err", err.Error()))
	}

	// Validate Receiver
	receiverValidated, err := c.svc.ValidateAddress(ctx, receiverAddr)
	if err != nil {
		logger.Error("receiver validation failed", logger.String("shipment_id", shipmentID), logger.String("err", err.Error()))
	}

	// Prepare data for next step
	event["sender_validated"] = senderValidated
	event["receiver_validated"] = receiverValidated
	event["event_type"] = "shipment.address.validated"

	// Publish to TopicShipmentAddressValidated
	if err := c.producer.Publish(ctx, shipmentID, event); err != nil {
		logger.Error("failed to publish shipment.address.validated", logger.String("shipment_id", shipmentID), logger.String("err", err.Error()))
		return err
	}

	logger.Info("addresses validated and event published", logger.String("shipment_id", shipmentID))
	return nil
}
