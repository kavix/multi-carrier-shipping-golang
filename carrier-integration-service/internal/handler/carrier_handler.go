package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shipping/carrier-integration-service/internal/service"
)

type CarrierHandler struct {
	svc *service.CarrierService
}

func NewCarrierHandler(svc *service.CarrierService) *CarrierHandler {
	return &CarrierHandler{svc: svc}
}

func (h *CarrierHandler) RegisterCarrier(c *gin.Context) {
	var req struct {
		Name      string `json:"name" binding:"required"`
		Code      string `json:"code" binding:"required"`
		APIKey    string `json:"api_key" binding:"required"`
		APISecret string `json:"api_secret" binding:"required"`
		BaseURL   string `json:"base_url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	carrier, err := h.svc.RegisterCarrier(c.Request.Context(), req.Name, req.Code, req.APIKey, req.APISecret, req.BaseURL)
	if err != nil {
		// Map DB unique-constraint errors to HTTP 409 Conflict
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			c.JSON(http.StatusConflict, gin.H{"error": "carrier with this code already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, carrier)
}

func (h *CarrierHandler) GetRates(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	weightStr := c.Query("weight")
	if from == "" || to == "" || weightStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from, to, and weight are required"})
		return
	}
	weight, _ := strconv.ParseFloat(weightStr, 64)
	rates, err := h.svc.GetCarrierRates(c.Request.Context(), from, to, weight)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rates)
}

func (h *CarrierHandler) GetTracking(c *gin.Context) {
	carrierCode := c.Query("carrier")
	trackingNumber := c.Query("tracking_number")
	if carrierCode == "" || trackingNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "carrier and tracking_number are required"})
		return
	}
	info, err := h.svc.GetTracking(c.Request.Context(), carrierCode, trackingNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, info)
}

func (h *CarrierHandler) GetPickupLocations(c *gin.Context) {
	carrierCode := c.Query("carrier")
	address := c.Query("address")
	limitStr := c.DefaultQuery("limit", "10")
	if carrierCode == "" || address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "carrier and address are required"})
		return
	}
	limit, _ := strconv.Atoi(limitStr)
	locations, err := h.svc.GetPickupLocations(c.Request.Context(), carrierCode, address, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, locations)
}

func (h *CarrierHandler) GetDropLocations(c *gin.Context) {
	carrierCode := c.Query("carrier")
	address := c.Query("address")
	limitStr := c.DefaultQuery("limit", "10")
	if carrierCode == "" || address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "carrier and address are required"})
		return
	}
	limit, _ := strconv.Atoi(limitStr)
	locations, err := h.svc.GetDropLocations(c.Request.Context(), carrierCode, address, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, locations)
}

func (h *CarrierHandler) Routes(r *gin.Engine) {
	r.POST("/carriers", h.RegisterCarrier)
	r.GET("/carriers/rates", h.GetRates)
	r.GET("/carriers/tracking", h.GetTracking)
	r.GET("/carriers/pickup-locations", h.GetPickupLocations)
	r.GET("/carriers/drop-locations", h.GetDropLocations)
}
