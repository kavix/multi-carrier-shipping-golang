package client

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/shipping/carrier-integration-service/internal/domain"
)

// CarrierClient interface for all carrier APIs
type CarrierClient interface {
	GetRates(from, to string, weight float64) ([]domain.CarrierRate, error)
	GetTracking(trackingNumber string) (*domain.TrackingInfo, error)
	GetPickupLocations(address string, limit int) ([]domain.PickupDropLocation, error)
	GetDropLocations(address string, limit int) ([]domain.PickupDropLocation, error)
	ValidatePostalCode(countryCode, postalCode string) (bool, error)
}

// DHLClient implements CarrierClient for DHL API
type DHLClient struct {
	apiKey    string
	apiSecret string
	baseURL   string
	client    *http.Client
}

func NewDHLClient(apiKey, apiSecret, baseURL string) *DHLClient {
	return &DHLClient{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		baseURL:   baseURL,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *DHLClient) GetRates(from, to string, weight float64) ([]domain.CarrierRate, error) {
	// Simulated DHL API call - replace with actual DHL API integration
	// DHL API: POST /rates
	return []domain.CarrierRate{
		{
			CarrierID:     "dhl-express",
			CarrierName:   "DHL",
			ServiceType:   "Express Worldwide",
			EstimatedDays: 2,
			Cost:          weight * 12.50,
			Currency:      "USD",
		},
		{
			CarrierID:     "dhl-economy",
			CarrierName:   "DHL",
			ServiceType:   "Economy Select",
			EstimatedDays: 5,
			Cost:          weight * 8.00,
			Currency:      "USD",
		},
	}, nil
}

func (c *DHLClient) GetTracking(trackingNumber string) (*domain.TrackingInfo, error) {
	// Simulated DHL tracking - replace with actual DHL Tracking API
	return &domain.TrackingInfo{
		TrackingNumber: trackingNumber,
		Carrier:        "DHL",
		Status:         "in_transit",
		Location:       "Frankfurt Hub",
		Timestamp:      time.Now(),
		Description:    "Shipment has arrived at DHL facility",
	}, nil
}

func (c *DHLClient) GetPickupLocations(address string, limit int) ([]domain.PickupDropLocation, error) {
	return []domain.PickupDropLocation{
		{ID: "dhl-1", Carrier: "DHL", Name: "DHL Service Point", Address: "123 Main St", City: "New York", Country: "US", PostalCode: "10001", Latitude: 40.7128, Longitude: -74.0060, Type: "pickup", DistanceKm: 1.2},
		{ID: "dhl-2", Carrier: "DHL", Name: "DHL Express Center", Address: "456 Broadway", City: "New York", Country: "US", PostalCode: "10013", Latitude: 40.7190, Longitude: -74.0020, Type: "pickup", DistanceKm: 2.5},
	}, nil
}

func (c *DHLClient) GetDropLocations(address string, limit int) ([]domain.PickupDropLocation, error) {
	return []domain.PickupDropLocation{
		{ID: "dhl-drop-1", Carrier: "DHL", Name: "DHL Drop Box", Address: "789 Park Ave", City: "New York", Country: "US", PostalCode: "10016", Latitude: 40.7489, Longitude: -73.9680, Type: "drop", DistanceKm: 0.8},
	}, nil
}

func (c *DHLClient) ValidatePostalCode(countryCode, postalCode string) (bool, error) {
	// Simulate DHL postal code validation
	return true, nil
}

// UPSClient implements CarrierClient for UPS API
type UPSClient struct {
	apiKey    string
	apiSecret string
	baseURL   string
	client    *http.Client
}

func NewUPSClient(apiKey, apiSecret, baseURL string) *UPSClient {
	return &UPSClient{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		baseURL:   baseURL,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *UPSClient) GetRates(from, to string, weight float64) ([]domain.CarrierRate, error) {
	return []domain.CarrierRate{
		{
			CarrierID:     "ups-express",
			CarrierName:   "UPS",
			ServiceType:   "UPS Worldwide Express",
			EstimatedDays: 1,
			Cost:          weight * 14.00,
			Currency:      "USD",
		},
		{
			CarrierID:     "ups-saver",
			CarrierName:   "UPS",
			ServiceType:   "UPS Worldwide Saver",
			EstimatedDays: 3,
			Cost:          weight * 10.00,
			Currency:      "USD",
		},
	}, nil
}

func (c *UPSClient) GetTracking(trackingNumber string) (*domain.TrackingInfo, error) {
	return &domain.TrackingInfo{
		TrackingNumber: trackingNumber,
		Carrier:        "UPS",
		Status:         "in_transit",
		Location:       "Louisville, KY",
		Timestamp:      time.Now(),
		Description:    "Arrived at UPS Worldport",
	}, nil
}

func (c *UPSClient) GetPickupLocations(address string, limit int) ([]domain.PickupDropLocation, error) {
	return []domain.PickupDropLocation{
		{ID: "ups-1", Carrier: "UPS", Name: "The UPS Store", Address: "300 5th Ave", City: "New York", Country: "US", PostalCode: "10016", Latitude: 40.7480, Longitude: -73.9850, Type: "pickup", DistanceKm: 0.5},
	}, nil
}

func (c *UPSClient) GetDropLocations(address string, limit int) ([]domain.PickupDropLocation, error) {
	return []domain.PickupDropLocation{
		{ID: "ups-drop-1", Carrier: "UPS", Name: "UPS Access Point", Address: "400 Madison Ave", City: "New York", Country: "US", PostalCode: "10017", Latitude: 40.7570, Longitude: -73.9770, Type: "drop", DistanceKm: 1.0},
	}, nil
}

func (c *UPSClient) ValidatePostalCode(countryCode, postalCode string) (bool, error) {
	return true, nil
}

// CarrierClientFactory creates the right client based on carrier code
func CarrierClientFactory(carrier *domain.Carrier) (CarrierClient, error) {
	switch strings.ToLower(carrier.Code) {
	case "dhl":
		return NewDHLClient(carrier.APIKey, carrier.APISecret, carrier.BaseURL), nil
	case "fedex":
		return NewFedExClient(carrier.APIKey, carrier.APISecret, carrier.BaseURL), nil
	case "ups":
		return NewUPSClient(carrier.APIKey, carrier.APISecret, carrier.BaseURL), nil
	default:
		return nil, fmt.Errorf("unsupported carrier: %s", carrier.Code)
	}
}
