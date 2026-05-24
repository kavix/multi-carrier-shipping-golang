package domain

import "time"

type RateComparison struct {
	ID            string    `json:"id" db:"id"`
	ShipmentID    string    `json:"shipment_id" db:"shipment_id"`
	UserID        string    `json:"user_id" db:"user_id"`
	FromAddress   string    `json:"from_address" db:"from_address"`
	ToAddress     string    `json:"to_address" db:"to_address"`
	Weight        float64   `json:"weight" db:"weight"`
	BestCarrier   string    `json:"best_carrier" db:"best_carrier"`
	BestService   string    `json:"best_service" db:"best_service"`
	BestCost      float64   `json:"best_cost" db:"best_cost"`
	BestDays      int       `json:"best_days" db:"best_days"`
	AllRatesJSON  string    `json:"-" db:"all_rates_json"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type RateResult struct {
	CarrierID     string  `json:"carrier_id"`
	CarrierName   string  `json:"carrier_name"`
	ServiceType   string  `json:"service_type"`
	Cost          float64 `json:"cost"`
	Currency      string  `json:"currency"`
	EstimatedDays int     `json:"estimated_days"`
	Rating        float64 `json:"rating"`
}
