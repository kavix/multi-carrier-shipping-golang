package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kavindus/multi-carrier-shipping-golang/internal/config"
	"github.com/kavindus/multi-carrier-shipping-golang/internal/label"
)

func main() {
	cfg := config.Load() // Also loads .env into os environment

	// Label service uses port 8082 by default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	shipmentServiceURL := os.Getenv("SHIPMENT_SERVICE_URL")
	if shipmentServiceURL == "" {
		shipmentServiceURL = "http://localhost:8081"
	}

	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		authServiceURL = "http://localhost:8083"
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	if dbHost == "" { dbHost = "localhost" }
	if dbPort == "" { dbPort = "5432" }
	if dbUser == "" { dbUser = "shipping_user" }
	if dbPass == "" { dbPass = "shipping_pass" }
	if dbName == "" { dbName = "label_db" }

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPass, dbName)

	var logger *slog.Logger
	if cfg.Env == "production" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	slog.SetDefault(logger)

	logger.Info("Starting Label Microservice", slog.String("env", cfg.Env), slog.String("port", port))

	// 1. Initialize Postgres Database
	repo, err := label.NewPostgresLabelRepository(dsn)
	if err != nil {
		logger.Error("Failed to initialize postgres repository", slog.Any("error", err))
		os.Exit(1)
	}
	defer repo.Close()

	// 2. Initialize FedEx client (credentials hardcoded for sandbox)
	fedexClient := label.NewFedExClient(
		"https://apis-sandbox.fedex.com",
		"l7c62f6ca219c04c6ba1854d564537a3df",
		"5bed90c7d3e34beb8f731d6e7bc9781d",
	)

	// 3. Initialize DHL client (credentials loaded from .env)
	dhlAPIKey := os.Getenv("DHL_API_KEY")
	if dhlAPIKey == "" {
		logger.Warn("DHL_API_KEY not set; DHL location search will be unavailable")
	}
	dhlClient := label.NewDHLClient("https://api-sandbox.dhl.com", dhlAPIKey)
	logger.Info("DHL Location Finder client initialized", slog.String("base_url", "https://api-sandbox.dhl.com"))

	// 4. Wrap both in a multi-carrier router
	multiCarrier := label.NewMultiCarrierClient(fedexClient, dhlClient)

	// 5. Initialize Service & Handlers
	svc := label.NewLabelService(repo, multiCarrier, shipmentServiceURL, authServiceURL)
	hdlr := label.NewLabelHandler(svc)
	router := label.ConfigureRouter(hdlr, logger)

	// 6. Configure Server
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
		logger.Info("Label HTTP Server listening", slog.String("address", serverAddr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Error("Fatal label server error", slog.Any("error", err))
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

		logger.Info("Label Service exited cleanly.")
	}
}

