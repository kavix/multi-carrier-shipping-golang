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
) (*Label, error) {
	// Validation
	if carrier == "" {
		return nil, ErrCarrierRequired
	}
	if weight <= 0 {
		return nil, ErrInvalidWeight
	}

	var label *Label
	var err error

	if s.carrierSvc != nil {
		label, err = s.carrierSvc.GenerateLabel(ctx, shipmentID, carrier, weight, origin, destination)
		if err != nil {
			return nil, fmt.Errorf("failed to generate label via carrier: %w", err)
		}
	} else {
		// Mock fallback
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

	// Persist label
	if err := s.repo.Create(ctx, label); err != nil {
		return nil, err
	}

	// Query FedEx locations if carrier is FedEx
	if strings.EqualFold(carrier, "FedEx") && s.carrierSvc != nil {
		fmt.Printf("\n=== QUERYING FEDEX SANDBOX LOCATIONS FOR SHIPMENT LABEL %s ===\n", label.TrackingNumber)

		// Query Origin
		fmt.Printf("[Origin Location Search] Querying: %s\n", origin)
		originLocs, err := s.carrierSvc.SearchLocations(ctx, origin)
		if err != nil {
			fmt.Printf("[Origin Error] Failed to retrieve FedEx locations: %v\n", err)
		} else {
			printLocations("Origin", origin, originLocs)
		}

		// Query Destination
		fmt.Printf("[Destination Location Search] Querying: %s\n", destination)
		destLocs, err := s.carrierSvc.SearchLocations(ctx, destination)
		if err != nil {
			fmt.Printf("[Destination Error] Failed to retrieve FedEx locations: %v\n", err)
		} else {
			printLocations("Destination", destination, destLocs)
		}
		fmt.Println("==========================================================================")
	}

	return label, nil
}

func printLocations(locType, addressStr string, locs []LocationDetail) {
	if len(locs) == 0 {
		fmt.Printf("No FedEx dropoff locations found within the search radius of %s.\n\n", addressStr)
		return
	}

	fmt.Printf("Nearest FedEx Drop-off Locations for %s (%s):\n", locType, addressStr)
	for i, loc := range locs {
		addrStr := strings.Join(loc.StreetLines, ", ")
		if loc.City != "" {
			addrStr += fmt.Sprintf(", %s", loc.City)
		}
		if loc.StateOrProvinceCode != "" {
			addrStr += fmt.Sprintf(", %s", loc.StateOrProvinceCode)
		}
		if loc.PostalCode != "" {
			addrStr += fmt.Sprintf(" %s", loc.PostalCode)
		}
		if loc.CountryCode != "" {
			addrStr += fmt.Sprintf(", %s", loc.CountryCode)
		}

		fmt.Printf("  %d. Name:     %s\n", i+1, loc.Name)
		fmt.Printf("     Distance: %.2f %s (NEAREST)\n", loc.Distance, loc.Units)
		fmt.Printf("     Address:  %s\n", addrStr)
	}
	fmt.Println()
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
