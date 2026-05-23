package label

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type SQLiteLabelRepository struct {
	db *sql.DB
}

// NewSQLiteLabelRepository instantiates a new SQLite database connection for Labels.
func NewSQLiteLabelRepository(dbPath string) (*SQLiteLabelRepository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db: %w", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS labels (
		id TEXT PRIMARY KEY,
		shipment_id TEXT NOT NULL,
		tracking_number TEXT NOT NULL,
		label_url TEXT NOT NULL,
		status TEXT NOT NULL,
		created_at INTEGER NOT NULL
	);`

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &SQLiteLabelRepository{db: db}, nil
}

func (r *SQLiteLabelRepository) Create(ctx context.Context, l *Label) error {
	query := `INSERT INTO labels (id, shipment_id, tracking_number, label_url, status, created_at) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, l.ID, l.ShipmentID, l.TrackingNumber, l.LabelURL, l.Status, l.CreatedAt.Unix())
	if err != nil {
		return ErrLabelAlreadyCancelled // or duplicate error
	}
	return nil
}

func (r *SQLiteLabelRepository) GetByID(ctx context.Context, id string) (*Label, error) {
	query := `SELECT id, shipment_id, tracking_number, label_url, status, created_at FROM labels WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)

	var l Label
	var createdUnix int64
	err := row.Scan(&l.ID, &l.ShipmentID, &l.TrackingNumber, &l.LabelURL, &l.Status, &createdUnix)
	if err == sql.ErrNoRows {
		return nil, ErrLabelNotFound
	} else if err != nil {
		return nil, err
	}

	l.CreatedAt = time.Unix(createdUnix, 0)
	return &l, nil
}

func (r *SQLiteLabelRepository) GetByTracking(ctx context.Context, trackingNum string) (*Label, error) {
	query := `SELECT id, shipment_id, tracking_number, label_url, status, created_at FROM labels WHERE tracking_number = ?`
	row := r.db.QueryRowContext(ctx, query, trackingNum)

	var l Label
	var createdUnix int64
	err := row.Scan(&l.ID, &l.ShipmentID, &l.TrackingNumber, &l.LabelURL, &l.Status, &createdUnix)
	if err == sql.ErrNoRows {
		return nil, ErrLabelNotFound
	} else if err != nil {
		return nil, err
	}

	l.CreatedAt = time.Unix(createdUnix, 0)
	return &l, nil
}

func (r *SQLiteLabelRepository) Update(ctx context.Context, l *Label) error {
	query := `UPDATE labels SET shipment_id = ?, tracking_number = ?, label_url = ?, status = ?, created_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, l.ShipmentID, l.TrackingNumber, l.LabelURL, l.Status, l.CreatedAt.Unix(), l.ID)
	return err
}

func (r *SQLiteLabelRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM labels WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *SQLiteLabelRepository) Close() error {
	return r.db.Close()
}
