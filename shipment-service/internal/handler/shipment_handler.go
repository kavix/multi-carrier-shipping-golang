package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shipping/shipment-service/internal/service"
)

type ShipmentHandler struct {
	svc *service.ShipmentService
}

func NewShipmentHandler(svc *service.ShipmentService) *ShipmentHandler {
	return &ShipmentHandler{svc: svc}
}

func (h *ShipmentHandler) CreateShipment(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok || userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user_id"})
		return
	}
	var req service.CreateShipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	shipment, err := h.svc.CreateShipment(c.Request.Context(), userID.(string), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, shipment)
}

func (h *ShipmentHandler) GetShipment(c *gin.Context) {
	id := c.Param("id")
	shipment, err := h.svc.GetShipment(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, shipment)
}

func (h *ShipmentHandler) ListShipments(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok || userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user_id"})
		return
	}
	shipments, err := h.svc.ListUserShipments(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, shipments)
}

func (h *ShipmentHandler) UpdateShipment(c *gin.Context) {
	id := c.Param("id")
	var req service.UpdateShipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	shipment, err := h.svc.UpdateShipment(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, shipment)
}

func (h *ShipmentHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.UpdateStatus(c.Request.Context(), id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

func (h *ShipmentHandler) DeleteShipment(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteShipment(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "shipment deleted"})
}

func (h *ShipmentHandler) Routes(r *gin.Engine) {
	r.POST("/shipments", h.CreateShipment)
	r.GET("/shipments", h.ListShipments)
	r.GET("/shipments/:id", h.GetShipment)
	r.PUT("/shipments/:id", h.UpdateShipment)
	r.PATCH("/shipments/:id/status", h.UpdateStatus)
	r.DELETE("/shipments/:id", h.DeleteShipment)
}
