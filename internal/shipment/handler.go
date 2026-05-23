package shipment

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ShipmentHandler struct {
	service ShipmentService
}

func NewShipmentHandler(service ShipmentService) *ShipmentHandler {
	return &ShipmentHandler{
		service: service,
	}
}

type CreateShipmentRequest struct {
	Carrier     string  `json:"carrier"`
	Weight      float64 `json:"weight"`
	Origin      string  `json:"origin"`
	Destination string  `json:"destination"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *ShipmentHandler) Create(c *gin.Context) {
	var req CreateShipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid request payload")
		return
	}

	shipment, label, err := h.service.CreateShipment(
		c.Request.Context(),
		req.Carrier,
		req.Weight,
		req.Origin,
		req.Destination,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"shipment": shipment,
		"label":    label,
	})
}

func (h *ShipmentHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.writeError(c, http.StatusBadRequest, "missing shipment id")
		return
	}

	shipment, err := h.service.GetShipment(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, shipment)
}

func (h *ShipmentHandler) List(c *gin.Context) {
	shipments, err := h.service.ListShipments(c.Request.Context())
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, shipments)
}

func (h *ShipmentHandler) CancelInternal(c *gin.Context) {
	trackingNumber := c.Param("tracking_number")
	if trackingNumber == "" {
		h.writeError(c, http.StatusBadRequest, "missing tracking number")
		return
	}

	err := h.service.CancelShipmentByTracking(c.Request.Context(), trackingNumber)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "shipment successfully cancelled",
	})
}

func (h *ShipmentHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrShipmentNotFound):
		h.writeError(c, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrShipmentAlreadyExists):
		h.writeError(c, http.StatusConflict, err.Error())
	case errors.Is(err, ErrCarrierRequired),
		errors.Is(err, ErrInvalidWeight):
		h.writeError(c, http.StatusBadRequest, err.Error())
	default:
		h.writeError(c, http.StatusInternalServerError, err.Error())
	}
}

func (h *ShipmentHandler) writeError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, ErrorResponse{Error: message})
}

// RequestLogger middleware specifically for Shipment service
func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		param := gin.LogFormatterParams{
			Request: c.Request,
			Keys:    c.Keys,
		}

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Info("HTTP Request Received",
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", time.Since(start)),
			slog.String("client_ip", c.ClientIP()),
			slog.String("error_message", param.ErrorMessage),
		)
	}
}

// ConfigureRouter configures the Gin router engine and registers routes and middleware.
func ConfigureRouter(handler *ShipmentHandler, logger *slog.Logger) http.Handler {
	r := gin.New()

	r.Use(RequestLogger(logger))
	r.Use(gin.Recovery())

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "shipment-service",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	v1 := r.Group("/api/v1")
	{
		v1.POST("/shipments", handler.Create)
		v1.GET("/shipments", handler.List)
		v1.GET("/shipments/:id", handler.Get)

		// Internal webhook-like endpoints called by other microservices
		v1.PUT("/shipments/tracking/:tracking_number/cancel", handler.CancelInternal)
	}

	return r
}
