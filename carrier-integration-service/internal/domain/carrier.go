package domain

import "time"

type Carrier struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Code      string    `json:"code" db:"code"`
	APIKey    string    `json:"-" db:"api_key"`
	APISecret string    `json:"-" db:"api_secret"`
	BaseURL   string    `json:"base_url" db:"base_url"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CarrierRate struct {
	CarrierID       string  `json:"carrier_id"`
	CarrierName     string  `json:"carrier_name"`
	ServiceType     string  `json:"service_type"`
	EstimatedDays   int     `json:"estimated_days"`
	Cost            float64 `json:"cost"`
	Currency        string  `json:"currency"`
}

type TrackingInfo struct {
	TrackingNumber string    `json:"tracking_number"`
	Carrier        string    `json:"carrier"`
	Status         string    `json:"status"`
	Location       string    `json:"location"`
	Timestamp      time.Time `json:"timestamp"`
	Description    string    `json:"description"`
}

type PickupDropLocation struct {
	ID          string  `json:"id"`
	Carrier     string  `json:"carrier"`
	Name        string  `json:"name"`
	Address     string  `json:"address"`
	City        string  `json:"city"`
	Country     string  `json:"country"`
	PostalCode  string  `json:"postal_code"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Type        string  `json:"type"` // "pickup" or "drop"
	DistanceKm  float64 `json:"distance_km"`
}
