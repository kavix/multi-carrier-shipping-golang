package carrierstats

import (
	"context"
	"time"
)

// CarrierStatsLog captures each upstream FreightPulse request and its outcome.
type CarrierStatsLog struct {
	ID              string    `bson:"id" json:"id"`
	Endpoint        string    `bson:"endpoint" json:"endpoint"`
	URL             string    `bson:"url" json:"url"`
	StatusCode      int       `bson:"status_code" json:"status_code"`
	DurationMS      int64     `bson:"duration_ms" json:"duration_ms"`
	ResponseSize    int       `bson:"response_size" json:"response_size"`
	Success         bool      `bson:"success" json:"success"`
	Error           string    `bson:"error,omitempty" json:"error,omitempty"`
	ResponsePreview string    `bson:"response_preview,omitempty" json:"response_preview,omitempty"`
	CreatedAt       time.Time `bson:"created_at" json:"created_at"`
}

// CarrierStatsRepository defines the MongoDB persistence contract for logs.
type CarrierStatsRepository interface {
	Create(ctx context.Context, log *CarrierStatsLog) error
	List(ctx context.Context, limit int64) ([]*CarrierStatsLog, error)
	Close(ctx context.Context) error
}

// CarrierStatsService defines the application operations for FreightPulse data.
type CarrierStatsService interface {
	GetPortCongestion(ctx context.Context) ([]byte, error)
	GetFreightRates(ctx context.Context) ([]byte, error)
	GetFuelPrices(ctx context.Context) ([]byte, error)
	GetDisruptions(ctx context.Context) ([]byte, error)
	GetCarriers(ctx context.Context) ([]byte, error)
	ListLogs(ctx context.Context, limit int64) ([]*CarrierStatsLog, error)
}
