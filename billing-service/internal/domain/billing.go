package domain

import "time"

type Invoice struct {
	ID          string    `json:"id" db:"id"`
	ShipmentID  string    `json:"shipment_id" db:"shipment_id"`
	UserID      string    `json:"user_id" db:"user_id"`
	Amount      float64   `json:"amount" db:"amount"`
	Currency    string    `json:"currency" db:"currency"`
	Status      string    `json:"status" db:"status"` // pending, paid, failed, refunded
	Description string    `json:"description" db:"description"`
	StripeID    string    `json:"-" db:"stripe_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	PaidAt      *time.Time `json:"paid_at,omitempty" db:"paid_at"`
}

type Payment struct {
	ID        string    `json:"id" db:"id"`
	InvoiceID string    `json:"invoice_id" db:"invoice_id"`
	Amount    float64   `json:"amount" db:"amount"`
	Currency  string    `json:"currency" db:"currency"`
	Status    string    `json:"status" db:"status"`
	Method    string    `json:"method" db:"method"` // stripe, paypal, etc.
	StripeID  string    `json:"-" db:"stripe_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
