package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFedExClient_GetRates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "mock-token",
				"expires_in":   3600,
			})
			return
		}

		if r.URL.Path == "/rate/v1/rates/quotes" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"output": map[string]interface{}{
					"rateReplyDetails": []interface{}{
						map[string]interface{}{
							"serviceType": "FEDEX_GROUND",
							"ratedShipmentDetails": []interface{}{
								map[string]interface{}{
									"totalNetCharge": 15.50,
									"currency":       "USD",
								},
							},
						},
					},
				},
			})
			return
		}
	}))
	defer server.Close()

	client := NewFedExClient("key", "secret", server.URL)
	rates, err := client.GetRates("10001", "90210", 5.0)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(rates) != 1 {
		t.Errorf("expected 1 rate, got %d", len(rates))
	}
	if rates[0].CarrierName != "FedEx" {
		t.Errorf("expected carrier FedEx, got %s", rates[0].CarrierName)
	}
	if rates[0].Cost != 15.50 {
		t.Errorf("expected cost 15.50, got %f", rates[0].Cost)
	}
}

func TestFedExClient_ValidatePostalCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "mock-token",
				"expires_in":   3600,
			})
			return
		}

		if r.URL.Path == "/country/v1/postal/validate" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"output": map[string]interface{}{
					"cleanedPostalCode": "10001",
					"alerts":            []interface{}{},
				},
			})
			return
		}
	}))
	defer server.Close()

	client := NewFedExClient("key", "secret", server.URL)
	valid, err := client.ValidatePostalCode("US", "10001")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !valid {
		t.Error("expected valid true, got false")
	}
}

func TestFedExClient_GetLocations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "mock-token",
				"expires_in":   3600,
			})
			return
		}

		if r.URL.Path == "/location/v1/locations" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"output": map[string]interface{}{
					"locationDetailList": []interface{}{
						map[string]interface{}{
							"locationId": "LOC1",
							"locationContactAndAddress": map[string]interface{}{
								"contact": map[string]interface{}{
									"companyName": "FedEx Test Office",
								},
								"address": map[string]interface{}{
									"streetLines": []interface{}{"123 Main St"},
									"city":        "New York",
									"countryCode": "US",
									"postalCode":  "10001",
								},
							},
						},
					},
				},
			})
			return
		}
	}))
	defer server.Close()

	client := NewFedExClient("key", "secret", server.URL)
	locations, err := client.GetPickupLocations("New York", 1)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(locations) != 1 {
		t.Errorf("expected 1 location, got %d", len(locations))
	}
	if locations[0].ID != "LOC1" {
		t.Errorf("expected LOC1, got %s", locations[0].ID)
	}
	if locations[0].Name != "FedEx Test Office" {
		t.Errorf("expected FedEx Test Office, got %s", locations[0].Name)
	}
	if locations[0].Address != "123 Main St" {
		t.Errorf("expected 123 Main St, got %s", locations[0].Address)
	}
}
