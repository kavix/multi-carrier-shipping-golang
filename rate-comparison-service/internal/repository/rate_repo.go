package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/shipping/rate-comparison-service/internal/domain"
)

type RateRepo struct {
	db *sql.DB
}

func NewRateRepo(db *sql.DB) *RateRepo {
	return &RateRepo{db: db}
}

func (r *RateRepo) Create(ctx context.Context, rc *domain.RateComparison) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO rate_comparisons (id, shipment_id, user_id, from_address, to_address, weight, 
		best_carrier, best_service, best_cost, best_days, all_rates_json, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		rc.ID, rc.ShipmentID, rc.UserID, rc.FromAddress, rc.ToAddress, rc.Weight,
		rc.BestCarrier, rc.BestService, rc.BestCost, rc.BestDays, rc.AllRatesJSON, rc.CreatedAt)
	if err != nil {
		return fmt.Errorf("create rate comparison: %w", err)
	}
	return nil
}

func (r *RateRepo) GetByShipmentID(ctx context.Context, shipmentID string) (*domain.RateComparison, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, shipment_id, user_id, from_address, to_address, weight, 
		best_carrier, best_service, best_cost, best_days, all_rates_json, created_at 
		FROM rate_comparisons WHERE shipment_id = $1`, shipmentID)
	var rc domain.RateComparison
	err := row.Scan(&rc.ID, &rc.ShipmentID, &rc.UserID, &rc.FromAddress, &rc.ToAddress, &rc.Weight,
		&rc.BestCarrier, &rc.BestService, &rc.BestCost, &rc.BestDays, &rc.AllRatesJSON, &rc.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("rate comparison not found")
		}
		return nil, fmt.Errorf("get rate comparison: %w", err)
	}
	return &rc, nil
}
