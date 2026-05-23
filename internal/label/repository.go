package label

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type PostgresLabelRepository struct {
	db *sql.DB
}

// NewPostgresLabelRepository instantiates a new PostgreSQL database connection for Labels.
func NewPostgresLabelRepository(dsn string) (*PostgresLabelRepository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres db: %w", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS labels (
		id TEXT PRIMARY KEY,
		shipment_id TEXT NOT NULL,
		tracking_number TEXT NOT NULL,
		label_url TEXT NOT NULL,
		status TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL
	);`

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &PostgresLabelRepository{db: db}, nil
}

func (r *PostgresLabelRepository) Create(ctx context.Context, l *Label) error {
	query := `INSERT INTO labels (id, shipment_id, tracking_number, label_url, status, created_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query, l.ID, l.ShipmentID, l.TrackingNumber, l.LabelURL, l.Status, l.CreatedAt)
	if err != nil {
		return ErrLabelAlreadyCancelled // or duplicate error
	}
	return nil
}

func (r *PostgresLabelRepository) GetByID(ctx context.Context, id string) (*Label, error) {
	query := `SELECT id, shipment_id, tracking_number, label_url, status, created_at FROM labels WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	var l Label
	err := row.Scan(&l.ID, &l.ShipmentID, &l.TrackingNumber, &l.LabelURL, &l.Status, &l.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrLabelNotFound
	} else if err != nil {
		return nil, err
	}

	return &l, nil
}

func (r *PostgresLabelRepository) GetByTracking(ctx context.Context, trackingNum string) (*Label, error) {
	query := `SELECT id, shipment_id, tracking_number, label_url, status, created_at FROM labels WHERE tracking_number = $1`
	row := r.db.QueryRowContext(ctx, query, trackingNum)

	var l Label
	err := row.Scan(&l.ID, &l.ShipmentID, &l.TrackingNumber, &l.LabelURL, &l.Status, &l.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrLabelNotFound
	} else if err != nil {
		return nil, err
	}

	return &l, nil
}

func (r *PostgresLabelRepository) Update(ctx context.Context, l *Label) error {
	query := `UPDATE labels SET shipment_id = $1, tracking_number = $2, label_url = $3, status = $4, created_at = $5 WHERE id = $6`
	_, err := r.db.ExecContext(ctx, query, l.ShipmentID, l.TrackingNumber, l.LabelURL, l.Status, l.CreatedAt, l.ID)
	return err
}

func (r *PostgresLabelRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM labels WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *PostgresLabelRepository) Close() error {
	return r.db.Close()
}
