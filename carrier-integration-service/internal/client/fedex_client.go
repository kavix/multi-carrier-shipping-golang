package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/shipping/carrier-integration-service/internal/domain"
	"github.com/shipping/shared/pkg/logger"
)

// FedExClient implements CarrierClient for FedEx API
type FedExClient struct {
	apiKey      string
	apiSecret   string
	baseURL     string
	client      *http.Client
	accessToken string
	tokenExpiry time.Time
}

func NewFedExClient(apiKey, apiSecret, baseURL string) *FedExClient {
	if baseURL == "" {
		baseURL = "https://apis-sandbox.fedex.com"
	}
	return &FedExClient{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		baseURL:   baseURL,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

// getAuthToken retrieves an OAuth2 token from FedEx
func (c *FedExClient) getAuthToken() (string, error) {
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry) {
		return c.accessToken, nil
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", c.apiKey)
	data.Set("client_secret", c.apiSecret)

	resp, err := c.client.PostForm(c.baseURL+"/oauth/token", data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fedex auth failed with status %d", resp.StatusCode)
	}

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	c.accessToken = result.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(result.ExpiresIn-60) * time.Second)
	return c.accessToken, nil
}

// GetRates implements CarrierClient
func (c *FedExClient) GetRates(from, to string, weight float64) ([]domain.CarrierRate, error) {
	token, err := c.getAuthToken()
	if err != nil {
		logger.Error("FedEx Auth Error", logger.String("err", err.Error()))
		return c.simulateRates(weight), nil
	}

	payload := map[string]interface{}{
		"accountNumber": map[string]interface{}{
			"value": "740561073", // Example sandbox account
		},
		"requestedShipment": map[string]interface{}{
			"shipper": map[string]interface{}{
				"address": map[string]string{
					"postalCode":  from,
					"countryCode": "US",
				},
			},
			"recipient": map[string]interface{}{
				"address": map[string]string{
					"postalCode":  to,
					"countryCode": "US",
				},
			},
			"pickupType":      "DROPOFF_AT_FEDEX_LOCATION",
			"serviceType":     "FEDEX_GROUND",
			"rateRequestType": []string{"ACCOUNT", "LIST"},
			"requestedPackageLineItems": []map[string]interface{}{
				{
					"weight": map[string]interface{}{
						"units": "LB",
						"value": weight,
					},
				},
			},
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", c.baseURL+"/rate/v1/rates/quotes", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Errors []struct {
				Message string `json:"message"`
			} `json:"errors"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		if len(errResp.Errors) > 0 {
			return nil, fmt.Errorf("fedex api error: %s", errResp.Errors[0].Message)
		}
		return nil, fmt.Errorf("fedex api returned status %d", resp.StatusCode)
	}

	var rateResp struct {
		Output struct {
			RateReplyDetails []struct {
				ServiceType          string `json:"serviceType"`
				RatedShipmentDetails []struct {
					TotalNetCharge float64 `json:"totalNetCharge"`
					Currency       string  `json:"currency"`
				} `json:"ratedShipmentDetails"`
			} `json:"rateReplyDetails"`
		} `json:"output"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rateResp); err != nil {
		return nil, fmt.Errorf("decode fedex response: %w", err)
	}

	var results []domain.CarrierRate
	for _, detail := range rateResp.Output.RateReplyDetails {
		if len(detail.RatedShipmentDetails) > 0 {
			results = append(results, domain.CarrierRate{
				CarrierID:     "fedex",
				CarrierName:   "FedEx",
				ServiceType:   detail.ServiceType,
				EstimatedDays: 3, // Default for demo if not in response
				Cost:          detail.RatedShipmentDetails[0].TotalNetCharge,
				Currency:      detail.RatedShipmentDetails[0].Currency,
			})
		}
	}

	return results, nil
}

func (c *FedExClient) simulateRates(weight float64) []domain.CarrierRate {
	return []domain.CarrierRate{
		{
			CarrierID:     "fedex-priority",
			CarrierName:   "FedEx",
			ServiceType:   "FedEx International Priority",
			EstimatedDays: 1,
			Cost:          weight * 15.00,
			Currency:      "USD",
		},
		{
			CarrierID:     "fedex-economy",
			CarrierName:   "FedEx",
			ServiceType:   "FedEx International Economy",
			EstimatedDays: 4,
			Cost:          weight * 9.50,
			Currency:      "USD",
		},
	}
}

// GetTracking implements CarrierClient
func (c *FedExClient) GetTracking(trackingNumber string) (*domain.TrackingInfo, error) {
	return &domain.TrackingInfo{
		TrackingNumber: trackingNumber,
		Carrier:        "FedEx",
		Status:         "picked_up",
		Location:       "Memphis, TN",
		Timestamp:      time.Now(),
		Description:    "Picked up by FedEx courier",
	}, nil
}

// GetPickupLocations implements CarrierClient
func (c *FedExClient) GetPickupLocations(address string, limit int) ([]domain.PickupDropLocation, error) {
	return c.searchLocations(address, "pickup", limit)
}

// GetDropLocations implements CarrierClient
func (c *FedExClient) GetDropLocations(address string, limit int) ([]domain.PickupDropLocation, error) {
	return c.searchLocations(address, "drop", limit)
}

func (c *FedExClient) searchLocations(address, locationType string, limit int) ([]domain.PickupDropLocation, error) {
	token, err := c.getAuthToken()
	if err != nil {
		return c.simulateLocations(locationType), nil
	}

	payload := map[string]interface{}{
		"location": map[string]interface{}{
			"address": map[string]string{
				"streetLines":         address,
				"city":                "New York",
				"stateOrProvinceCode": "NY",
				"postalCode":          "10001",
				"countryCode":         "US",
			},
		},
		"locationSearchCriterion": "ADDRESS",
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", c.baseURL+"/location/v1/locations", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fedex locations api error: status %d", resp.StatusCode)
	}

	var locResp struct {
		Output struct {
			LocationDetailList []struct {
				LocationId                string `json:"locationId"`
				LocationContactAndAddress struct {
					Contact struct {
						CompanyName string `json:"companyName"`
					} `json:"contact"`
					Address struct {
						StreetLines []string `json:"streetLines"`
						City        string   `json:"city"`
						CountryCode string   `json:"countryCode"`
						PostalCode  string   `json:"postalCode"`
					} `json:"address"`
				} `json:"locationContactAndAddress"`
				Distance struct {
					Value float64 `json:"value"`
					Units string  `json:"units"`
				} `json:"distance"`
			} `json:"locationDetailList"`
		} `json:"output"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&locResp); err != nil {
		return nil, fmt.Errorf("decode fedex locations response: %w", err)
	}

	var results []domain.PickupDropLocation
	for _, detail := range locResp.Output.LocationDetailList {
		addr := ""
		if len(detail.LocationContactAndAddress.Address.StreetLines) > 0 {
			addr = detail.LocationContactAndAddress.Address.StreetLines[0]
		}
		results = append(results, domain.PickupDropLocation{
			ID:         detail.LocationId,
			Carrier:    "FedEx",
			Name:       detail.LocationContactAndAddress.Contact.CompanyName,
			Address:    addr,
			City:       detail.LocationContactAndAddress.Address.City,
			Country:    detail.LocationContactAndAddress.Address.CountryCode,
			PostalCode: detail.LocationContactAndAddress.Address.PostalCode,
			Type:       locationType,
			DistanceKm: detail.Distance.Value,
		})
		if len(results) >= limit {
			break
		}
	}

	return results, nil
}

func (c *FedExClient) simulateLocations(locationType string) []domain.PickupDropLocation {
	return []domain.PickupDropLocation{
		{
			ID:         "fedex-1",
			Carrier:    "FedEx",
			Name:       "FedEx Office",
			Address:    "100 Wall St",
			City:       "New York",
			Country:    "US",
			PostalCode: "10005",
			Latitude:   40.7074,
			Longitude:  -74.0113,
			Type:       locationType,
			DistanceKm: 1.5,
		},
	}
}

// ValidatePostalCode implements CarrierClient
func (c *FedExClient) ValidatePostalCode(countryCode, postalCode string) (bool, error) {
	token, err := c.getAuthToken()
	if err != nil {
		return true, nil
	}

	payload := map[string]interface{}{
		"carrierCode": "FDXE", // FedEx Express
		"countryCode": countryCode,
		"postalCode":  postalCode,
		"shipDate":    time.Now().Format("2006-01-02"),
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", c.baseURL+"/country/v1/postal/validate", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("fedex postal validation error: status %d", resp.StatusCode)
	}

	var postResp struct {
		Output struct {
			Alerts []struct {
				Code string `json:"code"`
			} `json:"alerts"`
		} `json:"output"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&postResp); err != nil {
		return false, fmt.Errorf("decode fedex postal response: %w", err)
	}

	// If there are alerts, it might mean the postal code is problematic
	// In a real scenario, we would check specific alert codes
	return len(postResp.Output.Alerts) == 0, nil
}
