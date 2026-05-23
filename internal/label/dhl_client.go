package label

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DHLClient implements the CarrierService contract for DHL.
type DHLClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewDHLClient instantiates a new DHL Location Finder client.
func NewDHLClient(baseURL, apiKey string) *DHLClient {
	return &DHLClient{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// dhlLocationResponse mirrors the DHL Location Finder Unified API response.
type dhlLocationResponse struct {
	Locations []struct {
		URL      string `json:"url"`
		Name     string `json:"name"`
		Distance int    `json:"distance"` // metres
		Location struct {
			Type      string `json:"type"` // locker, postoffice, servicepoint, pobox
			IDs       []struct {
				LocationID string `json:"locationId"`
				Provider   string `json:"provider"`
			} `json:"ids"`
			Keyword   string `json:"keyword"`
			KeywordID string `json:"keywordId"`
		} `json:"location"`
		Place struct {
			Address struct {
				CountryCode     string `json:"countryCode"`
				PostalCode      string `json:"postalCode"`
				AddressLocality string `json:"addressLocality"`
				StreetAddress   string `json:"streetAddress"`
			} `json:"address"`
			Geo struct {
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			} `json:"geo"`
		} `json:"place"`
		OpeningHours []struct {
			Opens     string `json:"opens"`
			Closes    string `json:"closes"`
			DayOfWeek string `json:"dayOfWeek"` // e.g. "http://schema.org/Monday"
		} `json:"openingHours"`
		ServiceTypes []string `json:"serviceTypes"`
	} `json:"locations"`
}

// SearchLocations implements CarrierService — queries DHL find-by-address API.
func (c *DHLClient) SearchLocations(ctx context.Context, carrier, addressStr string) ([]LocationDetail, error) {
	city, _, postal, country := ParseAddress(addressStr)

	params := url.Values{}
	params.Set("countryCode", country)
	if postal != "" {
		params.Set("postalCode", postal)
	}
	if city != "" {
		params.Set("addressLocality", city)
	}
	// Limit to 10 nearest results
	params.Set("limit", "10")

	reqURL := fmt.Sprintf("%s/location-finder/v1/find-by-address?%s", c.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("dhl: failed to create request: %w", err)
	}
	req.Header.Set("DHL-API-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("dhl: request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("dhl: failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dhl: API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var raw dhlLocationResponse
	if err := json.Unmarshal(bodyBytes, &raw); err != nil {
		return nil, fmt.Errorf("dhl: failed to parse response JSON: %w", err)
	}

	results := make([]LocationDetail, 0, len(raw.Locations))
	for _, loc := range raw.Locations {
		// Convert metres to km for display
		distanceKM := float64(loc.Distance) / 1000.0

		// Map opening hours, stripping the schema.org prefix from day names
		var hours []OpeningHours
		for _, oh := range loc.OpeningHours {
			day := oh.DayOfWeek
			if idx := strings.LastIndex(day, "/"); idx >= 0 {
				day = day[idx+1:] // e.g. "Monday"
			}
			hours = append(hours, OpeningHours{
				DayOfWeek: day,
				Opens:     oh.Opens,
				Closes:    oh.Closes,
			})
		}

		name := loc.Name
		if name == "" {
			name = loc.Location.Keyword
		}
		if name == "" {
			name = "DHL Location"
		}

		var streetLines []string
		if loc.Place.Address.StreetAddress != "" {
			streetLines = []string{loc.Place.Address.StreetAddress}
		}

		results = append(results, LocationDetail{
			Carrier:      "DHL",
			LocationType: loc.Location.Type,
			Distance:     distanceKM,
			Units:        "km",
			Name:         name,
			StreetLines:  streetLines,
			City:         loc.Place.Address.AddressLocality,
			PostalCode:   loc.Place.Address.PostalCode,
			CountryCode:  loc.Place.Address.CountryCode,
			OpeningHours: hours,
			ServiceTypes: loc.ServiceTypes,
		})
	}

	return results, nil
}

// GenerateLabel produces a simulated DHL shipment label (sandbox stub).
func (c *DHLClient) GenerateLabel(ctx context.Context, shipmentID, carrier string, weight float64, origin, destination string) (*Label, error) {
	// DHL tracking numbers: 10 digits, starting with JD
	trackingNumber := fmt.Sprintf("JD%010d", time.Now().UnixNano()%10000000000)

	fmt.Printf("\n=== SIMULATING DHL SANDBOX LABEL GENERATION ===\n")
	fmt.Printf("Endpoint: %s/shipments\n", c.baseURL)
	fmt.Printf("Carrier:  DHL\n")
	fmt.Printf("Weight:   %.2f kg\n", weight)
	fmt.Printf("Origin:   %s\n", origin)
	fmt.Printf("Dest:     %s\n", destination)
	fmt.Printf("Tracking: %s\n", trackingNumber)
	fmt.Println("=================================================")

	labelURL := fmt.Sprintf("https://dhl-sandbox-labels.s3.amazonaws.com/labels/%s.pdf", trackingNumber)

	now := time.Now()
	return &Label{
		ID:             fmt.Sprintf("lbl-%09d", now.UnixNano()%1000000000),
		ShipmentID:     shipmentID,
		TrackingNumber: trackingNumber,
		LabelURL:       labelURL,
		Status:         "ACTIVE",
		CreatedAt:      now,
	}, nil
}
