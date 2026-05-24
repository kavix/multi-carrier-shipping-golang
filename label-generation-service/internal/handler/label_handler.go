package handler

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/shipping/label-generation-service/internal/service"
)

type LabelHandler struct {
	svc *service.LabelService
}

func NewLabelHandler(svc *service.LabelService) *LabelHandler {
	return &LabelHandler{svc: svc}
}

func (h *LabelHandler) GenerateLabel(c *gin.Context) {
	var req struct {
		ShipmentID string `json:"shipment_id" binding:"required"`
		Carrier    string `json:"carrier" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	label, err := h.svc.GenerateLabel(c.Request.Context(), req.ShipmentID, req.Carrier)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, label)
}

func (h *LabelHandler) GetLabel(c *gin.Context) {
	shipmentID := c.Param("id")
	label, err := h.svc.GetLabel(c.Request.Context(), shipmentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, label)
}

func (h *LabelHandler) DownloadLabel(c *gin.Context) {
	shipmentID := c.Param("id")
	data, err := h.svc.DownloadLabel(c.Request.Context(), shipmentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.Data(http.StatusOK, "application/pdf", data)
}

func (h *LabelHandler) Routes(r *gin.Engine) {
	r.POST("/labels", h.GenerateLabel)
	r.GET("/labels/:id", h.GetLabel)
	r.GET("/labels/:id/download", h.DownloadLabel)
}
