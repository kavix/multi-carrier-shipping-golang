package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/shipping/return-service/internal/domain"
)

type ReturnRepo struct {
	db *sql.DB
}

func NewReturnRepo(db *sql.DB) *ReturnRepo {
	return &ReturnRepo{db: db}
}

func (r *ReturnRepo) Create(ctx context.Context, ret *domain.ReturnRequest) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO return_requests (id, shipment_id, user_id, reason, status, carrier, return_label_id, refund_amount, refund_status, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		ret.ID, ret.ShipmentID, ret.UserID, ret.Reason, ret.Status, ret.Carrier, ret.ReturnLabelID, ret.RefundAmount, ret.RefundStatus, ret.CreatedAt, ret.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create return: %w", err)
	}
	return nil
}

func (r *ReturnRepo) GetByID(ctx context.Context, id string) (*domain.ReturnRequest, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, shipment_id, user_id, reason, status, carrier, return_label_id, refund_amount, refund_status, created_at, updated_at 
		FROM return_requests WHERE id = $1`, id)
	var ret domain.ReturnRequest
	err := row.Scan(&ret.ID, &ret.ShipmentID, &ret.UserID, &ret.Reason, &ret.Status, &ret.Carrier, &ret.ReturnLabelID, &ret.RefundAmount, &ret.RefundStatus, &ret.CreatedAt, &ret.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("return not found")
		}
		return nil, fmt.Errorf("get return: %w", err)
	}
	return &ret, nil
}

func (r *ReturnRepo) GetByShipmentID(ctx context.Context, shipmentID string) ([]*domain.ReturnRequest, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, shipment_id, user_id, reason, status, carrier, return_label_id, refund_amount, refund_status, created_at, updated_at 
		FROM return_requests WHERE shipment_id = $1 ORDER BY created_at DESC`, shipmentID)
	if err != nil {
		return nil, fmt.Errorf("list returns: %w", err)
	}
	defer rows.Close()

	var returns []*domain.ReturnRequest
	for rows.Next() {
		var ret domain.ReturnRequest
		if err := rows.Scan(&ret.ID, &ret.ShipmentID, &ret.UserID, &ret.Reason, &ret.Status, &ret.Carrier, &ret.ReturnLabelID, &ret.RefundAmount, &ret.RefundStatus, &ret.CreatedAt, &ret.UpdatedAt); err != nil {
			continue
		}
		returns = append(returns, &ret)
	}
	return returns, nil
}

func (r *ReturnRepo) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE return_requests SET status=$1, updated_at=NOW() WHERE id=$2`, status, id)
	if err != nil {
		return fmt.Errorf("update return status: %w", err)
	}
	return nil
}

func (r *ReturnRepo) UpdateRefund(ctx context.Context, id string, amount float64, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE return_requests SET refund_amount=$1, refund_status=$2, updated_at=NOW() WHERE id=$3`, amount, status, id)
	if err != nil {
		return fmt.Errorf("update refund: %w", err)
	}
	return nil
}
