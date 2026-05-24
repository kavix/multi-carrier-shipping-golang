package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shipping/rate-comparison-service/internal/service"
)

type RateHandler struct {
	svc *service.RateService
}

func NewRateHandler(svc *service.RateService) *RateHandler {
	return &RateHandler{svc: svc}
}

func (h *RateHandler) CompareRates(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var req struct {
		ShipmentID string  `json:"shipment_id" binding:"required"`
		From       string  `json:"from" binding:"required"`
		To         string  `json:"to" binding:"required"`
		Weight     float64 `json:"weight" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	comparison, err := h.svc.CompareRates(c.Request.Context(), userID.(string), req.ShipmentID, req.From, req.To, req.Weight)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, comparison)
}

func (h *RateHandler) GetComparison(c *gin.Context) {
	shipmentID := c.Query("shipment_id")
	if shipmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "shipment_id required"})
		return
	}
	comparison, err := h.svc.GetComparison(c.Request.Context(), shipmentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, comparison)
}

func (h *RateHandler) Routes(r *gin.Engine) {
	r.POST("/rates/compare", h.CompareRates)
	r.GET("/rates/comparison", h.GetComparison)
}
