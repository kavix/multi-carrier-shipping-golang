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
func (c *FedExClient) SearchLocations(ctx context.Context, carrier, addressStr string) ([]LocationDetail, error) {
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
			Carrier:             "FedEx",
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

// GetRatesAndTransitTimes calls POST /rate/v1/rates/quotes and returns structured RateQuote slice.
// Works for worldwide shipments — no country restriction applied.
func (c *FedExClient) GetRatesAndTransitTimes(ctx context.Context, weight float64, originPostal, originCountry, destPostal, destCountry string) ([]RateQuote, string, error) {
	token, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("fedex rates: token error: %w", err)
	}

	// Build minimal worldwide rate request per the Rates and Transit Times API spec
	type money struct {
		Amount   float64 `json:"amount"`
		Currency string  `json:"currency"`
	}
	type addr struct {
		PostalCode  string `json:"postalCode,omitempty"`
		CountryCode string `json:"countryCode"`
	}
	type contact struct {
		PersonName  string `json:"personName"`
		PhoneNumber string `json:"phoneNumber"`
	}
	type party struct {
		Contact contact `json:"contact"`
		Address addr    `json:"address"`
	}
	type weightObj struct {
		Units string  `json:"units"`
		Value float64 `json:"value"`
	}
	type pkg struct {
		Weight weightObj `json:"weight"`
	}

	reqBody := map[string]interface{}{
		"accountNumber": map[string]string{
			"value": "740561073", // sandbox account
		},
		"requestedShipment": map[string]interface{}{
			"shipper":      party{Contact: contact{"Shipper", "1234567890"}, Address: addr{PostalCode: originPostal, CountryCode: originCountry}},
			"recipient":    party{Contact: contact{"Recipient", "0987654321"}, Address: addr{PostalCode: destPostal, CountryCode: destCountry}},
			"pickupType":   "DROPOFF_AT_FEDEX_LOCATION",
			"rateRequestType": []string{"LIST", "ACCOUNT"},
			"requestedPackageLineItems": []pkg{
				{Weight: weightObj{Units: "KG", Value: weight}},
			},
		},
	}

	b, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/rate/v1/rates/quotes", bytes.NewReader(b))
	if err != nil {
		return nil, "", fmt.Errorf("fedex rates: request build error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-locale", "en_US")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("fedex rates: http error: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("fedex rates: API returned %d: %s", resp.StatusCode, string(body))
	}

	// Parse response — matches RateOutputVO schema from API docs
	var raw struct {
		Output struct {
			QuoteDate       string `json:"quoteDate"`
			RateReplyDetails []struct {
				ServiceType  string `json:"serviceType"`
				ServiceName  string `json:"serviceName"`
				RatedShipmentDetails []struct {
					RateType         string  `json:"rateType"`
					TotalBaseCharge  float64 `json:"totalBaseCharge"`
					TotalNetCharge   float64 `json:"totalNetCharge"`
					Currency         string  `json:"currency"`
					ShipmentRateDetail struct {
						FuelSurchargePercent float64 `json:"fuelSurchargePercent"`
						TotalSurcharges      float64 `json:"totalSurcharges"`
						RateZone             string  `json:"rateZone"`
						SurCharges []struct {
							Type        string  `json:"type"`
							Description string  `json:"description"`
							Amount      float64 `json:"amount"`
						} `json:"surCharges"`
					} `json:"shipmentRateDetail"`
				} `json:"ratedShipmentDetails"`
				OperationalDetail struct {
					TransitTime  string `json:"transitTime"`
					DeliveryDay  string `json:"deliveryDay"`
					CommitDate   string `json:"commitDate"`
					DestinationPostalCode string `json:"destinationPostalCode"`
				} `json:"operationalDetail"`
				Commit struct {
					DateDetail struct {
						DayOfWeek    string `json:"dayOfWeek"`
						DayCxsFormat string `json:"dayCxsFormat"`
					} `json:"dateDetail"`
				} `json:"commit"`
			} `json:"rateReplyDetails"`
		} `json:"output"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, "", fmt.Errorf("fedex rates: JSON parse error: %w", err)
	}

	var quotes []RateQuote
	seen := map[string]bool{}
	for _, svc := range raw.Output.RateReplyDetails {
		// Pick the best rate detail (prefer LIST rate for display)
		var best *struct {
			RateType         string  `json:"rateType"`
			TotalBaseCharge  float64 `json:"totalBaseCharge"`
			TotalNetCharge   float64 `json:"totalNetCharge"`
			Currency         string  `json:"currency"`
			ShipmentRateDetail struct {
				FuelSurchargePercent float64 `json:"fuelSurchargePercent"`
				TotalSurcharges      float64 `json:"totalSurcharges"`
				RateZone             string  `json:"rateZone"`
				SurCharges []struct {
					Type        string  `json:"type"`
					Description string  `json:"description"`
					Amount      float64 `json:"amount"`
				} `json:"surCharges"`
			} `json:"shipmentRateDetail"`
		}
		for i := range svc.RatedShipmentDetails {
			d := &svc.RatedShipmentDetails[i]
			if best == nil || d.RateType == "LIST" {
				best = (*struct {
					RateType         string  `json:"rateType"`
					TotalBaseCharge  float64 `json:"totalBaseCharge"`
					TotalNetCharge   float64 `json:"totalNetCharge"`
					Currency         string  `json:"currency"`
					ShipmentRateDetail struct {
						FuelSurchargePercent float64 `json:"fuelSurchargePercent"`
						TotalSurcharges      float64 `json:"totalSurcharges"`
						RateZone             string  `json:"rateZone"`
						SurCharges []struct {
							Type        string  `json:"type"`
							Description string  `json:"description"`
							Amount      float64 `json:"amount"`
						} `json:"surCharges"`
					} `json:"shipmentRateDetail"`
				})(d)
			}
		}
		if best == nil {
			continue
		}
		key := svc.ServiceType + best.Currency
		if seen[key] {
			continue
		}
		seen[key] = true

		var surcharges []Surcharge
		for _, s := range best.ShipmentRateDetail.SurCharges {
			surcharges = append(surcharges, Surcharge{
				Type:        s.Type,
				Description: s.Description,
				Amount:      s.Amount,
				Currency:    best.Currency,
			})
		}

		commitDT := svc.OperationalDetail.CommitDate
		if svc.Commit.DateDetail.DayCxsFormat != "" {
			commitDT = svc.Commit.DateDetail.DayCxsFormat
		}
		deliveryDay := svc.OperationalDetail.DeliveryDay
		if svc.Commit.DateDetail.DayOfWeek != "" {
			deliveryDay = svc.Commit.DateDetail.DayOfWeek
		}

		quotes = append(quotes, RateQuote{
			ServiceType:        svc.ServiceType,
			ServiceName:        svc.ServiceName,
			Currency:           best.Currency,
			BaseCharge:         best.TotalBaseCharge,
			TotalNetCharge:     best.TotalNetCharge,
			FuelSurcharge:      best.ShipmentRateDetail.FuelSurchargePercent,
			TotalSurcharges:    best.ShipmentRateDetail.TotalSurcharges,
			Surcharges:         surcharges,
			TransitTime:        svc.OperationalDetail.TransitTime,
			DeliveryDay:        deliveryDay,
			CommitDateTime:     commitDT,
			DeliveryPostalCode: svc.OperationalDetail.DestinationPostalCode,
			RateZone:           best.ShipmentRateDetail.RateZone,
		})
	}

	return quotes, raw.Output.QuoteDate, nil
}

// ValidatePostalCode calls POST /country/v1/postal/validate and returns structured PostalInfo.
// Works worldwide — uses FDXE (FedEx Express) carrier code which has broadest international coverage.
func (c *FedExClient) ValidatePostalCode(ctx context.Context, postalCode, stateOrProvince, countryCode string) (*PostalInfo, error) {
	token, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("fedex postal: token error: %w", err)
	}

	// Use tomorrow as shipDate (must be within next 10 days, not today)
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")

	reqBody := map[string]interface{}{
		"carrierCode":         "FDXE", // FedEx Express — best worldwide coverage
		"countryCode":         strings.ToUpper(countryCode),
		"stateOrProvinceCode": stateOrProvince,
		"postalCode":          postalCode,
		"shipDate":            tomorrow,
		"checkForMismatch":    false, // Don't validate state vs postal match for international
	}

	b, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/country/v1/postal/validate", bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("fedex postal: request build: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-locale", "en_US")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fedex postal: http error: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fedex postal: API returned %d: %s", resp.StatusCode, string(body))
	}

	var raw struct {
		Output struct {
			CountryCode         string `json:"countryCode"`
			StateOrProvinceCode string `json:"stateOrProvinceCode"`
			CityFirstInitials   string `json:"cityFirstInitials"`
			CleanedPostalCode   string `json:"cleanedPostalCode"`
			LocationDescriptions []struct {
				LocationID     string `json:"locationId"`
				LocationNumber string `json:"locationNumber"`
				ServiceArea    string `json:"serviceArea"`
				AirportID      string `json:"airportId"`
			} `json:"locationDescriptions"`
		} `json:"output"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("fedex postal: JSON parse: %w", err)
	}

	info := &PostalInfo{
		PostalCode:          raw.Output.CleanedPostalCode,
		CityFirstInitials:   raw.Output.CityFirstInitials,
		StateOrProvinceCode: raw.Output.StateOrProvinceCode,
		CountryCode:         raw.Output.CountryCode,
	}
	if raw.Output.CleanedPostalCode == "" {
		info.PostalCode = postalCode // use original if cleaned not returned
	}
	if len(raw.Output.LocationDescriptions) > 0 {
		info.AirportID = raw.Output.LocationDescriptions[0].AirportID
		info.ServiceArea = raw.Output.LocationDescriptions[0].ServiceArea
		info.LocationID = raw.Output.LocationDescriptions[0].LocationID
	}

	return info, nil
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
	fmt.Printf("Weight: %.2f kg\n", weight)
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
