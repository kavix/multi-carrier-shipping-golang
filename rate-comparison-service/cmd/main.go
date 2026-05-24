package main

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/shipping/shared/pkg/kafka"
	"github.com/shipping/shared/pkg/logger"
	"github.com/shipping/rate-comparison-service/internal/config"
	"github.com/shipping/rate-comparison-service/internal/handler"
	"github.com/shipping/rate-comparison-service/internal/repository"
	"github.com/shipping/rate-comparison-service/internal/service"
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

	producer := kafka.NewProducer(cfg.KafkaBrokers, kafka.TopicRatesCompared)
	defer producer.Close()

	repo := repository.NewRateRepo(db)
	svc := service.NewRateService(repo, producer)
	h := handler.NewRateHandler(svc)

	r := gin.Default()
	h.Routes(r)

	log.Info("rate-comparison-service starting", logger.String("port", cfg.Port))
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("server", logger.String("err", err.Error()))
	}
}
