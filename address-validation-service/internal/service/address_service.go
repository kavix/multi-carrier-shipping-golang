package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/shipping/address-validation-service/internal/domain"
	"github.com/shipping/address-validation-service/internal/repository"
	"github.com/shipping/shared/pkg/kafka"
	"github.com/shipping/shared/pkg/logger"
	"github.com/shipping/shared/pkg/utils"
)

type AddressService struct {
	repo     *repository.AddressRepo
	producer *kafka.Producer
}

func NewAddressService(repo *repository.AddressRepo, producer *kafka.Producer) *AddressService {
	return &AddressService{repo: repo, producer: producer}
}

func (s *AddressService) ValidateAddress(ctx context.Context, rawAddress string) (*domain.ValidatedAddress, error) {
	// Check cache first
	cached, err := s.repo.GetByRawAddress(ctx, rawAddress)
	if err == nil && time.Since(cached.ValidatedAt) < 24*time.Hour {
		return cached, nil
	}

	// Call Google Address Validation API (or similar)
	// For demo, we simulate validation
	validated := s.simulateValidation(rawAddress)

	if err := s.repo.Save(ctx, validated); err != nil {
		logger.Error("failed to cache address", logger.String("err", err.Error()))
	}

	// Publish event
	event := map[string]interface{}{
		"address_id": validated.ID,
		"raw":        rawAddress,
		"is_valid":   validated.IsValid,
		"event_type": "address.validated",
	}
	if err := s.producer.Publish(ctx, validated.ID, event); err != nil {
		logger.Error("failed to publish address.validated", logger.String("err", err.Error()))
	}

	return validated, nil
}

func (s *AddressService) GetLocations(ctx context.Context, address, carrier string, limit int, locType string) ([]domain.Location, error) {
	// Call carrier integration service
	carrierServiceURL := getEnv("CARRIER_SERVICE_URL", "http://carrier-integration-service:8082")
	endpoint := fmt.Sprintf("/carriers/%s-locations", locType)
	url := fmt.Sprintf("%s%s?carrier=%s&address=%s&limit=%d", carrierServiceURL, endpoint, carrier, address, limit)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch locations: %w", err)
	}
	defer resp.Body.Close()

	var locations []domain.Location
	if err := json.NewDecoder(resp.Body).Decode(&locations); err != nil {
		return nil, fmt.Errorf("decode locations: %w", err)
	}
	return locations, nil
}

func (s *AddressService) simulateValidation(rawAddress string) *domain.ValidatedAddress {
	// Simulate address parsing and validation
	parts := strings.Split(rawAddress, ",")
	isValid := len(parts) >= 3

	validated := &domain.ValidatedAddress{
		ID:          utils.GenerateID(),
		RawAddress:  rawAddress,
		IsValid:     isValid,
		ValidatedAt: time.Now(),
	}

	if isValid {
		validated.Street = strings.TrimSpace(parts[0])
		validated.City = strings.TrimSpace(parts[1])
		validated.State = strings.TrimSpace(parts[2])
		validated.PostalCode = "10001"
		validated.Country = "US"
		validated.Latitude = 40.7128
		validated.Longitude = -74.0060
	}

	return validated
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
