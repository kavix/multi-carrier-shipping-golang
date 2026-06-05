package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"github.com/shipping/billing-service/internal/domain"
)

type BillingRepo struct {
	db *sql.DB
}

func NewBillingRepo(db *sql.DB) *BillingRepo {
	return &BillingRepo{db: db}
}

func (r *BillingRepo) CreateInvoice(ctx context.Context, inv *domain.Invoice) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO invoices (id, shipment_id, user_id, amount, currency, status, description, stripe_id, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		inv.ID, inv.ShipmentID, inv.UserID, inv.Amount, inv.Currency, inv.Status, inv.Description, inv.StripeID, inv.CreatedAt)
	if err != nil {
		return fmt.Errorf("create invoice: %w", err)
	}
	return nil
}

func (r *BillingRepo) GetInvoiceByID(ctx context.Context, id string) (*domain.Invoice, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, shipment_id, user_id, amount, currency, status, description, stripe_id, created_at, paid_at 
		FROM invoices WHERE id = $1`, id)
	var inv domain.Invoice
	var paidAt sql.NullTime
	err := row.Scan(&inv.ID, &inv.ShipmentID, &inv.UserID, &inv.Amount, &inv.Currency, &inv.Status, &inv.Description, &inv.StripeID, &inv.CreatedAt, &paidAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invoice not found")
		}
		return nil, fmt.Errorf("get invoice: %w", err)
	}
	if paidAt.Valid {
		inv.PaidAt = &paidAt.Time
	}
	return &inv, nil
}

func (r *BillingRepo) GetInvoiceByShipmentID(ctx context.Context, shipmentID string) (*domain.Invoice, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, shipment_id, user_id, amount, currency, status, description, stripe_id, created_at, paid_at 
		FROM invoices WHERE shipment_id = $1`, shipmentID)
	var inv domain.Invoice
	var paidAt sql.NullTime
	err := row.Scan(&inv.ID, &inv.ShipmentID, &inv.UserID, &inv.Amount, &inv.Currency, &inv.Status, &inv.Description, &inv.StripeID, &inv.CreatedAt, &paidAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invoice not found")
		}
		return nil, fmt.Errorf("get invoice: %w", err)
	}
	if paidAt.Valid {
		inv.PaidAt = &paidAt.Time
	}
	return &inv, nil
}

func (r *BillingRepo) UpdateInvoiceStatus(ctx context.Context, id, status, stripeID string) error {
	var paidAt interface{}
	if status == "paid" {
		paidAt = time.Now()
	} else {
		paidAt = nil
	}
	_, err := r.db.ExecContext(ctx,
		`UPDATE invoices SET status=$1, paid_at=$2, stripe_id=$3 WHERE id=$4`, status, paidAt, stripeID, id)
	if err != nil {
		return fmt.Errorf("update invoice: %w", err)
	}
	return nil
}

func (r *BillingRepo) CreatePayment(ctx context.Context, p *domain.Payment) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO payments (id, invoice_id, amount, currency, status, method, stripe_id, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		p.ID, p.InvoiceID, p.Amount, p.Currency, p.Status, p.Method, p.StripeID, p.CreatedAt)
	if err != nil {
		return fmt.Errorf("create payment: %w", err)
	}
	return nil
}
