package http

import (
	"encoding/json"
	"errors"
	"net/http"

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

func (h *ShipmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateShipmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	shipment, err := h.service.CreateShipment(
		r.Context(),
		req.Carrier,
		req.TrackingNumber,
		req.Weight,
		req.Origin,
		req.Destination,
	)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(shipment)
}

func (h *ShipmentHandler) Get(w http.ResponseWriter, r *http.Request) {
	// In Go 1.22+, PathValue retrieves values from the matched pattern (e.g. /shipments/{id})
	id := r.PathValue("id")
	if id == "" {
		h.writeError(w, http.StatusBadRequest, "missing shipment id")
		return
	}

	shipment, err := h.service.GetShipment(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(shipment)
}

func (h *ShipmentHandler) List(w http.ResponseWriter, r *http.Request) {
	shipments, err := h.service.ListShipments(r.Context())
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(shipments)
}

func (h *ShipmentHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrShipmentNotFound):
		h.writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrShipmentAlreadyExists):
		h.writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrCarrierRequired),
		errors.Is(err, domain.ErrTrackingNumberRequired),
		errors.Is(err, domain.ErrInvalidWeight),
		errors.Is(err, domain.ErrInvalidShipment):
		h.writeError(w, http.StatusBadRequest, err.Error())
	default:
		h.writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func (h *ShipmentHandler) writeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
