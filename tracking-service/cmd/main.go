package main

import (
	"context"
	"database/sql"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/shipping/shared/pkg/kafka"
	"github.com/shipping/shared/pkg/logger"
	"github.com/shipping/tracking-service/internal/config"
	"github.com/shipping/tracking-service/internal/consumer"
	"github.com/shipping/tracking-service/internal/handler"
	"github.com/shipping/tracking-service/internal/repository"
	"github.com/shipping/tracking-service/internal/service"
)

func main() {
	logger.Init()
	log := logger.Get()

	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DB)
	if err != nil {
		log.Fatal("db open", logger.String("err", err.Error()))
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatal("db ping", logger.String("err", err.Error()))
	}

	producer := kafka.NewProducer(cfg.KafkaBrokers, kafka.TopicTrackingUpdated)
	defer producer.Close()

	repo := repository.NewTrackingRepo(db)
	svc := service.NewTrackingService(repo, producer)

	// Start HTTP API
	h := handler.NewTrackingHandler(svc)
	r := gin.Default()
	h.Routes(r)

	go func() {
		log.Info("tracking-service API starting", logger.String("port", cfg.Port))
		if err := r.Run(":" + cfg.Port); err != nil {
			log.Fatal("API server", logger.String("err", err.Error()))
		}
	}()

	// Start Kafka consumer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info("shutting down tracking consumer")
		cancel()
	}()

	cons := consumer.NewTrackingConsumer(svc)
	if err := cons.Start(ctx, cfg.KafkaBrokers); err != nil {
		log.Fatal("consumer", logger.String("err", err.Error()))
	}
}
