package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kavindus/multi-carrier-shipping-golang/internal/domain"
)

type ShipmentHandler struct {
	service domain.ShipmentService
}

func NewShipmentHandler(service domain.ShipmentService) *ShipmentHandler {
	return &ShipmentHandler{
		service: service,
	}
}

type CreateShipmentRequest struct {
	Carrier        string  `json:"carrier"`
	TrackingNumber string  `json:"tracking_number"`
	Weight         float64 `json:"weight"`
	Origin         string  `json:"origin"`
	Destination    string  `json:"destination"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *ShipmentHandler) Create(c *gin.Context) {
	var req CreateShipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid request payload")
		return
	}

	shipment, err := h.service.CreateShipment(
		c.Request.Context(),
		req.Carrier,
		req.TrackingNumber,
		req.Weight,
		req.Origin,
		req.Destination,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, shipment)
}

func (h *ShipmentHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.writeError(c, http.StatusBadRequest, "missing shipment id")
		return
	}

	shipment, err := h.service.GetShipment(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, shipment)
}

func (h *ShipmentHandler) List(c *gin.Context) {
	shipments, err := h.service.ListShipments(c.Request.Context())
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, shipments)
}

func (h *ShipmentHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrShipmentNotFound):
		h.writeError(c, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrShipmentAlreadyExists):
		h.writeError(c, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrCarrierRequired),
		errors.Is(err, domain.ErrTrackingNumberRequired),
		errors.Is(err, domain.ErrInvalidWeight),
		errors.Is(err, domain.ErrInvalidShipment):
		h.writeError(c, http.StatusBadRequest, err.Error())
	default:
		h.writeError(c, http.StatusInternalServerError, "internal server error")
	}
}

func (h *ShipmentHandler) writeError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, ErrorResponse{Error: message})
}
