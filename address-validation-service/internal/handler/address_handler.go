package handler

import (
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
	"github.com/shipping/address-validation-service/internal/service"
)

type AddressHandler struct {
	svc *service.AddressService
}

func NewAddressHandler(svc *service.AddressService) *AddressHandler {
	return &AddressHandler{svc: svc}
}

func (h *AddressHandler) ValidateAddress(c *gin.Context) {
	var req struct {
		Address string `json:"address" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	validated, err := h.svc.ValidateAddress(c.Request.Context(), req.Address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, validated)
}

func (h *AddressHandler) GetPickupLocations(c *gin.Context) {
	address := c.Query("address")
	carrier := c.Query("carrier")
	limitStr := c.DefaultQuery("limit", "10")
	if address == "" || carrier == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address and carrier required"})
		return
	}
	limit, _ := strconv.Atoi(limitStr)
	locations, err := h.svc.GetLocations(c.Request.Context(), address, carrier, limit, "pickup")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, locations)
}

func (h *AddressHandler) GetDropLocations(c *gin.Context) {
	address := c.Query("address")
	carrier := c.Query("carrier")
	limitStr := c.DefaultQuery("limit", "10")
	if address == "" || carrier == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address and carrier required"})
		return
	}
	limit, _ := strconv.Atoi(limitStr)
	locations, err := h.svc.GetLocations(c.Request.Context(), address, carrier, limit, "drop")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, locations)
}

func (h *AddressHandler) Routes(r *gin.Engine) {
	r.POST("/addresses/validate", h.ValidateAddress)
	r.GET("/addresses/pickup-locations", h.GetPickupLocations)
	r.GET("/addresses/drop-locations", h.GetDropLocations)
}
