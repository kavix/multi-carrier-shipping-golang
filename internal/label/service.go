package label

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type labelService struct {
	repo               LabelRepository
	carrierSvc         CarrierService
	shipmentServiceURL string
	authServiceURL     string
	httpClient         *http.Client
}

// NewLabelService instantiates a new LabelService implementation.
func NewLabelService(repo LabelRepository, carrierSvc CarrierService, shipmentServiceURL, authServiceURL string) LabelService {
	return &labelService{
		repo:               repo,
		carrierSvc:         carrierSvc,
		shipmentServiceURL: strings.TrimSuffix(shipmentServiceURL, "/"),
		authServiceURL:     strings.TrimSuffix(authServiceURL, "/"),
		httpClient:         &http.Client{Timeout: 10 * time.Second},
	}
}

// Helper to verify auth token with Auth Microservice
func (s *labelService) verifyToken(ctx context.Context, token string) (string, error) {
	if token == "" {
		return "", errors.New("unauthorized: missing token")
	}

	verifyURL := fmt.Sprintf("%s/api/v1/auth/verify?token=%s", s.authServiceURL, token)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, verifyURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create verify request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("auth service is currently unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return "", errors.New("unauthorized: invalid or expired session token")
	} else if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth verification failed with status %d", resp.StatusCode)
	}

	var data struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("failed to decode verify response: %w", err)
	}

	return data.Username, nil
}

// Helper to log user actions with Auth Microservice
func (s *labelService) logAction(ctx context.Context, username, action string) {
	if username == "" || s.authServiceURL == "" {
		return
	}

	payload := map[string]string{
		"username": username,
		"action":   action,
	}
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return
	}

	logURL := fmt.Sprintf("%s/api/v1/auth/logs", s.authServiceURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, logURL, bytes.NewReader(jsonBytes))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err == nil {
		resp.Body.Close()
	}
}

func (s *labelService) CreateLabel(
	ctx context.Context,
	shipmentID, carrier string,
	weight float64,
	origin, destination string,
) (*LabelCreateResponse, error) {
	// ── Validation ───────────────────────────────────────────────────
	if carrier == "" {
		return nil, ErrCarrierRequired
	}
	if weight <= 0 {
		return nil, ErrInvalidWeight
	}

	// ── 1. Generate the label ─────────────────────────────────────────
	var label *Label
	var err error

	if s.carrierSvc != nil {
		label, err = s.carrierSvc.GenerateLabel(ctx, shipmentID, carrier, weight, origin, destination)
		if err != nil {
			return nil, fmt.Errorf("failed to generate label via carrier: %w", err)
		}
	} else {
		now := time.Now()
		trackingNumber := fmt.Sprintf("MOCK%09d", now.UnixNano()%1000000000)
		label = &Label{
			ID:             fmt.Sprintf("lbl-%09d", now.UnixNano()%1000000000),
			ShipmentID:     shipmentID,
			TrackingNumber: trackingNumber,
			LabelURL:       fmt.Sprintf("https://mock-carrier-labels.s3.amazonaws.com/labels/%s.pdf", trackingNumber),
			Status:         "ACTIVE",
			CreatedAt:      now,
		}
	}

	// ── 2. Persist label ──────────────────────────────────────────────
	if err := s.repo.Create(ctx, label); err != nil {
		return nil, err
	}

	// ── 3. Build rich response ────────────────────────────────────────
	resp := &LabelCreateResponse{
		Label:   label,
		Carrier: strings.ToUpper(carrier),
	}

	// Only enrich with FedEx-specific data when FedEx is the carrier
	if strings.EqualFold(carrier, "fedex") && s.carrierSvc != nil {
		// Try to get a *FedExClient from the carrier service (MultiCarrierClient or direct)
		var fedex *FedExClient
		switch v := s.carrierSvc.(type) {
		case *FedExClient:
			fedex = v
		case *MultiCarrierClient:
			fedex = v.fedex
		}

		if fedex != nil {
			_, _, originPostal, originCountry := ParseAddress(origin)
			_, _, destPostal, destCountry := ParseAddress(destination)

			// 3a. Postal validation — best-effort
			if originPostal != "" {
				if pInfo, pErr := fedex.ValidatePostalCode(ctx, originPostal, "", originCountry); pErr == nil {
					resp.OriginPostal = pInfo
				}
			}
			if destPostal != "" {
				if pInfo, pErr := fedex.ValidatePostalCode(ctx, destPostal, "", destCountry); pErr == nil {
					resp.DestinationPostal = pInfo
				}
			}

			// 3b. Rate quotes — best-effort
			if quotes, quoteDate, rErr := fedex.GetRatesAndTransitTimes(
				ctx, weight, originPostal, originCountry, destPostal, destCountry,
			); rErr == nil {
				resp.RateQuotes = quotes
				resp.QuoteDate = quoteDate
			}
		}

		// 3c. Drop-off locations — best-effort
		if originLocs, lErr := s.carrierSvc.SearchLocations(ctx, carrier, origin); lErr == nil {
			resp.OriginDropOffLocations = originLocs
		}
		if destLocs, lErr := s.carrierSvc.SearchLocations(ctx, carrier, destination); lErr == nil {
			resp.DestinationDropOffLocations = destLocs
		}

		// Print structured summary to server logs
		s.printFedExSummary(resp)
	}

	return resp, nil
}

