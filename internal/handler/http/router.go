package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kavindus/multi-carrier-shipping-golang/internal/handler/middleware"
)

// NewRouter configures the Gin router engine and registers routes and middleware.
func NewRouter(handler *ShipmentHandler, logger *slog.Logger) http.Handler {
	// Create a clean Gin instance without default middleware
	r := gin.New()

	// Register custom slog request logger and standard Gin recovery middleware
	r.Use(middleware.RequestLogger(logger))
	r.Use(gin.Recovery())

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// Shipment Resource Endpoints
	v1 := r.Group("/api/v1")
	{
		v1.POST("/shipments", handler.Create)
		v1.GET("/shipments", handler.List)
		v1.GET("/shipments/:id", handler.Get)
	}

	return r
}
