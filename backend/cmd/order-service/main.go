package main

import (
	"net/http"

	"github.com/example/multi-carrier-shipping-golang/backend/internal/config"
	httputil "github.com/example/multi-carrier-shipping-golang/backend/internal/http"
	model "github.com/example/multi-carrier-shipping-golang/backend/internal/model"
	order "github.com/example/multi-carrier-shipping-golang/backend/internal/order"
)

func main() {
	addr := config.GetEnv("ORDER_ADDR", ":8081")
	mux := http.NewServeMux()
	mux.HandleFunc("/order/quote", quoteHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		httputil.WriteJSON(w, http.StatusOK, model.ErrorResponse{Error: "order-service ok"})
	})
	http.ListenAndServe(addr, mux)
}

func quoteHandler(w http.ResponseWriter, r *http.Request) {
	origin := r.URL.Query().Get("origin")
	destination := r.URL.Query().Get("destination")
	weight := r.URL.Query().Get("weight")

	if origin == "" || destination == "" || weight == "" {
		httputil.WriteJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "origin, destination and weight are required"})
		return
	}

	quote := order.GenerateQuote(origin, destination, weight)
	httputil.WriteJSON(w, http.StatusOK, quote)
}
