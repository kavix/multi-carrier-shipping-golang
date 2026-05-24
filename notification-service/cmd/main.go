package main

import (
	"context"
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
	svc := service.NewNotificationService()
	cons := consumer.NewNotificationConsumer(svc)

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
