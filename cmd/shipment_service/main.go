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
	"github.com/kavindus/multi-carrier-shipping-golang/internal/shipment"
)

func main() {
	cfg := config.Load()

	// Shipment service uses port 8081 by default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	labelServiceURL := os.Getenv("LABEL_SERVICE_URL")
	if labelServiceURL == "" {
		labelServiceURL = "http://localhost:8082"
	}

	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		authServiceURL = "http://localhost:8083"
	}

	notificationServiceURL := os.Getenv("NOTIFICATION_SERVICE_URL")
	if notificationServiceURL == "" {
		notificationServiceURL = "http://localhost:8084"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "shipments.db"
	}

	var logger *slog.Logger
	if cfg.Env == "production" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	slog.SetDefault(logger)

	logger.Info("Starting Shipment Microservice", slog.String("env", cfg.Env), slog.String("port", port))

	// 1. Initialize SQLite Database
	repo, err := shipment.NewSQLiteShipmentRepository(dbPath)
	if err != nil {
		logger.Error("Failed to initialize sqlite repository", slog.Any("error", err))
		os.Exit(1)
	}
	defer repo.Close()

	// 2. Initialize Service & Handlers
	kafkaBrokersStr := os.Getenv("KAFKA_BROKERS")
	var kafkaBrokers []string
	if kafkaBrokersStr != "" {
		kafkaBrokers = strings.Split(kafkaBrokersStr, ",")
	}

	svc := shipment.NewShipmentService(repo, labelServiceURL, authServiceURL, notificationServiceURL, kafkaBrokers)
	hdlr := shipment.NewShipmentHandler(svc)
	router := shipment.ConfigureRouter(hdlr, logger)

	// 3. Configure Server
	serverAddr := fmt.Sprintf(":%s", port)
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	serverErrors := make(chan error, 1)

	go func() {
		logger.Info("Shipment HTTP Server listening", slog.String("address", serverAddr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Error("Fatal shipment server error", slog.Any("error", err))
		os.Exit(1)

	case sig := <-shutdownSignal:
		logger.Info("Shutdown signal received", slog.String("signal", sig.String()))

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Could not gracefully shut down server", slog.Any("error", err))
			_ = server.Close()
			os.Exit(1)
		}

		logger.Info("Shipment Service exited cleanly.")
	}
}
