package carrierstats

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type MongoCarrierStatsRepository struct {
	client     *mongo.Client
	collection *mongo.Collection
}

// NewMongoCarrierStatsRepository creates a MongoDB-backed log repository.
func NewMongoCarrierStatsRepository(ctx context.Context, uri, databaseName string) (*MongoCarrierStatsRepository, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongo: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("failed to ping mongo: %w", err)
	}

	return &MongoCarrierStatsRepository{
		client:     client,
		collection: client.Database(databaseName).Collection("carrier_stats_logs"),
	}, nil
}

func (r *MongoCarrierStatsRepository) Create(ctx context.Context, log *CarrierStatsLog) error {
	if log == nil {
		return fmt.Errorf("log cannot be nil")
	}
	if log.ID == "" {
		log.ID = generateLogID()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}

	_, err := r.collection.InsertOne(ctx, log)
	if err != nil {
		return fmt.Errorf("failed to insert carrier stats log: %w", err)
	}
	return nil
}

func (r *MongoCarrierStatsRepository) List(ctx context.Context, limit int64) ([]*CarrierStatsLog, error) {
	findOpts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	if limit > 0 {
		findOpts.SetLimit(limit)
	}

	cursor, err := r.collection.Find(ctx, bson.D{}, findOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to query carrier stats logs: %w", err)
	}
	defer cursor.Close(ctx)

	var logs []*CarrierStatsLog
	for cursor.Next(ctx) {
		var log CarrierStatsLog
		if err := cursor.Decode(&log); err != nil {
			return nil, fmt.Errorf("failed to decode carrier stats log: %w", err)
		}
		logs = append(logs, &log)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate carrier stats logs: %w", err)
	}

	return logs, nil
}

func (r *MongoCarrierStatsRepository) Close(ctx context.Context) error {
	return r.client.Disconnect(ctx)
}
