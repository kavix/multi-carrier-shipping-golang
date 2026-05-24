package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/shipping/tracking-service/internal/domain"
)

type TrackingRepo struct {
	db *sql.DB
}

func NewTrackingRepo(db *sql.DB) *TrackingRepo {
	return &TrackingRepo{db: db}
}

func (r *TrackingRepo) Create(ctx context.Context, t *domain.TrackingEvent) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO tracking_events (id, shipment_id, tracking_number, carrier, status, location, description, timestamp, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		t.ID, t.ShipmentID, t.TrackingNumber, t.Carrier, t.Status, t.Location, t.Description, t.Timestamp, t.CreatedAt)
	if err != nil {
		return fmt.Errorf("create tracking event: %w", err)
	}
	return nil
}

func (r *TrackingRepo) GetByShipmentID(ctx context.Context, shipmentID string) ([]*domain.TrackingEvent, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, shipment_id, tracking_number, carrier, status, location, description, timestamp, created_at 
		FROM tracking_events WHERE shipment_id = $1 ORDER BY timestamp DESC`, shipmentID)
	if err != nil {
		return nil, fmt.Errorf("list tracking: %w", err)
	}
	defer rows.Close()

	var events []*domain.TrackingEvent
	for rows.Next() {
		var t domain.TrackingEvent
		if err := rows.Scan(&t.ID, &t.ShipmentID, &t.TrackingNumber, &t.Carrier, &t.Status, &t.Location, &t.Description, &t.Timestamp, &t.CreatedAt); err != nil {
			continue
		}
		events = append(events, &t)
	}
	return events, nil
}

func (r *TrackingRepo) GetLatestByShipmentID(ctx context.Context, shipmentID string) (*domain.TrackingEvent, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, shipment_id, tracking_number, carrier, status, location, description, timestamp, created_at 
		FROM tracking_events WHERE shipment_id = $1 ORDER BY timestamp DESC LIMIT 1`, shipmentID)
	var t domain.TrackingEvent
	err := row.Scan(&t.ID, &t.ShipmentID, &t.TrackingNumber, &t.Carrier, &t.Status, &t.Location, &t.Description, &t.Timestamp, &t.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no tracking found")
		}
		return nil, fmt.Errorf("get latest tracking: %w", err)
	}
	return &t, nil
}
