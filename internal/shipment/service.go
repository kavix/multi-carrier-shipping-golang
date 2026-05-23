package shipment

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type shipmentService struct {
	repo            ShipmentRepository
	labelServiceURL string
	httpClient      *http.Client
}

// NewShipmentService instantiates a new ShipmentService implementation.
func NewShipmentService(repo ShipmentRepository, labelServiceURL string) ShipmentService {
	return &shipmentService{
		repo:            repo,
		labelServiceURL: strings.TrimSuffix(labelServiceURL, "/"),
		httpClient:      &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *shipmentService) CreateShipment(
	ctx context.Context,
	carrier string,
	weight float64,
	origin, destination string,
) (*Shipment, *Label, error) {
	// 1. Validation
	if carrier == "" {
		return nil, nil, ErrCarrierRequired
	}
	if weight <= 0 {
		return nil, nil, ErrInvalidWeight
	}

	// 2. Construct Shipment with ID
	id, err := generateUUID()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate shipment id: %w", err)
	}

	now := time.Now()
	shipment := &Shipment{
		ID:             id,
		Carrier:        carrier,
		TrackingNumber: "PENDING",
		Weight:         weight,
		Origin:         origin,
		Destination:    destination,
		Status:         "PENDING",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Persist the initial shipment in Shipment DB
	if err := s.repo.Create(ctx, shipment); err != nil {
		return nil, nil, err
	}

	// 3. Make HTTP request to Label Service to generate label
	labelReqBody := struct {
		ShipmentID  string  `json:"shipment_id"`
		Carrier     string  `json:"carrier"`
		Weight      float64 `json:"weight"`
		Origin      string  `json:"origin"`
		Destination string  `json:"destination"`
	}{
		ShipmentID:  shipment.ID,
		Carrier:     shipment.Carrier,
		Weight:      shipment.Weight,
		Origin:      shipment.Origin,
		Destination: shipment.Destination,
	}

	jsonBytes, err := json.Marshal(labelReqBody)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal label request: %w", err)
	}

	labelURL := fmt.Sprintf("%s/api/v1/labels", s.labelServiceURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, labelURL, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create label request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("label service is currently unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, nil, fmt.Errorf("label service returned error status %d", resp.StatusCode)
	}

	var label Label
	if err := json.NewDecoder(resp.Body).Decode(&label); err != nil {
		return nil, nil, fmt.Errorf("failed to decode label details: %w", err)
	}

	// 4. Update shipment with generated tracking number
	shipment.TrackingNumber = label.TrackingNumber
	shipment.Status = "CREATED"
	shipment.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, shipment); err != nil {
		return nil, nil, err
	}

	return shipment, &label, nil
}

func (s *shipmentService) GetShipment(ctx context.Context, id string) (*Shipment, error) {
	if id == "" {
		return nil, ErrShipmentNotFound
	}
	return s.repo.GetByID(ctx, id)
}

func (s *shipmentService) ListShipments(ctx context.Context) ([]*Shipment, error) {
	return s.repo.List(ctx)
}

func (s *shipmentService) CancelShipmentByTracking(ctx context.Context, trackingNumber string) error {
	if trackingNumber == "" {
		return ErrShipmentNotFound
	}

	shipment, err := s.repo.GetByTracking(ctx, trackingNumber)
	if err != nil {
		return err
	}

	shipment.Status = "CANCELLED"
	shipment.UpdatedAt = time.Now()

	return s.repo.Update(ctx, shipment)
}

// generateUUID creates a standard secure random UUID
func generateUUID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant is 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}
