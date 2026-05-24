package domain

import "time"

type ValidatedAddress struct {
	ID          string    `json:"id" db:"id"`
	RawAddress  string    `json:"raw_address" db:"raw_address"`
	Street      string    `json:"street" db:"street"`
	City        string    `json:"city" db:"city"`
	State       string    `json:"state" db:"state"`
	PostalCode  string    `json:"postal_code" db:"postal_code"`
	Country     string    `json:"country" db:"country"`
	Latitude    float64   `json:"latitude" db:"latitude"`
	Longitude   float64   `json:"longitude" db:"longitude"`
	IsValid     bool      `json:"is_valid" db:"is_valid"`
	ValidatedAt time.Time `json:"validated_at" db:"validated_at"`
}

type Location struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Address    string  `json:"address"`
	City       string  `json:"city"`
	Country    string  `json:"country"`
	PostalCode string  `json:"postal_code"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Type       string  `json:"type"` // "pickup" or "drop"
	DistanceKm float64 `json:"distance_km"`
}
