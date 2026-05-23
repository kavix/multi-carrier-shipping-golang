package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/kavindus/multi-carrier-shipping-golang/internal/config"
	"github.com/kavindus/multi-carrier-shipping-golang/internal/notification"
)

func main() {
	cfg := config.Load()

	// Notification service uses port 8084 by default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "notifications.db"
	}

	var logger *slog.Logger
	if cfg.Env == "production" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	slog.SetDefault(logger)

	logger.Info("Starting Customer Notification Microservice", slog.String("env", cfg.Env), slog.String("port", port))

	// 1. Initialize SQLite Database
	repo, err := notification.NewSQLiteNotificationRepository(dbPath)
	if err != nil {
		logger.Error("Failed to initialize sqlite repository for notifications", slog.Any("error", err))
		os.Exit(1)
	}
	defer repo.Close()

	// 2. Initialize Service & Handlers
	svc := notification.NewNotificationService(repo)
	hdlr := notification.NewNotificationHandler(svc)
	router := notification.ConfigureRouter(hdlr, logger)

	// 2.5 Initialize and Start Kafka Consumer if configured
	kafkaBrokersStr := os.Getenv("KAFKA_BROKERS")
	var consumer *notification.KafkaConsumer
	consumerCtx, consumerCancel := context.WithCancel(context.Background())
	defer consumerCancel()

	if kafkaBrokersStr != "" {
		brokers := strings.Split(kafkaBrokersStr, ",")
		logger.Info("Initializing Kafka Consumer for shipment-notifications", slog.Any("brokers", brokers))
		consumer = notification.NewKafkaConsumer(brokers, "shipment-notifications", "customer-notification-group", svc)
		if consumer != nil {
			go consumer.Start(consumerCtx)
		}
	}

	// 3. Configure Server
	serverAddr := fmt.Sprintf(":%s", port)
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	serverErrors := make(chan error, 1)

	go func() {
		logger.Info("Notification HTTP Server listening", slog.String("address", serverAddr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Error("Fatal notification server error", slog.Any("error", err))
		os.Exit(1)

	case sig := <-shutdownSignal:
		logger.Info("Shutdown signal received", slog.String("signal", sig.String()))

		consumerCancel()
		if consumer != nil {
			logger.Info("Closing Kafka Consumer...")
			if err := consumer.Close(); err != nil {
				logger.Error("Failed to close Kafka Consumer", slog.Any("error", err))
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Could not gracefully shut down notification server", slog.Any("error", err))
			_ = server.Close()
			os.Exit(1)
		}

		logger.Info("Notification Service exited cleanly.")
	}
}
