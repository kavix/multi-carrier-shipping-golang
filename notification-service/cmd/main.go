package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/shipping/notification-service/internal/config"
	"github.com/shipping/notification-service/internal/consumer"
	"github.com/shipping/notification-service/internal/service"
	"github.com/shipping/shared/pkg/logger"
)

func main() {
	logger.Init()
	log := logger.Get()

	cfg := config.Load()
	svc := service.NewNotificationService(cfg)
	cons := consumer.NewNotificationConsumer(svc)

	// Simple HTTP health check server
	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok", "service":"notification-service"}`))
		})
		log.Info("notification-service health check starting on :8089")
		if err := http.ListenAndServe(":8089", nil); err != nil {
			log.Error("health check server failed", logger.String("err", err.Error()))
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info("shutting down notification service")
		cancel()
	}()

	log.Info("notification-service starting")
	if err := cons.Start(ctx, cfg.KafkaBrokers); err != nil {
		log.Fatal("consumer", logger.String("err", err.Error()))
	}
}
