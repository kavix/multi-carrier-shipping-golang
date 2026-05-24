package service

import (
	"context"
	"fmt"
	"time"

	"github.com/shipping/carrier-integration-service/internal/client"
	"github.com/shipping/carrier-integration-service/internal/domain"
	"github.com/shipping/carrier-integration-service/internal/repository"
	"github.com/shipping/shared/pkg/utils"
)

type CarrierService struct {
	repo *repository.CarrierRepo
}

func NewCarrierService(repo *repository.CarrierRepo) *CarrierService {
	return &CarrierService{repo: repo}
}

func (s *CarrierService) RegisterCarrier(ctx context.Context, name, code, apiKey, apiSecret, baseURL string) (*domain.Carrier, error) {
	carrier := &domain.Carrier{
		ID:        utils.GenerateID(),
		Name:      name,
		Code:      code,
		APIKey:    apiKey,
		APISecret: apiSecret,
		BaseURL:   baseURL,
		IsActive:  true,
		CreatedAt: time.Now(),
	}
	if err := s.repo.Create(ctx, carrier); err != nil {
		return nil, fmt.Errorf("register carrier: %w", err)
	}
	return carrier, nil
}

func (s *CarrierService) GetCarrierRates(ctx context.Context, from, to string, weight float64) ([]domain.CarrierRate, error) {
	carriers, err := s.repo.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("list carriers: %w", err)
	}

	var allRates []domain.CarrierRate
	for _, carrier := range carriers {
		c, err := client.CarrierClientFactory(carrier)
		if err != nil {
			continue
		}
		rates, err := c.GetRates(from, to, weight)
		if err != nil {
			continue
		}
		allRates = append(allRates, rates...)
	}
	return allRates, nil
}

func (s *CarrierService) GetTracking(ctx context.Context, carrierCode, trackingNumber string) (*domain.TrackingInfo, error) {
	carrier, err := s.repo.GetByCode(ctx, carrierCode)
	if err != nil {
		return nil, err
	}
	c, err := client.CarrierClientFactory(carrier)
	if err != nil {
		return nil, err
	}
	return c.GetTracking(trackingNumber)
}

func (s *CarrierService) GetPickupLocations(ctx context.Context, carrierCode, address string, limit int) ([]domain.PickupDropLocation, error) {
	carrier, err := s.repo.GetByCode(ctx, carrierCode)
	if err != nil {
		return nil, err
	}
	c, err := client.CarrierClientFactory(carrier)
	if err != nil {
		return nil, err
	}
	return c.GetPickupLocations(address, limit)
}

func (s *CarrierService) GetDropLocations(ctx context.Context, carrierCode, address string, limit int) ([]domain.PickupDropLocation, error) {
	carrier, err := s.repo.GetByCode(ctx, carrierCode)
	if err != nil {
		return nil, err
	}
	c, err := client.CarrierClientFactory(carrier)
	if err != nil {
		return nil, err
	}
	return c.GetDropLocations(address, limit)
}
