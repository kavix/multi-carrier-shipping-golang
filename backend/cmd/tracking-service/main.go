package main

import (
	"net/http"

	"github.com/example/multi-carrier-shipping-golang/backend/internal/config"
	httputil "github.com/example/multi-carrier-shipping-golang/backend/internal/http"
	model "github.com/example/multi-carrier-shipping-golang/backend/internal/model"
	tracking "github.com/example/multi-carrier-shipping-golang/backend/internal/tracking"
)

func main() {
	addr := config.GetEnv("TRACKING_ADDR", ":8083")
	mux := http.NewServeMux()
	mux.HandleFunc("/track", trackingHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		httputil.WriteJSON(w, http.StatusOK, model.ErrorResponse{Error: "tracking-service ok"})
	})
	http.ListenAndServe(addr, mux)
}

func trackingHandler(w http.ResponseWriter, r *http.Request) {
	trackingNumber := r.URL.Query().Get("trackingNumber")
	if trackingNumber == "" {
		httputil.WriteJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "trackingNumber is required"})
		return
	}
	response := tracking.TrackShipment(trackingNumber)
	httputil.WriteJSON(w, http.StatusOK, response)
}
