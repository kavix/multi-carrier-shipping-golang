package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/shipping/label-generation-service/internal/domain"
)

type LabelRepo struct {
	db *sql.DB
}

func NewLabelRepo(db *sql.DB) *LabelRepo {
	return &LabelRepo{db: db}
}

func (r *LabelRepo) Create(ctx context.Context, l *domain.ShippingLabel) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO labels (id, shipment_id, carrier, tracking_number, label_data, label_url, format, status, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		l.ID, l.ShipmentID, l.Carrier, l.TrackingNumber, l.LabelData, l.LabelURL, l.Format, l.Status, l.CreatedAt)
	if err != nil {
		return fmt.Errorf("create label: %w", err)
	}
	return nil
}

func (r *LabelRepo) GetByShipmentID(ctx context.Context, shipmentID string) (*domain.ShippingLabel, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, shipment_id, carrier, tracking_number, label_data, label_url, format, status, created_at 
		FROM labels WHERE shipment_id = $1`, shipmentID)
	var l domain.ShippingLabel
	err := row.Scan(&l.ID, &l.ShipmentID, &l.Carrier, &l.TrackingNumber, &l.LabelData, &l.LabelURL, &l.Format, &l.Status, &l.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("label not found")
		}
		return nil, fmt.Errorf("get label: %w", err)
	}
	return &l, nil
}
