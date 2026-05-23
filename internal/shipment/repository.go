package shipment

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type SQLiteShipmentRepository struct {
	db *sql.DB
}

// NewSQLiteShipmentRepository instantiates a new SQLite database connection.
func NewSQLiteShipmentRepository(dbPath string) (*SQLiteShipmentRepository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db: %w", err)
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
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);`

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Dynamic upgrade: Ensure 'username' and 'email' columns exist for backward compatibility with legacy DBs.
	// We ignore the errors if they already exist.
	_, _ = db.Exec("ALTER TABLE shipments ADD COLUMN username TEXT NOT NULL DEFAULT ''")
	_, _ = db.Exec("ALTER TABLE shipments ADD COLUMN email TEXT NOT NULL DEFAULT ''")

	return &SQLiteShipmentRepository{db: db}, nil
}

func (r *SQLiteShipmentRepository) Create(ctx context.Context, s *Shipment) error {
	query := `INSERT INTO shipments (id, carrier, tracking_number, weight, origin, destination, status, username, email, created_at, updated_at) 
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, s.ID, s.Carrier, s.TrackingNumber, s.Weight, s.Origin, s.Destination, s.Status, s.Username, s.Email, s.CreatedAt.Unix(), s.UpdatedAt.Unix())
	if err != nil {
		if strings.Contains(err.Error(), "constraint failed") || strings.Contains(err.Error(), "UNIQUE") {
			return ErrShipmentAlreadyExists
		}
		return fmt.Errorf("failed to create shipment: %w", err)
	}
	return nil
}

func (r *SQLiteShipmentRepository) GetByID(ctx context.Context, id string) (*Shipment, error) {
	query := `SELECT id, carrier, tracking_number, weight, origin, destination, status, username, email, created_at, updated_at FROM shipments WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)

	var s Shipment
	var createdUnix, updatedUnix int64
	err := row.Scan(&s.ID, &s.Carrier, &s.TrackingNumber, &s.Weight, &s.Origin, &s.Destination, &s.Status, &s.Username, &s.Email, &createdUnix, &updatedUnix)
	if err == sql.ErrNoRows {
		return nil, ErrShipmentNotFound
	} else if err != nil {
		return nil, err
	}

	s.CreatedAt = time.Unix(createdUnix, 0)
	s.UpdatedAt = time.Unix(updatedUnix, 0)
	return &s, nil
}

func (r *SQLiteShipmentRepository) GetByTracking(ctx context.Context, trackingNum string) (*Shipment, error) {
	query := `SELECT id, carrier, tracking_number, weight, origin, destination, status, username, email, created_at, updated_at FROM shipments WHERE tracking_number = ?`
	row := r.db.QueryRowContext(ctx, query, trackingNum)

	var s Shipment
	var createdUnix, updatedUnix int64
	err := row.Scan(&s.ID, &s.Carrier, &s.TrackingNumber, &s.Weight, &s.Origin, &s.Destination, &s.Status, &s.Username, &s.Email, &createdUnix, &updatedUnix)
	if err == sql.ErrNoRows {
		return nil, ErrShipmentNotFound
	} else if err != nil {
		return nil, err
	}

	s.CreatedAt = time.Unix(createdUnix, 0)
	s.UpdatedAt = time.Unix(updatedUnix, 0)
	return &s, nil
}

func (r *SQLiteShipmentRepository) List(ctx context.Context) ([]*Shipment, error) {
	query := `SELECT id, carrier, tracking_number, weight, origin, destination, status, username, email, created_at, updated_at FROM shipments`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Shipment
	for rows.Next() {
		var s Shipment
		var createdUnix, updatedUnix int64
		if err := rows.Scan(&s.ID, &s.Carrier, &s.TrackingNumber, &s.Weight, &s.Origin, &s.Destination, &s.Status, &s.Username, &s.Email, &createdUnix, &updatedUnix); err != nil {
			return nil, err
		}
		s.CreatedAt = time.Unix(createdUnix, 0)
		s.UpdatedAt = time.Unix(updatedUnix, 0)
		list = append(list, &s)
	}

	return list, nil
}

func (r *SQLiteShipmentRepository) ListByUsername(ctx context.Context, username string) ([]*Shipment, error) {
	query := `SELECT id, carrier, tracking_number, weight, origin, destination, status, username, email, created_at, updated_at FROM shipments WHERE username = ?`
	rows, err := r.db.QueryContext(ctx, query, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Shipment
	for rows.Next() {
		var s Shipment
		var createdUnix, updatedUnix int64
		if err := rows.Scan(&s.ID, &s.Carrier, &s.TrackingNumber, &s.Weight, &s.Origin, &s.Destination, &s.Status, &s.Username, &s.Email, &createdUnix, &updatedUnix); err != nil {
			return nil, err
		}
		s.CreatedAt = time.Unix(createdUnix, 0)
		s.UpdatedAt = time.Unix(updatedUnix, 0)
		list = append(list, &s)
	}

	return list, nil
}

func (r *SQLiteShipmentRepository) Update(ctx context.Context, s *Shipment) error {
	query := `UPDATE shipments SET carrier = ?, tracking_number = ?, weight = ?, origin = ?, destination = ?, status = ?, username = ?, email = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, s.Carrier, s.TrackingNumber, s.Weight, s.Origin, s.Destination, s.Status, s.Username, s.Email, s.UpdatedAt.Unix(), s.ID)
	return err
}

func (r *SQLiteShipmentRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM shipments WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *SQLiteShipmentRepository) Close() error {
	return r.db.Close()
}
