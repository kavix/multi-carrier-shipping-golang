package main

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/shipping/return-service/internal/config"
	"github.com/shipping/return-service/internal/handler"
	"github.com/shipping/return-service/internal/repository"
	"github.com/shipping/return-service/internal/service"
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

	producer := kafka.NewProducer(cfg.KafkaBrokers, kafka.TopicReturnCreated)
	defer producer.Close()

	repo := repository.NewReturnRepo(db)
	svc := service.NewReturnService(repo, producer)
	h := handler.NewReturnHandler(svc)

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.DownstreamContextMiddleware())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "return-service"})
	})
	h.Routes(r)

	log.Info("return-service starting", logger.String("port", cfg.Port))
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("server", logger.String("err", err.Error()))
	}
}
