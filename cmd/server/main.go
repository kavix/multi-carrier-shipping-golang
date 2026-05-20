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
	delivery "github.com/kavindus/multi-carrier-shipping-golang/internal/handler/http"
	"github.com/kavindus/multi-carrier-shipping-golang/internal/repository"
	"github.com/kavindus/multi-carrier-shipping-golang/internal/service"
)

func main() {
	// 1. Load Configurations
	cfg := config.Load()

	// 2. Initialize Structured Logger (JSON for production, Text for development)
	var logger *slog.Logger
	if cfg.Env == "production" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	slog.SetDefault(logger)

	logger.Info("Starting multi-carrier shipping microservice", slog.String("env", cfg.Env))

	// 3. Initialize Repositories (In-memory mock)
	repo := repository.NewMemoryShipmentRepository()
	logger.Info("Successfully initialized in-memory database repository")

	// 4. Initialize Services (Dependency Injection)
	svc := service.NewShipmentService(repo)

	// 5. Initialize Delivery Handlers & Router
	hdlr := delivery.NewShipmentHandler(svc)
	router := delivery.NewRouter(hdlr, logger)

	// 6. Configure HTTP Server
	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 7. Setup channels for graceful shutdown listening to OS signals
	serverErrors := make(chan error, 1)

	// Start server in background
	go func() {
		logger.Info("HTTP Server is listening", slog.String("address", serverAddr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	// Listening to OS shutdown signals
	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)

	// 8. Graceful Shutdown Block
	select {
	case err := <-serverErrors:
		logger.Error("Fatal server error during startup", slog.Any("error", err))
		os.Exit(1)

	case sig := <-shutdownSignal:
		logger.Info("Shutdown signal received, starting graceful teardown", slog.String("signal", sig.String()))

		// Create shutdown context with timeout (e.g. 15 seconds) to process remaining requests
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Could not gracefully shut down server", slog.Any("error", err))
			if err := server.Close(); err != nil {
				logger.Error("Forced shutdown closing connection error", slog.Any("error", err))
			}
			os.Exit(1)
		}

		logger.Info("Server exited cleanly. Teardown complete.")
	}
}
