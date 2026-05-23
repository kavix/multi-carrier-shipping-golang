package label

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type oauthToken struct {
	AccessToken string
	ExpiresAt   time.Time
}

// FedExClient implements the CarrierService contract.
type FedExClient struct {
	baseURL    string
	apiKey     string
	apiSecret  string
	httpClient *http.Client

	mu         sync.Mutex
	tokenCache *oauthToken
}

// NewFedExClient instantiates a new FedEx Integration Client.
func NewFedExClient(baseURL, apiKey, apiSecret string) *FedExClient {
	return &FedExClient{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// SearchLocations implements the CarrierService interface.
func (c *FedExClient) SearchLocations(ctx context.Context, addressStr string) ([]LocationDetail, error) {
	city, state, postal, country := ParseAddress(addressStr)

	token, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve oauth token: %w", err)
	}

	reqBody := struct {
		ControlParams struct {
			Distance struct {
				Units string  `json:"units"`
				Value float64 `json:"value"`
			} `json:"distance"`
		} `json:"locationsSummaryRequestControlParameters"`
		LocationSearchCriterion string `json:"locationSearchCriterion"`
		Location                struct {
			Address struct {
				City                string `json:"city,omitempty"`
				StateOrProvinceCode string `json:"stateOrProvinceCode,omitempty"`
				PostalCode          string `json:"postalCode,omitempty"`
				CountryCode         string `json:"countryCode,omitempty"`
			} `json:"address"`
		} `json:"location"`
	}{}

	reqBody.ControlParams.Distance.Units = "MI"
	reqBody.ControlParams.Distance.Value = 50.0 // search within 50 miles radius
	reqBody.LocationSearchCriterion = "ADDRESS"
	reqBody.Location.Address.City = city
	reqBody.Location.Address.StateOrProvinceCode = state
	reqBody.Location.Address.PostalCode = postal
	reqBody.Location.Address.CountryCode = country

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal locations request: %w", err)
	}

	reqURL := fmt.Sprintf("%s/location/v1/locations", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create locations request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("X-locale", "en_US")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to dispatch locations request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read locations response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("locations request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var respData struct {
		Output struct {
			LocationDetailList []struct {
				Distance struct {
					Value float64 `json:"value"`
					Units string  `json:"units"`
				} `json:"distance"`
				ContactAndAddress struct {
					Contact struct {
						CompanyName string `json:"companyName"`
					} `json:"contact"`
					Address struct {
						StreetLines         []string `json:"streetLines"`
						City                string   `json:"city"`
						StateOrProvinceCode string   `json:"stateOrProvinceCode"`
						PostalCode          string   `json:"postalCode"`
						CountryCode         string   `json:"countryCode"`
					} `json:"address"`
					AddressAncillaryDetail struct {
						DisplayName string `json:"displayName"`
					} `json:"addressAncillaryDetail"`
				} `json:"contactAndAddress"`
			} `json:"locationDetailList"`
		} `json:"output"`
	}

	if err := json.Unmarshal(bodyBytes, &respData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal locations response JSON: %w", err)
	}

	var results []LocationDetail
	for _, loc := range respData.Output.LocationDetailList {
		name := loc.ContactAndAddress.Contact.CompanyName
		if name == "" {
			name = loc.ContactAndAddress.AddressAncillaryDetail.DisplayName
		}
		if name == "" {
			name = "FedEx Location"
		}

		results = append(results, LocationDetail{
			Distance:            loc.Distance.Value,
			Units:               loc.Distance.Units,
			Name:                name,
			StreetLines:         loc.ContactAndAddress.Address.StreetLines,
			City:                loc.ContactAndAddress.Address.City,
			StateOrProvinceCode: loc.ContactAndAddress.Address.StateOrProvinceCode,
			PostalCode:          loc.ContactAndAddress.Address.PostalCode,
			CountryCode:         loc.ContactAndAddress.Address.CountryCode,
		})
	}

	return results, nil
}

// GenerateLabel implements the CarrierService interface.
func (c *FedExClient) GenerateLabel(ctx context.Context, shipmentID, carrier string, weight float64, origin, destination string) (*Label, error) {
	// For demo/sandbox purposes, let's create a realistic FedEx tracking number
	// Format: FTX + 9 digits
	trackingNumber := fmt.Sprintf("FTX%09d", time.Now().UnixNano()%1000000000)

	// In a real application, this would make a POST call to c.baseURL + "/ship/v1/shipments"
	// utilizing the oauthToken via c.getAccessToken(ctx).
	fmt.Printf("\n=== SIMULATING FEDEX SANDBOX LABEL GENERATION ===\n")
	fmt.Printf("Endpoint: %s/ship/v1/shipments\n", c.baseURL)
	fmt.Printf("Carrier: FedEx\n")
	fmt.Printf("Weight: %.2f lbs\n", weight)
	fmt.Printf("Origin: %s\n", origin)
	fmt.Printf("Destination: %s\n", destination)
	fmt.Printf("Generated Tracking Number: %s\n", trackingNumber)
	fmt.Println("=================================================")

	// A realistic PDF base64 encoded label mock URL
	labelURL := fmt.Sprintf("https://fedex-sandbox-labels.s3.amazonaws.com/labels/%s.pdf", trackingNumber)

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

func (c *FedExClient) getAccessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	if c.tokenCache != nil && time.Now().Before(c.tokenCache.ExpiresAt) {
		token := c.tokenCache.AccessToken
		c.mu.Unlock()
		return token, nil
	}
	c.mu.Unlock()

	reqURL := fmt.Sprintf("%s/oauth/token", c.baseURL)

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", c.apiKey)
	form.Set("client_secret", c.apiSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to dispatch token request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read token response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	if err := json.Unmarshal(bodyBytes, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response JSON: %w", err)
	}

	c.mu.Lock()
	c.tokenCache = &oauthToken{
		AccessToken: tokenResp.AccessToken,
		ExpiresAt:   time.Now().Add(time.Duration(tokenResp.ExpiresIn)*time.Second - 1*time.Minute),
	}
	token := c.tokenCache.AccessToken
	c.mu.Unlock()

	return token, nil
}

// ParseAddress parses a comma-separated address string into individual structured parts.
func ParseAddress(addressStr string) (city, state, postal, country string) {
	parts := strings.Split(addressStr, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	var filtered []string
	for _, p := range parts {
		if p != "" {
			filtered = append(filtered, p)
		}
	}

	n := len(filtered)
	if n == 0 {
		return "", "", "", "US"
	}

	isPostal := func(s string) bool {
		hasDigit := false
		for _, r := range s {
			if r >= '0' && r <= '9' {
				hasDigit = true
				break
			}
		}
		return hasDigit && len(s) >= 3
	}

	isCountry := func(s string) bool {
		return len(s) == 2
	}

	if n == 1 {
		if isCountry(filtered[0]) {
			return "", "", "", strings.ToUpper(filtered[0])
		}
		return filtered[0], "", "", "US"
	}

	if n == 2 {
		last := filtered[1]
		first := filtered[0]

		countryVal := "US"
		if isCountry(last) {
			countryVal = strings.ToUpper(last)
		}

		if isPostal(first) {
			return "", "", first, countryVal
		}
		return first, "", "", countryVal
	}

	if n == 3 {
		last := filtered[2]
		mid := filtered[1]
		first := filtered[0]

		countryVal := "US"
		if isCountry(last) {
			countryVal = strings.ToUpper(last)
		}

		if isPostal(mid) {
			return first, "", mid, countryVal
		}

		if len(mid) == 2 {
			return first, strings.ToUpper(mid), "", countryVal
		}

		return first, mid, "", countryVal
	}

	if n == 4 {
		countryVal := "US"
		if isCountry(filtered[3]) {
			countryVal = strings.ToUpper(filtered[3])
		}
		return filtered[0], strings.ToUpper(filtered[1]), filtered[2], countryVal
	}

	countryVal := "US"
	if isCountry(filtered[n-1]) {
		countryVal = strings.ToUpper(filtered[n-1])
	}

	postalVal := filtered[n-2]
	stateVal := strings.ToUpper(filtered[n-3])
	cityVal := filtered[n-4]

	return cityVal, stateVal, postalVal, countryVal
}
