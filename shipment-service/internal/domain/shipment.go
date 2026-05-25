package domain

import "time"

type Shipment struct {
	ID               string    `json:"id" db:"id"`
	UserID           string    `json:"user_id" db:"user_id"`
	SenderName       string    `json:"sender_name" db:"sender_name"`
	SenderAddress    string    `json:"sender_address" db:"sender_address"`
	SenderEmail      string    `json:"sender_email" db:"sender_email"`
	ReceiverName     string    `json:"receiver_name" db:"receiver_name"`
	ReceiverAddress  string    `json:"receiver_address" db:"receiver_address"`
	ReceiverEmail    string    `json:"receiver_email" db:"receiver_email"`
	Weight           float64   `json:"weight" db:"weight"`
	Dimensions       string    `json:"dimensions" db:"dimensions"`
	Carrier          string    `json:"carrier" db:"carrier"`
	ServiceType      string    `json:"service_type" db:"service_type"`
	Status           string    `json:"status" db:"status"`
	TrackingNumber   string    `json:"tracking_number" db:"tracking_number"`
	LabelID          string    `json:"label_id" db:"label_id"`
	LabelURL         string    `json:"label_url" db:"label_url"`
	Cost             float64   `json:"cost" db:"cost"`
	PickupLocationID string    `json:"pickup_location_id" db:"pickup_location_id"`
	DropLocationID   string    `json:"drop_location_id" db:"drop_location_id"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

type ShipmentStatus string

const (
	StatusPending   ShipmentStatus = "pending"
	StatusCreated   ShipmentStatus = "created"
	StatusPickedUp  ShipmentStatus = "picked_up"
	StatusInTransit ShipmentStatus = "in_transit"
	StatusDelivered ShipmentStatus = "delivered"
	StatusFailed    ShipmentStatus = "failed"
	StatusReturned  ShipmentStatus = "returned"
)
