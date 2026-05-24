package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/shipping/shared/pkg/kafka"
	"github.com/shipping/shared/pkg/logger"
	"github.com/shipping/shared/pkg/utils"
	"github.com/shipping/tracking-service/internal/domain"
	"github.com/shipping/tracking-service/internal/repository"
)

type TrackingService struct {
	repo     *repository.TrackingRepo
	producer *kafka.Producer
}

func NewTrackingService(repo *repository.TrackingRepo, producer *kafka.Producer) *TrackingService {
	return &TrackingService{repo: repo, producer: producer}
}

func (s *TrackingService) AddTrackingEvent(ctx context.Context, shipmentID, trackingNumber, carrier, status, location, description string) (*domain.TrackingEvent, error) {
	event := &domain.TrackingEvent{
		ID:             utils.GenerateID(),
		ShipmentID:     shipmentID,
		TrackingNumber: trackingNumber,
		Carrier:        carrier,
		Status:         status,
		Location:       location,
		Description:    description,
		Timestamp:      time.Now(),
		CreatedAt:      time.Now(),
	}

	if err := s.repo.Create(ctx, event); err != nil {
		return nil, fmt.Errorf("add tracking: %w", err)
	}

	// Publish tracking update
	kafkaEvent := map[string]interface{}{
		"shipment_id":     shipmentID,
		"tracking_number": trackingNumber,
		"carrier":         carrier,
		"status":          status,
		"location":        location,
		"event_type":      "tracking.updated",
		"timestamp":       time.Now(),
	}
	if err := s.producer.Publish(ctx, shipmentID, kafkaEvent); err != nil {
		logger.Error("failed to publish tracking.updated", logger.String("err", err.Error()))
	}

	return event, nil
}

func (s *TrackingService) GetTrackingHistory(ctx context.Context, shipmentID string) (*domain.TrackingHistory, error) {
	events, err := s.repo.GetByShipmentID(ctx, shipmentID)
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, fmt.Errorf("no tracking history")
	}

	latest := events[0]

	// Convert slice of pointers to slice of values
	eventValues := make([]domain.TrackingEvent, len(events))
	for i, e := range events {
		eventValues[i] = *e
	}

	return &domain.TrackingHistory{
		ShipmentID:     shipmentID,
		TrackingNumber: latest.TrackingNumber,
		Carrier:        latest.Carrier,
		CurrentStatus:  latest.Status,
		Events:         eventValues,
	}, nil
}

func (s *TrackingService) PollCarrierTracking(ctx context.Context, carrierCode, trackingNumber string) (*domain.TrackingEvent, error) {
	// Call carrier integration service to get latest tracking
	carrierServiceURL := getEnv("CARRIER_SERVICE_URL", "http://carrier-integration-service:8082")
	url := fmt.Sprintf("%s/carriers/tracking?carrier=%s&tracking_number=%s", carrierServiceURL, carrierCode, trackingNumber)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("poll carrier: %w", err)
	}
	defer resp.Body.Close()

	var info struct {
		TrackingNumber string    `json:"tracking_number"`
		Carrier        string    `json:"carrier"`
		Status         string    `json:"status"`
		Location       string    `json:"location"`
		Timestamp      time.Time `json:"timestamp"`
		Description    string    `json:"description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decode tracking: %w", err)
	}

	return s.AddTrackingEvent(ctx, "", trackingNumber, carrierCode, info.Status, info.Location, info.Description)
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
