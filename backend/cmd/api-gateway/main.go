package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/example/multi-carrier-shipping-golang/backend/internal/config"
)

var client = &http.Client{Timeout: 5 * time.Second}

func main() {
	addr := config.GetEnv("GATEWAY_ADDR", ":8080")
	mux := http.NewServeMux()
	mux.HandleFunc("/api/quote", proxyHandler("http://localhost:8081/order/quote"))
	mux.HandleFunc("/api/carriers", proxyHandler("http://localhost:8082/carriers"))
	mux.HandleFunc("/api/track", proxyHandler("http://localhost:8083/track"))
	mux.HandleFunc("/health", healthHandler)

	log.Printf("API gateway listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, `{"status":"ok"}`)
}

func proxyHandler(target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		targetURL, err := url.Parse(target)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		proxyReq, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		proxyReq.Header = r.Header.Clone()
		proxyReq.URL.RawQuery = r.URL.RawQuery

		resp, err := client.Do(proxyReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		copyHeaders(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	}
}

func copyHeaders(dst, src http.Header) {
	for k, values := range src {
		for _, v := range values {
			dst.Add(k, v)
		}
	}
}
