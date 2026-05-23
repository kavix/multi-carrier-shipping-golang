package main

import (
	"net/http"

	carrier "github.com/example/multi-carrier-shipping-golang/backend/internal/carrier"
	"github.com/example/multi-carrier-shipping-golang/backend/internal/config"
	httputil "github.com/example/multi-carrier-shipping-golang/backend/internal/http"
	model "github.com/example/multi-carrier-shipping-golang/backend/internal/model"
)

func main() {
	addr := config.GetEnv("CARRIER_ADDR", ":8082")
	mux := http.NewServeMux()
	mux.HandleFunc("/carriers", carriersHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		httputil.WriteJSON(w, http.StatusOK, model.ErrorResponse{Error: "carrier-service ok"})
	})
	http.ListenAndServe(addr, mux)
}

func carriersHandler(w http.ResponseWriter, r *http.Request) {
	carriers := carrier.ListCarriers()
	httputil.WriteJSON(w, http.StatusOK, carriers)
}
