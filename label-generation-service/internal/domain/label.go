package domain

import "time"

type ShippingLabel struct {
	ID             string    `json:"id" db:"id"`
	ShipmentID     string    `json:"shipment_id" db:"shipment_id"`
	Carrier        string    `json:"carrier" db:"carrier"`
	TrackingNumber string    `json:"tracking_number" db:"tracking_number"`
	LabelData      string    `json:"label_data" db:"label_data"` // Base64 encoded PDF
	LabelURL       string    `json:"label_url" db:"label_url"`
	Format         string    `json:"format" db:"format"` // PDF, ZPL, PNG
	Status         string    `json:"status" db:"status"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}
