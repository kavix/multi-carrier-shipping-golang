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

// OpeningHours represents the daily operating schedule for a location.
type OpeningHours struct {
	DayOfWeek string `json:"day_of_week"`
	Opens     string `json:"opens"`
	Closes    string `json:"closes"`
}

// LocationDetail represents a carrier drop-off / pick-up location (unified across carriers).
type LocationDetail struct {
	// Common fields
	Carrier             string   `json:"carrier"`
	LocationType        string   `json:"location_type,omitempty"` // e.g. "locker", "postoffice", "servicepoint"
	Distance            float64  `json:"distance"`
	Units               string   `json:"units"`
	Name                string   `json:"name"`
	StreetLines         []string `json:"street_lines"`
	City                string   `json:"city"`
	StateOrProvinceCode string   `json:"state_or_province_code,omitempty"`
	PostalCode          string   `json:"postal_code,omitempty"`
	CountryCode         string   `json:"country_code"`

	// Enriched fields (populated when available from carrier)
	OpeningHours []OpeningHours `json:"opening_hours,omitempty"`
	ServiceTypes []string       `json:"service_types,omitempty"`
}

// Surcharge represents a single surcharge or fee on a rate quote.
type Surcharge struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
}

// RateQuote holds a single service-level rate and transit time from FedEx.
type RateQuote struct {
	ServiceType        string     `json:"service_type"`
	ServiceName        string     `json:"service_name"`
	Currency           string     `json:"currency"`
	BaseCharge         float64    `json:"base_charge"`
	TotalNetCharge     float64    `json:"total_net_charge"`
	FuelSurcharge      float64    `json:"fuel_surcharge_percent"`
	TotalSurcharges    float64    `json:"total_surcharges"`
	Surcharges         []Surcharge `json:"surcharges,omitempty"`
	TransitTime        string     `json:"transit_time,omitempty"`       // e.g. "TWO_DAYS"
	DeliveryDay        string     `json:"delivery_day,omitempty"`       // e.g. "MON"
	CommitDateTime     string     `json:"commit_date_time,omitempty"`   // e.g. "2025-06-02T08:30:00"
	DeliveryPostalCode string     `json:"delivery_postal_code,omitempty"`
	RateZone           string     `json:"rate_zone,omitempty"`
}

// PostalInfo contains validated postal code details from FedEx.
type PostalInfo struct {
	PostalCode          string `json:"postal_code"`
	CityFirstInitials   string `json:"city_first_initials,omitempty"`
	StateOrProvinceCode string `json:"state_or_province_code,omitempty"`
	CountryCode         string `json:"country_code"`
	AirportID           string `json:"airport_id,omitempty"`
	ServiceArea         string `json:"service_area,omitempty"`
	LocationID          string `json:"location_id,omitempty"`
}

// LabelCreateResponse is the rich structured API response returned when a shipment label is created.
type LabelCreateResponse struct {
	// Core label
	Label *Label `json:"label"`

	// Carrier-enriched data (FedEx-specific)
	Carrier     string `json:"carrier"`
	QuoteDate   string `json:"quote_date,omitempty"`

	// Origin/destination postal validation
	OriginPostal      *PostalInfo `json:"origin_postal,omitempty"`
	DestinationPostal *PostalInfo `json:"destination_postal,omitempty"`

	// Available service rates sorted by net charge (cheapest first)
	RateQuotes []RateQuote `json:"rate_quotes,omitempty"`

	// Drop-off locations near origin
	OriginDropOffLocations      []LocationDetail `json:"origin_drop_off_locations,omitempty"`
	// Drop-off locations near destination
	DestinationDropOffLocations []LocationDetail `json:"destination_drop_off_locations,omitempty"`
}

// Sentinel errors representing domain-specific failures.
var (
	ErrLabelNotFound          = errors.New("label not found")
	ErrLabelAlreadyCancelled  = errors.New("label is already cancelled")
	ErrCarrierRequired        = errors.New("carrier is required")
	ErrTrackingNumberRequired = errors.New("tracking number is required")
	ErrInvalidWeight          = errors.New("shipment weight must be greater than zero")
	ErrUnsupportedCarrier     = errors.New("unsupported carrier")
)

// LabelRepository defines DB access interface for Labels.
type LabelRepository interface {
	Create(ctx context.Context, label *Label) error
	GetByID(ctx context.Context, id string) (*Label, error)
	GetByTracking(ctx context.Context, trackingNum string) (*Label, error)
	Update(ctx context.Context, label *Label) error
	Delete(ctx context.Context, id string) error
}

// CarrierService defines the integration client with external carriers.
// The carrier argument allows a single MultiCarrierClient to dispatch to the correct backend.
type CarrierService interface {
	SearchLocations(ctx context.Context, carrier, addressStr string) ([]LocationDetail, error)
	GenerateLabel(ctx context.Context, shipmentID, carrier string, weight float64, origin, destination string) (*Label, error)
}

// LabelService defines business operations for Labels.
type LabelService interface {
	// CreateLabel generates a shipping label and returns a rich structured response
	// containing rate quotes, postal validation, and nearby drop-off locations.
	CreateLabel(ctx context.Context, shipmentID, carrier string, weight float64, origin, destination string) (*LabelCreateResponse, error)
	GetLabelByTracking(ctx context.Context, trackingNumber string) (*Label, error)
	TrackLabel(ctx context.Context, trackingNumber string) (string, error)
	CancelLabel(ctx context.Context, token, trackingNumber string) error
}

