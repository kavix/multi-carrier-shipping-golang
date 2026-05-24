package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/shipping/api-gateway/internal/handler"
	"github.com/shipping/shared/pkg/logger"
	"github.com/shipping/shared/pkg/middleware"
)

func main() {
	logger.Init()
	log := logger.Get()

	h := handler.NewGatewayHandler()

	r := gin.Default()
	r.Use(middleware.RequestLogger(log))
	r.Use(middleware.AuthMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "api-gateway"})
	})

	h.Routes(r)

	port := getEnv("PORT", "8080")
	log.Info("api-gateway starting", logger.String("port", port))
	if err := r.Run(":" + port); err != nil {
		log.Fatal("server failed", logger.String("err", err.Error()))
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
