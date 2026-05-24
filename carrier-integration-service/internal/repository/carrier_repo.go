package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/shipping/carrier-integration-service/internal/domain"
)

type CarrierRepo struct {
	db *sql.DB
}

func NewCarrierRepo(db *sql.DB) *CarrierRepo {
	return &CarrierRepo{db: db}
}

func (r *CarrierRepo) Create(ctx context.Context, c *domain.Carrier) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO carriers (id, name, code, api_key, api_secret, base_url, is_active, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		c.ID, c.Name, c.Code, c.APIKey, c.APISecret, c.BaseURL, c.IsActive, c.CreatedAt)
	if err != nil {
		return fmt.Errorf("create carrier: %w", err)
	}
	return nil
}

func (r *CarrierRepo) GetByCode(ctx context.Context, code string) (*domain.Carrier, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, code, api_key, api_secret, base_url, is_active, created_at FROM carriers WHERE code = $1`, code)
	var c domain.Carrier
	err := row.Scan(&c.ID, &c.Name, &c.Code, &c.APIKey, &c.APISecret, &c.BaseURL, &c.IsActive, &c.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("carrier not found")
		}
		return nil, fmt.Errorf("get carrier: %w", err)
	}
	return &c, nil
}

func (r *CarrierRepo) ListActive(ctx context.Context) ([]*domain.Carrier, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, code, api_key, api_secret, base_url, is_active, created_at FROM carriers WHERE is_active = true`)
	if err != nil {
		return nil, fmt.Errorf("list carriers: %w", err)
	}
	defer rows.Close()

	var carriers []*domain.Carrier
	for rows.Next() {
		var c domain.Carrier
		if err := rows.Scan(&c.ID, &c.Name, &c.Code, &c.APIKey, &c.APISecret, &c.BaseURL, &c.IsActive, &c.CreatedAt); err != nil {
			continue
		}
		carriers = append(carriers, &c)
	}
	return carriers, nil
}
