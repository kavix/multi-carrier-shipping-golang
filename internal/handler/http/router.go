package http

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/kavindus/multi-carrier-shipping-golang/internal/handler/middleware"
)

// NewRouter configures the router multiplexer using Go 1.22's enhanced path matching.
func NewRouter(handler *ShipmentHandler, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	// Health Check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "UP"})
	})

	// Shipment Resource Endpoints
	mux.HandleFunc("POST /api/v1/shipments", handler.Create)
	mux.HandleFunc("GET /api/v1/shipments", handler.List)
	mux.HandleFunc("GET /api/v1/shipments/{id}", handler.Get)

	// Apply structured logging middleware
	return middleware.RequestLogger(logger)(mux)
}
