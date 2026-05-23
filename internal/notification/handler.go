package notification

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	service NotificationService
}

func NewNotificationHandler(service NotificationService) *NotificationHandler {
	return &NotificationHandler{
		service: service,
	}
}

type CreateNotificationRequest struct {
	Recipient string `json:"recipient" binding:"required"`
	Method    string `json:"method" binding:"required"` // "EMAIL" or "TELEGRAM"
	Subject   string `json:"subject"`
	Body      string `json:"body" binding:"required"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *NotificationHandler) Create(c *gin.Context) {
	var req CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid request payload: recipient, method, and body are required")
		return
	}

	var logRecord *NotificationLog
	var err error

	switch req.Method {
	case "EMAIL":
		logRecord, err = h.service.SendEmailNotification(c.Request.Context(), req.Recipient, req.Subject, req.Body)
	case "TELEGRAM":
		logRecord, err = h.service.SendTelegramNotification(c.Request.Context(), req.Recipient, req.Body)
	default:
		h.writeError(c, http.StatusBadRequest, "invalid method: must be EMAIL or TELEGRAM")
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  err.Error(),
			"record": logRecord,
		})
		return
	}

	c.JSON(http.StatusCreated, logRecord)
}

func (h *NotificationHandler) List(c *gin.Context) {
	logs, err := h.service.ListLogs(c.Request.Context())
	if err != nil {
		h.writeError(c, http.StatusInternalServerError, "failed to retrieve notification logs")
		return
	}

	c.JSON(http.StatusOK, logs)
}

func (h *NotificationHandler) writeError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, ErrorResponse{Error: message})
}

// CORSMiddleware handles cross-origin resource requests
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RequestLogger middleware specifically for Notification service
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
func ConfigureRouter(handler *NotificationHandler, logger *slog.Logger) http.Handler {
	r := gin.New()

	r.Use(CORSMiddleware())
	r.Use(RequestLogger(logger))
	r.Use(gin.Recovery())

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "customer-notification-service",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	v1 := r.Group("/api/v1")
	{
		v1.POST("/notifications", handler.Create)
		v1.GET("/notifications", handler.List)
	}

	return r
}
