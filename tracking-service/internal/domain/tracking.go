package domain

import "time"

type TrackingEvent struct {
	ID             string    `json:"id" db:"id"`
	ShipmentID     string    `json:"shipment_id" db:"shipment_id"`
	TrackingNumber string    `json:"tracking_number" db:"tracking_number"`
	Carrier        string    `json:"carrier" db:"carrier"`
	Status         string    `json:"status" db:"status"`
	Location       string    `json:"location" db:"location"`
	Description    string    `json:"description" db:"description"`
	Timestamp      time.Time `json:"timestamp" db:"timestamp"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

type TrackingHistory struct {
	ShipmentID     string           `json:"shipment_id"`
	TrackingNumber string           `json:"tracking_number"`
	Carrier        string           `json:"carrier"`
	CurrentStatus  string           `json:"current_status"`
	Events         []TrackingEvent  `json:"events"`
}
