package auth

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service AuthService
}

func NewAuthHandler(service AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

type AuthRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LogActionRequest struct {
	Username string `json:"username" binding:"required"`
	Action   string `json:"action" binding:"required"`
}

type TokenResponse struct {
	Token     string    `json:"token"`
	Username  string    `json:"username"`
	ExpiresAt time.Time `json:"expires_at"`
}

func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return r.URL.Query().Get("token")
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
		return parts[1]
	}
	return ""
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	err := h.service.Register(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "user registered successfully"})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	session, err := h.service.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, TokenResponse{
		Token:     session.Token,
		Username:  session.Username,
		ExpiresAt: session.ExpiresAt,
	})
}

func (h *AuthHandler) Verify(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing token query parameter"})
		return
	}

	username, err := h.service.VerifyToken(c.Request.Context(), token)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"username": username})
}

func (h *AuthHandler) Log(c *gin.Context) {
	var req LogActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	err := h.service.LogAction(c.Request.Context(), req.Username, req.Action)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record log action"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "action recorded successfully"})
}

func (h *AuthHandler) GetLogs(c *gin.Context) {
	token := extractToken(c.Request)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	logs, err := h.service.GetAuditLogs(c.Request.Context(), token)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, logs)
}

func (h *AuthHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrUserAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, ErrInvalidCredentials),
		errors.Is(err, ErrSessionNotFound),
		errors.Is(err, ErrSessionExpired):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
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

// RequestLogger middleware specifically for Auth service
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
func ConfigureRouter(handler *AuthHandler, logger *slog.Logger) http.Handler {
	r := gin.New()

	r.Use(CORSMiddleware())
	r.Use(RequestLogger(logger))
	r.Use(gin.Recovery())

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "auth-service",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	v1 := r.Group("/api/v1/auth")
	{
		v1.POST("/register", handler.Register)
		v1.POST("/login", handler.Login)
		v1.GET("/verify", handler.Verify)
		v1.POST("/logs", handler.Log)
		v1.GET("/logs", handler.GetLogs)
	}

	return r
}
