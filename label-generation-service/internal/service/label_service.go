package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/shipping/label-generation-service/internal/domain"
	"github.com/shipping/label-generation-service/internal/repository"
	"github.com/shipping/shared/pkg/kafka"
	"github.com/shipping/shared/pkg/logger"
	"github.com/shipping/shared/pkg/utils"
)

type LabelService struct {
	repo     *repository.LabelRepo
	producer *kafka.Producer
}

func NewLabelService(repo *repository.LabelRepo, producer *kafka.Producer) *LabelService {
	return &LabelService{repo: repo, producer: producer}
}

func (s *LabelService) GenerateLabel(ctx context.Context, shipmentID, carrier string) (*domain.ShippingLabel, error) {
	// Call carrier integration service to generate label
	carrierServiceURL := getEnv("CARRIER_SERVICE_URL", "http://carrier-integration-service:8082")

	// In production, this calls the carrier API to generate a real label
	// For demo, we simulate label generation
	trackingNumber := fmt.Sprintf("TRACK-%s-%d", carrier, time.Now().Unix())

	// Simulate PDF label data (base64 encoded dummy PDF)
	dummyPDF := "JVBERi0xLjQKJcOkw7zDtsO8CjIgMCBvYmoKPDwKL0xlbmd0aCAzIDAgUgovRmlsdGVyIC9GbGF0ZURlY29kZQo+PgpzdHJlYW0KeJzLSMxLLUmNzNFLzs8rzi8KSkxPBQBwBwcGCmVuZHN0cmVhbQplbmRvYmoKCjQgMCBvYmoKPDwKL1R5cGUgL1BhZ2UKL1BhcmVudCA1IDAgUgovUmVzb3VyY2VzIDw8Ci9Gb250IDw8Ci9GMSA2IDAgUgo+Pgo+PgovTWVkaWFCb3ggWzAgMCA2MTIgNzkyXQo+PgpzdHJlYW0KeJzLSMxLLUmNzNFLzs8rzi8KSkxPBQBwBwcGCmVuZHN0cmVhbQplbmRvYmoKCjUgMCBvYmoKPDwKL1R5cGUgL1BhZ2VzCi9LaWRzIFs0IDAgUl0KPj4KZW5kb2JqCgo2IDAgb2JqCjw8Ci9UeXBlIC9Gb250Ci9TdWJ0eXBlIC9UeXBlMQovQmFzZUZvbnQgL0hlbHZldGljYQo+PgplbmRvYmoKCnhyZWYKMCA3CjAwMDAwMDAwMDAgNjU1MzUgZiAKMDAwMDAwMDAxMCAwMDAwMCBuIAowMDAwMDAwMDc5IDAwMDAwIG4gCjAwMDAwMDAxMzIgMDAwMDAgbiAKMDAwMDAwMDIwNyAwMDAwMCBuIAowMDAwMDAwMzA4IDAwMDAwIG4gCjAwMDAwMDAzNTkgMDAwMDAgbiAKdHJhaWxlcgo8PAovU2l6ZSA3Ci9Sb290IDUgMCBSCj4+CnN0YXJ0eHJlZgo0MzcKJSVFT0YK"

	label := &domain.ShippingLabel{
		ID:             utils.GenerateID(),
		ShipmentID:     shipmentID,
		Carrier:        carrier,
		TrackingNumber: trackingNumber,
		LabelData:      dummyPDF,
		LabelURL:       fmt.Sprintf("%s/labels/download/%s", carrierServiceURL, shipmentID),
		Format:         "PDF",
		Status:         "generated",
		CreatedAt:      time.Now(),
	}

	if err := s.repo.Create(ctx, label); err != nil {
		return nil, fmt.Errorf("save label: %w", err)
	}

	// Publish event
	event := map[string]interface{}{
		"label_id":        label.ID,
		"shipment_id":     shipmentID,
		"carrier":         carrier,
		"tracking_number": trackingNumber,
		"event_type":      "label.generated",
	}
	if err := s.producer.Publish(ctx, label.ID, event); err != nil {
		logger.Error("failed to publish label.generated", logger.String("err", err.Error()))
	}

	logger.Info("label generated", logger.String("shipment_id", shipmentID), logger.String("carrier", carrier))
	return label, nil
}

func (s *LabelService) GetLabel(ctx context.Context, shipmentID string) (*domain.ShippingLabel, error) {
	return s.repo.GetByShipmentID(ctx, shipmentID)
}

func (s *LabelService) DownloadLabel(ctx context.Context, shipmentID string) ([]byte, error) {
	label, err := s.repo.GetByShipmentID(ctx, shipmentID)
	if err != nil {
		return nil, err
	}
	data, err := base64.StdEncoding.DecodeString(label.LabelData)
	if err != nil {
		return nil, fmt.Errorf("decode label: %w", err)
	}
	return data, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
