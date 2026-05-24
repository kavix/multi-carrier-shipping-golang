package main

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/shipping/label-generation-service/internal/config"
	"github.com/shipping/label-generation-service/internal/handler"
	"github.com/shipping/label-generation-service/internal/repository"
	"github.com/shipping/label-generation-service/internal/service"
	"github.com/shipping/shared/pkg/kafka"
	"github.com/shipping/shared/pkg/logger"
	"github.com/shipping/shared/pkg/middleware"
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

	producer := kafka.NewProducer(cfg.KafkaBrokers, kafka.TopicLabelGenerated)
	defer producer.Close()

	repo := repository.NewLabelRepo(db)
	svc := service.NewLabelService(repo, producer)
	h := handler.NewLabelHandler(svc)

	r := gin.Default()
	r.Use(middleware.DownstreamContextMiddleware())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "label-generation-service"})
	})
	h.Routes(r)

	log.Info("label-generation-service starting", logger.String("port", cfg.Port))
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("server", logger.String("err", err.Error()))
	}
}
