package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shipping/shared/pkg/kafka"
	"github.com/shipping/shared/pkg/logger"
	"github.com/shipping/shipment-service/internal/repository"
)

type AddressConsumer struct {
	consumer *kafka.Consumer
	repo     *repository.ShipmentRepo
}

func NewAddressConsumer(brokers []string, repo *repository.ShipmentRepo) *AddressConsumer {
	ac := &AddressConsumer{repo: repo}
	c := kafka.NewConsumer(brokers, kafka.TopicShipmentAddressValidated, "shipment-address-group", ac.handle)
	ac.consumer = c
	return ac
}

func (c *AddressConsumer) Start(ctx context.Context) {
	logger.Info("starting address validation consumer for shipment update")
	if err := c.consumer.Start(ctx); err != nil {
		logger.Error("address consumer stopped", logger.String("err", err.Error()))
	}
}

func (c *AddressConsumer) handle(ctx context.Context, key string, payload json.RawMessage) error {
	var event map[string]interface{}
	if err := json.Unmarshal(payload, &event); err != nil {
		logger.Error("failed to unmarshal address event", logger.String("err", err.Error()))
		return nil
	}

	shipmentID, ok := event["shipment_id"].(string)
	if !ok {
		return nil
	}

	logger.Info("address validated event received, updating shipment status", logger.String("shipment_id", shipmentID))

	// Update status to "validated"
	if err := c.repo.UpdateStatus(ctx, shipmentID, "validated"); err != nil {
		logger.Error("failed to update shipment status after validation", logger.String("shipment_id", shipmentID), logger.String("err", err.Error()))
		return err
	}

	// Optionally update standardized addresses if available
	shipment, err := c.repo.GetByID(ctx, shipmentID)
	if err == nil {
		updated := false
		if sv, ok := event["sender_validated"].(map[string]interface{}); ok && sv["is_valid"] == true {
			street, _ := sv["street"].(string)
			city, _ := sv["city"].(string)
			shipment.SenderAddress = fmt.Sprintf("%s, %s", street, city)
			updated = true
		}
		if rv, ok := event["receiver_validated"].(map[string]interface{}); ok && rv["is_valid"] == true {
			street, _ := rv["street"].(string)
			city, _ := rv["city"].(string)
			shipment.ReceiverAddress = fmt.Sprintf("%s, %s", street, city)
			updated = true
		}
		if updated {
			c.repo.Update(ctx, shipment)
		}
	}

	return nil
}