// printFedExSummary logs the rich response to server stdout for debugging.
func (s *labelService) printFedExSummary(resp *LabelCreateResponse) {
	fmt.Printf("\n╔══════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  FEDEX LABEL CREATED — %s\n", resp.Label.TrackingNumber)
	fmt.Printf("╚══════════════════════════════════════════════════════════════╝\n")

	if resp.OriginPostal != nil {
		p := resp.OriginPostal
		fmt.Printf("  Origin Postal:  %s  %s  %s  [airport: %s, area: %s]\n",
			p.PostalCode, p.StateOrProvinceCode, p.CountryCode, p.AirportID, p.ServiceArea)
	}
	if resp.DestinationPostal != nil {
		p := resp.DestinationPostal
		fmt.Printf("  Dest Postal:    %s  %s  %s  [airport: %s, area: %s]\n",
			p.PostalCode, p.StateOrProvinceCode, p.CountryCode, p.AirportID, p.ServiceArea)
	}

	if len(resp.RateQuotes) > 0 {
		fmt.Printf("\n  📦 Rate Quotes (quoteDate: %s):\n", resp.QuoteDate)
		for _, q := range resp.RateQuotes {
			fmt.Printf("     %-40s  %s %.2f  transit: %s  deliver: %s\n",
				q.ServiceName, q.Currency, q.TotalNetCharge, q.TransitTime, q.DeliveryDay)
		}
	}

	printLocations("Origin", "origin", resp.OriginDropOffLocations)
	printLocations("Destination", "destination", resp.DestinationDropOffLocations)
	fmt.Printf("══════════════════════════════════════════════════════════════\n\n")
}



func printLocations(locType, addressStr string, locs []LocationDetail) {
	if len(locs) == 0 {
		fmt.Printf("  ✗ No drop-off locations found within search radius of %q.\n", addressStr)
		return
	}

	carrier := ""
	if len(locs) > 0 {
		carrier = locs[0].Carrier
	}

	fmt.Printf("  ┌─ %d %s drop-off point(s) near %q\n", len(locs), carrier, addressStr)
	for i, loc := range locs {
		prefix := "  ├"
		if i == len(locs)-1 {
			prefix = "  └"
		}

		addrParts := append([]string{}, loc.StreetLines...)
		if loc.City != "" {
			addrParts = append(addrParts, loc.City)
		}
		if loc.StateOrProvinceCode != "" {
			addrParts = append(addrParts, loc.StateOrProvinceCode)
		}
		if loc.PostalCode != "" {
			addrParts = append(addrParts, loc.PostalCode)
		}
		if loc.CountryCode != "" {
			addrParts = append(addrParts, loc.CountryCode)
		}
		addrStr := strings.Join(addrParts, ", ")

		locTypeSuffix := ""
		if loc.LocationType != "" {
			locTypeSuffix = fmt.Sprintf(" [%s]", loc.LocationType)
		}

		fmt.Printf("%s─ #%d  %s%s\n", prefix, i+1, loc.Name, locTypeSuffix)
		fmt.Printf("  │     Distance: %.2f %s\n", loc.Distance, loc.Units)
		fmt.Printf("  │     Address:  %s\n", addrStr)

		if len(loc.ServiceTypes) > 0 {
			fmt.Printf("  │     Services: %s\n", strings.Join(loc.ServiceTypes, ", "))
		}

		if len(loc.OpeningHours) > 0 {
			fmt.Printf("  │     Hours:    ")
			for j, oh := range loc.OpeningHours {
				if j > 0 {
					fmt.Printf("  │               ")
				}
				fmt.Printf("%-12s  %s – %s\n", oh.DayOfWeek, oh.Opens, oh.Closes)
			}
		}
	}
}


func (s *labelService) GetLabelByTracking(ctx context.Context, trackingNumber string) (*Label, error) {
	if trackingNumber == "" {
		return nil, ErrTrackingNumberRequired
	}
	return s.repo.GetByTracking(ctx, trackingNumber)
}

func (s *labelService) TrackLabel(ctx context.Context, trackingNumber string) (string, error) {
	if trackingNumber == "" {
		return "", ErrTrackingNumberRequired
	}
	label, err := s.repo.GetByTracking(ctx, trackingNumber)
	if err != nil {
		return "", err
	}
	return label.Status, nil
}

func (s *labelService) CancelLabel(ctx context.Context, token, trackingNumber string) error {
	// 1. Authenticate session
	username, err := s.verifyToken(ctx, token)
	if err != nil {
		return err
	}

	if trackingNumber == "" {
		return ErrTrackingNumberRequired
	}

	label, err := s.repo.GetByTracking(ctx, trackingNumber)
	if err != nil {
		return err
	}

	if label.Status == "CANCELLED" {
		return ErrLabelAlreadyCancelled
	}

	// Update local state in SQLite
	label.Status = "CANCELLED"
	if err := s.repo.Update(ctx, label); err != nil {
		return err
	}

	// Synchronous call to Shipment Service to cancel the shipment
	if s.shipmentServiceURL != "" {
		cancelURL := fmt.Sprintf("%s/api/v1/shipments/tracking/%s/cancel", s.shipmentServiceURL, trackingNumber)
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, cancelURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create shipment cancel request: %w", err)
		}

		resp, err := s.httpClient.Do(req)
		if err != nil {
			fmt.Printf("[Warning] Failed to notify Shipment Service of label cancellation: %v\n", err)
		} else {
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				fmt.Printf("[Warning] Shipment Service cancellation returned status %d\n", resp.StatusCode)
			}
		}
	}

	// Log audit entry
	s.logAction(ctx, username, fmt.Sprintf("Cancel Label (Tracking: %s)", trackingNumber))

	return nil
}
