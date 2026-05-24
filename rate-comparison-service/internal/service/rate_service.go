package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"time"

	"github.com/shipping/rate-comparison-service/internal/domain"
	"github.com/shipping/rate-comparison-service/internal/repository"
	"github.com/shipping/shared/pkg/kafka"
	"github.com/shipping/shared/pkg/logger"
	"github.com/shipping/shared/pkg/utils"
)

type RateService struct {
	repo     *repository.RateRepo
	producer *kafka.Producer
}

func NewRateService(repo *repository.RateRepo, producer *kafka.Producer) *RateService {
	return &RateService{repo: repo, producer: producer}
}

func (s *RateService) CompareRates(ctx context.Context, userID, shipmentID, from, to string, weight float64) (*domain.RateComparison, error) {
	// Call carrier integration service to get rates from all carriers
	carrierServiceURL := getEnv("CARRIER_SERVICE_URL", "http://carrier-integration-service:8082")
	apiURL := fmt.Sprintf("%s/carriers/rates?from=%s&to=%s&weight=%.2f", 
		carrierServiceURL, url.QueryEscape(from), url.QueryEscape(to), weight)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("fetch carrier rates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("carrier service returned status: %d", resp.StatusCode)
	}

	var rates []domain.RateResult
	if err := json.NewDecoder(resp.Body).Decode(&rates); err != nil {
		return nil, fmt.Errorf("decode rates: %w", err)
	}

	if len(rates) == 0 {
		return nil, fmt.Errorf("no rates available")
	}

	// Sort by cost (lowest first)
	sort.Slice(rates, func(i, j int) bool {
		return rates[i].Cost < rates[j].Cost
	})

	best := rates[0]
	allRatesJSON, _ := json.Marshal(rates)

	comparison := &domain.RateComparison{
		ID:           utils.GenerateID(),
		ShipmentID:   shipmentID,
		UserID:       userID,
		FromAddress:  from,
		ToAddress:    to,
		Weight:       weight,
		BestCarrier:  best.CarrierName,
		BestService:  best.ServiceType,
		BestCost:     best.Cost,
		BestDays:     best.EstimatedDays,
		AllRatesJSON: string(allRatesJSON),
		CreatedAt:    time.Now(),
	}

	if err := s.repo.Create(ctx, comparison); err != nil {
		return nil, fmt.Errorf("save comparison: %w", err)
	}

	// Publish event
	event := map[string]interface{}{
		"comparison_id": comparison.ID,
		"shipment_id":   shipmentID,
		"user_id":       userID,
		"best_carrier":  best.CarrierName,
		"best_cost":     best.Cost,
		"event_type":    "rates.compared",
	}
	if err := s.producer.Publish(ctx, comparison.ID, event); err != nil {
		logger.Error("failed to publish rates.compared", logger.String("err", err.Error()))
	}

	return comparison, nil
}

func (s *RateService) GetComparison(ctx context.Context, shipmentID string) (*domain.RateComparison, error) {
	return s.repo.GetByShipmentID(ctx, shipmentID)
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
