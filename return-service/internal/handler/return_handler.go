package handler

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/shipping/return-service/internal/service"
)

type ReturnHandler struct {
	svc *service.ReturnService
}

func NewReturnHandler(svc *service.ReturnService) *ReturnHandler {
	return &ReturnHandler{svc: svc}
}

func (h *ReturnHandler) RequestReturn(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var req struct {
		ShipmentID string `json:"shipment_id" binding:"required"`
		Reason     string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ret, err := h.svc.RequestReturn(c.Request.Context(), userID.(string), req.ShipmentID, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, ret)
}

func (h *ReturnHandler) ApproveReturn(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Carrier string `json:"carrier" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ret, err := h.svc.ApproveReturn(c.Request.Context(), id, req.Carrier)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ret)
}

func (h *ReturnHandler) ProcessRefund(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Amount float64 `json:"amount" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.ProcessRefund(c.Request.Context(), id, req.Amount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "refund processed"})
}

func (h *ReturnHandler) GetReturn(c *gin.Context) {
	id := c.Param("id")
	ret, err := h.svc.GetReturn(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ret)
}

func (h *ReturnHandler) ListReturns(c *gin.Context) {
	shipmentID := c.Query("shipment_id")
	if shipmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "shipment_id required"})
		return
	}
	returns, err := h.svc.ListReturns(c.Request.Context(), shipmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, returns)
}

func (h *ReturnHandler) UpdateStatus(c *gin.Context) {
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

func (h *ReturnHandler) Routes(r *gin.Engine) {
	r.POST("/returns", h.RequestReturn)
	r.POST("/returns/:id/approve", h.ApproveReturn)
	r.POST("/returns/:id/refund", h.ProcessRefund)
	r.PUT("/returns/:id", h.UpdateStatus)
	r.GET("/returns/:id", h.GetReturn)
	r.GET("/returns", h.ListReturns)
}
