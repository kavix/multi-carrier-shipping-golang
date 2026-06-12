package main

import (
	"context"
	"database/sql"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/shipping/billing-service/internal/config"
	"github.com/shipping/billing-service/internal/consumer"
	"github.com/shipping/billing-service/internal/handler"
	"github.com/shipping/billing-service/internal/repository"
	"github.com/shipping/billing-service/internal/service"
	"github.com/shipping/shared/pkg/kafka"
	"github.com/shipping/shared/pkg/logger"
	"github.com/shipping/shared/pkg/middleware"
	"github.com/shipping/shared/pkg/utils"
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

	paymentProducer := kafka.NewProducer(cfg.KafkaBrokers, kafka.TopicPaymentProcessed)
	defer paymentProducer.Close()

	invoiceProducer := kafka.NewProducer(cfg.KafkaBrokers, kafka.TopicInvoiceGenerated)
	defer invoiceProducer.Close()

	repo := repository.NewBillingRepo(db)
	svc := service.NewBillingService(repo, paymentProducer, invoiceProducer, cfg.StripeSecretKey)
	h := handler.NewBillingHandler(svc)

	// Start Kafka consumer
	shipmentConsumer := consumer.NewShipmentConsumer(cfg.KafkaBrokers, svc)
	go shipmentConsumer.Start(context.Background())

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.DownstreamContextMiddleware())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "billing-service"})
	})
	h.Routes(r)

	log.Info("billing-service starting", logger.String("port", cfg.Port))
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("server", logger.String("err", err.Error()))
	}
}
