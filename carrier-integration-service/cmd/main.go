package main

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/shipping/carrier-integration-service/internal/config"
	"github.com/shipping/carrier-integration-service/internal/handler"
	"github.com/shipping/carrier-integration-service/internal/repository"
	"github.com/shipping/carrier-integration-service/internal/service"
	"github.com/shipping/shared/pkg/logger"
	"github.com/shipping/shared/pkg/middleware"
	"github.com/shipping/shared/pkg/utils"
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

	// Initialize database schema
	if err := utils.InitDB(db, "migrations"); err != nil {
		log.Fatal("db init", logger.String("err", err.Error()))
	}

	repo := repository.NewCarrierRepo(db)
	svc := service.NewCarrierService(repo)
	h := handler.NewCarrierHandler(svc)

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())
	// Extract user_id from headers set by API Gateway
	r.Use(middleware.DownstreamContextMiddleware())

	// Health endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "carrier-integration-service"})
	})
	h.Routes(r)

	log.Info("carrier-integration-service starting", logger.String("port", cfg.Port))
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("server", logger.String("err", err.Error()))
	}
}
