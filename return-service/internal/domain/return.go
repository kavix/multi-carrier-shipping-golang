package domain

import "time"

type ReturnRequest struct {
	ID              string    `json:"id" db:"id"`
	ShipmentID      string    `json:"shipment_id" db:"shipment_id"`
	UserID          string    `json:"user_id" db:"user_id"`
	Reason          string    `json:"reason" db:"reason"`
	Status          string    `json:"status" db:"status"` // requested, approved, rejected, in_transit, received, refunded
	Carrier         string    `json:"carrier" db:"carrier"`
	ReturnLabelID   string    `json:"return_label_id" db:"return_label_id"`
	RefundAmount    float64   `json:"refund_amount" db:"refund_amount"`
	RefundStatus    string    `json:"refund_status" db:"refund_status"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type ReturnStatus string

const (
	ReturnRequested   ReturnStatus = "requested"
	ReturnApproved    ReturnStatus = "approved"
	ReturnRejected    ReturnStatus = "rejected"
	ReturnInTransit   ReturnStatus = "in_transit"
	ReturnReceived    ReturnStatus = "received"
	ReturnRefunded    ReturnStatus = "refunded"
)
