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

	"github.com/kavindus/multi-carrier-shipping-golang/internal/auth"
	"github.com/kavindus/multi-carrier-shipping-golang/internal/config"
)

func main() {
	cfg := config.Load()

	// Auth service uses port 8083 by default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
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
	if dbName == "" { dbName = "auth_db" }

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPass, dbName)

	var logger *slog.Logger
	if cfg.Env == "production" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	slog.SetDefault(logger)

	logger.Info("Starting Auth Microservice", slog.String("env", cfg.Env), slog.String("port", port))

	// 1. Initialize Postgres database
	repo, err := auth.NewPostgresAuthRepository(dsn)
	if err != nil {
		logger.Error("Failed to initialize postgres auth repository", slog.Any("error", err))
		os.Exit(1)
	}
	defer repo.Close()

	// 2. Initialize Service & Handlers
	svc := auth.NewAuthService(repo)
	hdlr := auth.NewAuthHandler(svc)
	router := auth.ConfigureRouter(hdlr, logger)

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
		logger.Info("Auth HTTP Server listening", slog.String("address", serverAddr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Error("Fatal auth server error", slog.Any("error", err))
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

		logger.Info("Auth Service exited cleanly.")
	}
}
