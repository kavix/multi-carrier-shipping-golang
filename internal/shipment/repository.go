package shipment

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

type PostgresShipmentRepository struct {
	db *sql.DB
}

// NewPostgresShipmentRepository instantiates a new PostgreSQL database connection.
func NewPostgresShipmentRepository(dsn string) (*PostgresShipmentRepository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres db: %w", err)
	}

	// Schema now includes username and email columns
	schema := `
	CREATE TABLE IF NOT EXISTS shipments (
		id TEXT PRIMARY KEY,
		carrier TEXT NOT NULL,
		tracking_number TEXT NOT NULL,
		weight REAL NOT NULL,
		origin TEXT NOT NULL,
		destination TEXT NOT NULL,
		status TEXT NOT NULL,
		username TEXT NOT NULL,
		email TEXT NOT NULL DEFAULT '',
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL
	);`

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &PostgresShipmentRepository{db: db}, nil
}

func (r *PostgresShipmentRepository) Create(ctx context.Context, s *Shipment) error {
	query := `INSERT INTO shipments (id, carrier, tracking_number, weight, origin, destination, status, username, email, created_at, updated_at) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := r.db.ExecContext(ctx, query, s.ID, s.Carrier, s.TrackingNumber, s.Weight, s.Origin, s.Destination, s.Status, s.Username, s.Email, s.CreatedAt, s.UpdatedAt)
	if err != nil {
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "unique") || strings.Contains(errStr, "duplicate") {
			return ErrShipmentAlreadyExists
		}
		return fmt.Errorf("failed to create shipment: %w", err)
	}
	return nil
}

func (r *PostgresShipmentRepository) GetByID(ctx context.Context, id string) (*Shipment, error) {
	query := `SELECT id, carrier, tracking_number, weight, origin, destination, status, username, email, created_at, updated_at FROM shipments WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	var s Shipment
	err := row.Scan(&s.ID, &s.Carrier, &s.TrackingNumber, &s.Weight, &s.Origin, &s.Destination, &s.Status, &s.Username, &s.Email, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrShipmentNotFound
	} else if err != nil {
		return nil, err
	}

	return &s, nil
}

func (r *PostgresShipmentRepository) GetByTracking(ctx context.Context, trackingNum string) (*Shipment, error) {
	query := `SELECT id, carrier, tracking_number, weight, origin, destination, status, username, email, created_at, updated_at FROM shipments WHERE tracking_number = $1`
	row := r.db.QueryRowContext(ctx, query, trackingNum)

	var s Shipment
	err := row.Scan(&s.ID, &s.Carrier, &s.TrackingNumber, &s.Weight, &s.Origin, &s.Destination, &s.Status, &s.Username, &s.Email, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrShipmentNotFound
	} else if err != nil {
		return nil, err
	}

	return &s, nil
}

func (r *PostgresShipmentRepository) List(ctx context.Context) ([]*Shipment, error) {
	query := `SELECT id, carrier, tracking_number, weight, origin, destination, status, username, email, created_at, updated_at FROM shipments`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Shipment
	for rows.Next() {
		var s Shipment
		if err := rows.Scan(&s.ID, &s.Carrier, &s.TrackingNumber, &s.Weight, &s.Origin, &s.Destination, &s.Status, &s.Username, &s.Email, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &s)
	}

	return list, nil
}

func (r *PostgresShipmentRepository) ListByUsername(ctx context.Context, username string) ([]*Shipment, error) {
	query := `SELECT id, carrier, tracking_number, weight, origin, destination, status, username, email, created_at, updated_at FROM shipments WHERE username = $1`
	rows, err := r.db.QueryContext(ctx, query, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Shipment
	for rows.Next() {
		var s Shipment
		if err := rows.Scan(&s.ID, &s.Carrier, &s.TrackingNumber, &s.Weight, &s.Origin, &s.Destination, &s.Status, &s.Username, &s.Email, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &s)
	}

	return list, nil
}

func (r *PostgresShipmentRepository) Update(ctx context.Context, s *Shipment) error {
	query := `UPDATE shipments SET carrier = $1, tracking_number = $2, weight = $3, origin = $4, destination = $5, status = $6, username = $7, email = $8, updated_at = $9 WHERE id = $10`
	_, err := r.db.ExecContext(ctx, query, s.Carrier, s.TrackingNumber, s.Weight, s.Origin, s.Destination, s.Status, s.Username, s.Email, s.UpdatedAt, s.ID)
	return err
}

func (r *PostgresShipmentRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM shipments WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *PostgresShipmentRepository) Close() error {
	return r.db.Close()
}
