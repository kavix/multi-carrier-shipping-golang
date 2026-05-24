package handler

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/shipping/tracking-service/internal/service"
)

type TrackingHandler struct {
	svc *service.TrackingService
}

func NewTrackingHandler(svc *service.TrackingService) *TrackingHandler {
	return &TrackingHandler{svc: svc}
}

func (h *TrackingHandler) GetTracking(c *gin.Context) {
	shipmentID := c.Param("shipment_id")
	history, err := h.svc.GetTrackingHistory(c.Request.Context(), shipmentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, history)
}

func (h *TrackingHandler) AddEvent(c *gin.Context) {
	var req struct {
		ShipmentID     string `json:"shipment_id" binding:"required"`
		TrackingNumber string `json:"tracking_number" binding:"required"`
		Carrier        string `json:"carrier" binding:"required"`
		Status         string `json:"status" binding:"required"`
		Location       string `json:"location"`
		Description    string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	event, err := h.svc.AddTrackingEvent(c.Request.Context(), req.ShipmentID, req.TrackingNumber, req.Carrier, req.Status, req.Location, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, event)
}

func (h *TrackingHandler) Routes(r *gin.Engine) {
	r.GET("/tracking/:shipment_id", h.GetTracking)
	r.POST("/tracking/events", h.AddEvent)
}
