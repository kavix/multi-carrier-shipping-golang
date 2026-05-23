package label

import (
	"context"
	"errors"
	"time"
)

// Label represents a shipping label entity.
type Label struct {
	ID             string    `json:"id"`
	ShipmentID     string    `json:"shipment_id"`
	TrackingNumber string    `json:"tracking_number"`
	LabelURL       string    `json:"label_url"`
	Status         string    `json:"status"` // "ACTIVE", "CANCELLED"
	CreatedAt      time.Time `json:"created_at"`
}

// LocationDetail represents parsed FedEx drop-off locations.
type LocationDetail struct {
	Distance            float64  `json:"distance"`
	Units               string   `json:"units"`
	Name                string   `json:"name"`
	StreetLines         []string `json:"street_lines"`
	City                string   `json:"city"`
	StateOrProvinceCode string   `json:"state_or_province_code"`
	PostalCode          string   `json:"postal_code"`
	CountryCode         string   `json:"country_code"`
}

// Sentinel errors representing domain-specific failures.
var (
	ErrLabelNotFound         = errors.New("label not found")
	ErrLabelAlreadyCancelled = errors.New("label is already cancelled")
	ErrCarrierRequired       = errors.New("carrier is required")
	ErrTrackingNumberRequired = errors.New("tracking number is required")
	ErrInvalidWeight         = errors.New("shipment weight must be greater than zero")
)

// LabelRepository defines DB access interface for Labels.
type LabelRepository interface {
	Create(ctx context.Context, label *Label) error
	GetByID(ctx context.Context, id string) (*Label, error)
	GetByTracking(ctx context.Context, trackingNum string) (*Label, error)
	Update(ctx context.Context, label *Label) error
	Delete(ctx context.Context, id string) error
}

// CarrierService defines the integration client with external carriers like FedEx.
type CarrierService interface {
	SearchLocations(ctx context.Context, addressStr string) ([]LocationDetail, error)
	GenerateLabel(ctx context.Context, shipmentID, carrier string, weight float64, origin, destination string) (*Label, error)
}

// LabelService defines business operations for Labels.
type LabelService interface {
	CreateLabel(ctx context.Context, shipmentID, carrier string, weight float64, origin, destination string) (*Label, error)
	GetLabelByTracking(ctx context.Context, trackingNumber string) (*Label, error)
	TrackLabel(ctx context.Context, trackingNumber string) (string, error)
	CancelLabel(ctx context.Context, token, trackingNumber string) error
}
