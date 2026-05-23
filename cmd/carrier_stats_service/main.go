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

	"github.com/kavindus/multi-carrier-shipping-golang/internal/carrierstats"
	"github.com/kavindus/multi-carrier-shipping-golang/internal/config"
)

func main() {
	cfg := config.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}

	baseURL := os.Getenv("FREIGHTPULSE_BASE_URL")
	if baseURL == "" {
		baseURL = "https://freightpulsehq.com/api/v1"
	}

	apiKey := os.Getenv("FREIGHTPULSE_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("FREIGHTPULSEHQ_API_KEY")
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	mongoDB := os.Getenv("MONGO_DB")
	if mongoDB == "" {
		mongoDB = "carrier_stats_logs"
	}

	var logger *slog.Logger
	if cfg.Env == "production" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	slog.SetDefault(logger)

	logger.Info("Starting Global Carrier Stats Microservice",
		slog.String("env", cfg.Env),
		slog.String("port", port),
		slog.String("base_url", baseURL),
	)

	mongoCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	repo, err := carrierstats.NewMongoCarrierStatsRepository(mongoCtx, mongoURI, mongoDB)
	if err != nil {
		logger.Error("Failed to initialize MongoDB repository", slog.Any("error", err))
		os.Exit(1)
	}
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		_ = repo.Close(shutdownCtx)
	}()

	svc := carrierstats.NewCarrierStatsService(repo, baseURL, apiKey)
	hdlr := carrierstats.NewCarrierStatsHandler(svc)
	router := carrierstats.ConfigureRouter(hdlr, logger)

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
		logger.Info("Carrier stats HTTP server listening", slog.String("address", serverAddr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Error("Fatal carrier stats server error", slog.Any("error", err))
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

		logger.Info("Carrier Stats Service exited cleanly.")
	}
}
