package main

import (
	"context"
	"database/sql"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/shipping/shared/pkg/kafka"
	"github.com/shipping/shared/pkg/logger"
	"github.com/shipping/shared/pkg/middleware"
	"github.com/shipping/shared/pkg/utils"
	"github.com/shipping/shipment-service/internal/config"
	"github.com/shipping/shipment-service/internal/consumer"
	"github.com/shipping/shipment-service/internal/handler"
	"github.com/shipping/shipment-service/internal/repository"
	"github.com/shipping/shipment-service/internal/service"
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

	// Initialize database schema
	if err := utils.InitDB(db, "migrations"); err != nil {
		log.Fatal("db init", logger.String("err", err.Error()))
	}

	createdProducer := kafka.NewProducer(cfg.KafkaBrokers, kafka.TopicShipmentCreated)
	defer createdProducer.Close()

	updatedProducer := kafka.NewProducer(cfg.KafkaBrokers, kafka.TopicShipmentUpdated)
	defer updatedProducer.Close()

	statusProducer := kafka.NewProducer(cfg.KafkaBrokers, kafka.TopicShipmentStatusChanged)
	defer statusProducer.Close()

	deletedProducer := kafka.NewProducer(cfg.KafkaBrokers, kafka.TopicShipmentDeleted)
	defer deletedProducer.Close()

	repo := repository.NewShipmentRepo(db)
	svc := service.NewShipmentService(repo, createdProducer, updatedProducer, statusProducer, deletedProducer)
	h := handler.NewShipmentHandler(svc)

	// Start label consumer to update shipment with label link
	labelConsumer := consumer.NewLabelConsumer(cfg.KafkaBrokers, repo)
	go labelConsumer.Start(context.Background())

	// Start address consumer to update shipment status after validation
	addressConsumer := consumer.NewAddressConsumer(cfg.KafkaBrokers, repo)
	go addressConsumer.Start(context.Background())

	// Start invoice consumer to update shipment cost
	invoiceConsumer := consumer.NewInvoiceConsumer(cfg.KafkaBrokers, repo)
	go invoiceConsumer.Start(context.Background())

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())
	// Extract user_id from headers set by API Gateway
	r.Use(middleware.DownstreamContextMiddleware())

	// Health endpoint for readiness checks
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "shipment-service"})
	})
	h.Routes(r)

	log.Info("shipment-service starting", logger.String("port", cfg.Port))
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("server", logger.String("err", err.Error()))
	}
}
