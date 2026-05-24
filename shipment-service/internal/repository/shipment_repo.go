package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/shipping/shipment-service/internal/domain"
)

type ShipmentRepo struct {
	db *sql.DB
}

func NewShipmentRepo(db *sql.DB) *ShipmentRepo {
	return &ShipmentRepo{db: db}
}

func (r *ShipmentRepo) Create(ctx context.Context, s *domain.Shipment) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO shipments (id, user_id, sender_name, sender_address, receiver_name, receiver_address, 
		weight, dimensions, carrier, service_type, status, tracking_number, cost, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		s.ID, s.UserID, s.SenderName, s.SenderAddress, s.ReceiverName, s.ReceiverAddress,
		s.Weight, s.Dimensions, s.Carrier, s.ServiceType, s.Status, s.TrackingNumber, s.Cost, s.CreatedAt, s.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create shipment: %w", err)
	}
	return nil
}

func (r *ShipmentRepo) GetByID(ctx context.Context, id string) (*domain.Shipment, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, sender_name, sender_address, receiver_name, receiver_address, 
		weight, dimensions, carrier, service_type, status, tracking_number, cost, created_at, updated_at 
		FROM shipments WHERE id = $1`, id)
	var s domain.Shipment
	err := row.Scan(&s.ID, &s.UserID, &s.SenderName, &s.SenderAddress, &s.ReceiverName, &s.ReceiverAddress,
		&s.Weight, &s.Dimensions, &s.Carrier, &s.ServiceType, &s.Status, &s.TrackingNumber, &s.Cost, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("shipment not found")
		}
		return nil, fmt.Errorf("get shipment: %w", err)
	}
	return &s, nil
}

func (r *ShipmentRepo) GetByUserID(ctx context.Context, userID string) ([]*domain.Shipment, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, sender_name, sender_address, receiver_name, receiver_address, 
		weight, dimensions, carrier, service_type, status, tracking_number, cost, created_at, updated_at 
		FROM shipments WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list shipments: %w", err)
	}
	defer rows.Close()

	var shipments []*domain.Shipment
	for rows.Next() {
		var s domain.Shipment
		if err := rows.Scan(&s.ID, &s.UserID, &s.SenderName, &s.SenderAddress, &s.ReceiverName, &s.ReceiverAddress,
			&s.Weight, &s.Dimensions, &s.Carrier, &s.ServiceType, &s.Status, &s.TrackingNumber, &s.Cost, &s.CreatedAt, &s.UpdatedAt); err != nil {
			continue
		}
		shipments = append(shipments, &s)
	}
	return shipments, nil
}

func (r *ShipmentRepo) Update(ctx context.Context, s *domain.Shipment) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE shipments SET sender_name=$1, sender_address=$2, receiver_name=$3, receiver_address=$4,
		weight=$5, dimensions=$6, carrier=$7, service_type=$8, status=$9, tracking_number=$10, cost=$11, updated_at=$12 
		WHERE id=$13`,
		s.SenderName, s.SenderAddress, s.ReceiverName, s.ReceiverAddress,
		s.Weight, s.Dimensions, s.Carrier, s.ServiceType, s.Status, s.TrackingNumber, s.Cost, s.UpdatedAt, s.ID)
	if err != nil {
		return fmt.Errorf("update shipment: %w", err)
	}
	return nil
}

func (r *ShipmentRepo) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE shipments SET status=$1, updated_at=NOW() WHERE id=$2`, status, id)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	return nil
}

func (r *ShipmentRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM shipments WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete shipment: %w", err)
	}
	return nil
}
