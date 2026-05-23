package label

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type LabelHandler struct {
	service LabelService
}

func NewLabelHandler(service LabelService) *LabelHandler {
	return &LabelHandler{
		service: service,
	}
}

type CreateLabelRequest struct {
	ShipmentID  string  `json:"shipment_id"`
	Carrier     string  `json:"carrier"`
	Weight      float64 `json:"weight"`
	Origin      string  `json:"origin"`
	Destination string  `json:"destination"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *LabelHandler) Create(c *gin.Context) {
	var req CreateLabelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid request payload")
		return
	}

	label, err := h.service.CreateLabel(
		c.Request.Context(),
		req.ShipmentID,
		req.Carrier,
		req.Weight,
		req.Origin,
		req.Destination,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, label)
}

func (h *LabelHandler) Get(c *gin.Context) {
	trackingNumber := c.Param("tracking_number")
	if trackingNumber == "" {
		h.writeError(c, http.StatusBadRequest, "missing tracking number")
		return
	}

	label, err := h.service.GetLabelByTracking(c.Request.Context(), trackingNumber)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, label)
}

func (h *LabelHandler) Track(c *gin.Context) {
	trackingNumber := c.Param("tracking_number")
	if trackingNumber == "" {
		h.writeError(c, http.StatusBadRequest, "missing tracking number")
		return
	}

	status, err := h.service.TrackLabel(c.Request.Context(), trackingNumber)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tracking_number": trackingNumber,
		"status":          status,
	})
}

func (h *LabelHandler) Cancel(c *gin.Context) {
	trackingNumber := c.Param("tracking_number")
	if trackingNumber == "" {
		h.writeError(c, http.StatusBadRequest, "missing tracking number")
		return
	}

	err := h.service.CancelLabel(c.Request.Context(), trackingNumber)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "label successfully cancelled",
	})
}

func (h *LabelHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrLabelNotFound):
		h.writeError(c, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrLabelAlreadyCancelled):
		h.writeError(c, http.StatusConflict, err.Error())
	case errors.Is(err, ErrCarrierRequired),
		errors.Is(err, ErrTrackingNumberRequired),
		errors.Is(err, ErrInvalidWeight):
		h.writeError(c, http.StatusBadRequest, err.Error())
	default:
		h.writeError(c, http.StatusInternalServerError, "internal server error")
	}
}

func (h *LabelHandler) writeError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, ErrorResponse{Error: message})
}

// RequestLogger middleware specifically for Label service
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
func ConfigureRouter(handler *LabelHandler, logger *slog.Logger) http.Handler {
	r := gin.New()

	r.Use(RequestLogger(logger))
	r.Use(gin.Recovery())

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "label-service",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	v1 := r.Group("/api/v1")
	{
		v1.POST("/labels", handler.Create)
		v1.GET("/labels/:tracking_number", handler.Get)
		v1.GET("/labels/:tracking_number/track", handler.Track)
		v1.POST("/labels/:tracking_number/cancel", handler.Cancel)
	}

	return r
}
