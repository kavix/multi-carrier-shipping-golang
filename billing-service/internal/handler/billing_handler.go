package handler

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/shipping/billing-service/internal/service"
)

type BillingHandler struct {
	svc *service.BillingService
}

func NewBillingHandler(svc *service.BillingService) *BillingHandler {
	return &BillingHandler{svc: svc}
}

func (h *BillingHandler) CreateInvoice(c *gin.Context) {
	var req struct {
		ShipmentID  string  `json:"shipment_id" binding:"required"`
		UserID      string  `json:"user_id" binding:"required"`
		Amount      float64 `json:"amount" binding:"required,gt=0"`
		Description string  `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	invoice, err := h.svc.CreateInvoice(c.Request.Context(), req.ShipmentID, req.UserID, req.Amount, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, invoice)
}

func (h *BillingHandler) ProcessPayment(c *gin.Context) {
	var req struct {
		InvoiceID string `json:"invoice_id" binding:"required"`
		Method    string `json:"method" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sessionID, checkoutURL, err := h.svc.ProcessPayment(c.Request.Context(), req.InvoiceID, req.Method)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"session_id":   sessionID,
		"checkout_url": checkoutURL,
	})
}

func (h *BillingHandler) ConfirmPayment(c *gin.Context) {
	var req struct {
		SessionID string `json:"session_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	payment, err := h.svc.ConfirmPayment(c.Request.Context(), req.SessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, payment)
}

func (h *BillingHandler) GetInvoice(c *gin.Context) {
	id := c.Param("id")
	invoice, err := h.svc.GetInvoice(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, invoice)
}

func (h *BillingHandler) GetInvoiceByShipment(c *gin.Context) {
	shipmentID := c.Query("shipment_id")
	if shipmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "shipment_id required"})
		return
	}
	invoice, err := h.svc.GetInvoiceByShipment(c.Request.Context(), shipmentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, invoice)
}

func (h *BillingHandler) Routes(r *gin.Engine) {
	r.POST("/billing/invoices", h.CreateInvoice)
	r.POST("/billing/payments", h.ProcessPayment)
	r.POST("/billing/payments/confirm", h.ConfirmPayment)
	r.GET("/billing/invoices/:id", h.GetInvoice)
	r.GET("/billing/invoices", h.GetInvoiceByShipment)
}
