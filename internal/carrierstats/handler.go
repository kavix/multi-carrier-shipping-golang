package carrierstats

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type CarrierStatsHandler struct {
	service CarrierStatsService
}

func NewCarrierStatsHandler(service CarrierStatsService) *CarrierStatsHandler {
	return &CarrierStatsHandler{service: service}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *CarrierStatsHandler) PortCongestion(c *gin.Context) {
	data, err := h.service.GetPortCongestion(c.Request.Context())
	if err != nil {
		h.writeError(c, http.StatusBadGateway, err.Error())
		return
	}
	h.writeRawJSON(c, http.StatusOK, data)
}

func (h *CarrierStatsHandler) FreightRates(c *gin.Context) {
	data, err := h.service.GetFreightRates(c.Request.Context())
	if err != nil {
		h.writeError(c, http.StatusBadGateway, err.Error())
		return
	}
	h.writeRawJSON(c, http.StatusOK, data)
}

func (h *CarrierStatsHandler) FuelPrices(c *gin.Context) {
	data, err := h.service.GetFuelPrices(c.Request.Context())
	if err != nil {
		h.writeError(c, http.StatusBadGateway, err.Error())
		return
	}
	h.writeRawJSON(c, http.StatusOK, data)
}

func (h *CarrierStatsHandler) Disruptions(c *gin.Context) {
	data, err := h.service.GetDisruptions(c.Request.Context())
	if err != nil {
		h.writeError(c, http.StatusBadGateway, err.Error())
		return
	}
	h.writeRawJSON(c, http.StatusOK, data)
}

func (h *CarrierStatsHandler) Carriers(c *gin.Context) {
	data, err := h.service.GetCarriers(c.Request.Context())
	if err != nil {
		h.writeError(c, http.StatusBadGateway, err.Error())
		return
	}
	h.writeRawJSON(c, http.StatusOK, data)
}

func (h *CarrierStatsHandler) Logs(c *gin.Context) {
	limit := int64(50)
	if rawLimit := c.Query("limit"); rawLimit != "" {
		if parsedLimit, err := strconv.ParseInt(rawLimit, 10, 64); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	logs, err := h.service.ListLogs(c.Request.Context(), limit)
	if err != nil {
		h.writeError(c, http.StatusInternalServerError, "failed to retrieve carrier stats logs")
		return
	}

	c.JSON(http.StatusOK, logs)
}

func (h *CarrierStatsHandler) writeRawJSON(c *gin.Context, statusCode int, payload []byte) {
	c.Data(statusCode, "application/json", payload)
}

func (h *CarrierStatsHandler) writeError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, ErrorResponse{Error: message})
}

// CORSMiddleware handles cross-origin resource requests.
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-API-Key")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RequestLogger records basic request metadata.
func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		path := c.Request.URL.Path
		if raw := c.Request.URL.RawQuery; raw != "" {
			path = path + "?" + raw
		}

		c.Next()

		logger.Info("HTTP Request Received",
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", time.Since(startedAt)),
			slog.String("client_ip", c.ClientIP()),
		)
	}
}

// ConfigureRouter configures the Gin router engine and registers routes and middleware.
func ConfigureRouter(handler *CarrierStatsHandler, logger *slog.Logger) http.Handler {
	r := gin.New()

	r.Use(CORSMiddleware())
	r.Use(RequestLogger(logger))
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "global-carrier-stats-service",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	v1 := r.Group("/api/v1/carrier-stats")
	{
		v1.GET("/port-congestion", handler.PortCongestion)
		v1.GET("/freight-rates", handler.FreightRates)
		v1.GET("/fuel-prices", handler.FuelPrices)
		v1.GET("/disruptions", handler.Disruptions)
		v1.GET("/carriers", handler.Carriers)
		v1.GET("/logs", handler.Logs)
	}

	return r
}
