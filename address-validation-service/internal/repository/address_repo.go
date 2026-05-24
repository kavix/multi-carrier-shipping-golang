package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/shipping/address-validation-service/internal/domain"
)

type AddressRepo struct {
	db *sql.DB
}

func NewAddressRepo(db *sql.DB) *AddressRepo {
	return &AddressRepo{db: db}
}

func (r *AddressRepo) Save(ctx context.Context, a *domain.ValidatedAddress) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO validated_addresses (id, raw_address, street, city, state, postal_code, country, latitude, longitude, is_valid, validated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (raw_address) DO UPDATE SET
		street = EXCLUDED.street, city = EXCLUDED.city, state = EXCLUDED.state,
		postal_code = EXCLUDED.postal_code, country = EXCLUDED.country,
		latitude = EXCLUDED.latitude, longitude = EXCLUDED.longitude,
		is_valid = EXCLUDED.is_valid, validated_at = EXCLUDED.validated_at`,
		a.ID, a.RawAddress, a.Street, a.City, a.State, a.PostalCode, a.Country, a.Latitude, a.Longitude, a.IsValid, a.ValidatedAt)
	if err != nil {
		return fmt.Errorf("save address: %w", err)
	}
	return nil
}

func (r *AddressRepo) GetByRawAddress(ctx context.Context, rawAddress string) (*domain.ValidatedAddress, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, raw_address, street, city, state, postal_code, country, latitude, longitude, is_valid, validated_at 
		FROM validated_addresses WHERE raw_address = $1`, rawAddress)
	var a domain.ValidatedAddress
	err := row.Scan(&a.ID, &a.RawAddress, &a.Street, &a.City, &a.State, &a.PostalCode, &a.Country, &a.Latitude, &a.Longitude, &a.IsValid, &a.ValidatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("address not found")
		}
		return nil, fmt.Errorf("get address: %w", err)
	}
	return &a, nil
}
