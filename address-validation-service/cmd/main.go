package main

import (
	"context"
	"database/sql"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/shipping/address-validation-service/internal/config"
	"github.com/shipping/address-validation-service/internal/consumer"
	"github.com/shipping/address-validation-service/internal/handler"
	"github.com/shipping/address-validation-service/internal/repository"
	"github.com/shipping/address-validation-service/internal/service"
	"github.com/shipping/shared/pkg/kafka"
	"github.com/shipping/shared/pkg/logger"
	"github.com/shipping/shared/pkg/middleware"
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

	producer := kafka.NewProducer(cfg.KafkaBrokers, kafka.TopicAddressValidated)
	defer producer.Close()

	shipmentValidatedProducer := kafka.NewProducer(cfg.KafkaBrokers, kafka.TopicShipmentAddressValidated)
	defer shipmentValidatedProducer.Close()

	repo := repository.NewAddressRepo(db)
	svc := service.NewAddressService(repo, producer)
	h := handler.NewAddressHandler(svc)

	// Start shipment consumer for address validation
	shipmentConsumer := consumer.NewShipmentConsumer(cfg.KafkaBrokers, svc, shipmentValidatedProducer)
	go shipmentConsumer.Start(context.Background())

	r := gin.Default()
	r.Use(middleware.DownstreamContextMiddleware())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "address-validation-service"})
	})
	h.Routes(r)

	log.Info("address-validation-service starting", logger.String("port", cfg.Port))
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("server", logger.String("err", err.Error()))
	}
}
