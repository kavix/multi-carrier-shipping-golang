package handler

import (
	"bytes"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type GatewayHandler struct {
	services map[string]string
}

func NewGatewayHandler() *GatewayHandler {
	return &GatewayHandler{
		services: map[string]string{
			"shipment": getEnv("SHIPMENT_SERVICE_URL", "http://shipment-service:8081"),
			"carrier":  getEnv("CARRIER_SERVICE_URL", "http://carrier-integration-service:8082"),
			"rate":     getEnv("RATE_SERVICE_URL", "http://rate-comparison-service:8083"),
			"label":    getEnv("LABEL_SERVICE_URL", "http://label-generation-service:8084"),
			"tracking": getEnv("TRACKING_SERVICE_URL", "http://tracking-service:8085"),
			"address":  getEnv("ADDRESS_SERVICE_URL", "http://address-validation-service:8086"),
			"billing":  getEnv("BILLING_SERVICE_URL", "http://billing-service:8087"),
			"return":   getEnv("RETURN_SERVICE_URL", "http://return-service:8088"),
		},
	}
}

func (h *GatewayHandler) proxy(c *gin.Context, service string) {
	baseURL, ok := h.services[service]
	if !ok {
		c.JSON(http.StatusBadGateway, gin.H{"error": "unknown service"})
		return
	}
	url := baseURL + c.Request.URL.RequestURI()
	body, _ := io.ReadAll(c.Request.Body)
	req, _ := http.NewRequest(c.Request.Method, url, bytes.NewReader(body))

	// Forward all original request headers
	for k, v := range c.Request.Header {
		for _, val := range v {
			req.Header.Add(k, val)
		}
	}

	// Forward user_id from context as custom header for downstream services
	if userID, exists := c.Get("user_id"); exists {
		req.Header.Add("X-User-ID", userID.(string))
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

func (h *GatewayHandler) Routes(r *gin.Engine) {
	r.Any("/shipments", func(c *gin.Context) { h.proxy(c, "shipment") })
	r.Any("/shipments/:id", func(c *gin.Context) { h.proxy(c, "shipment") })
	// Proxy shipment status updates
	r.Any("/shipments/:id/status", func(c *gin.Context) { h.proxy(c, "shipment") })
	r.Any("/carriers", func(c *gin.Context) { h.proxy(c, "carrier") })
	r.Any("/carriers/:id", func(c *gin.Context) { h.proxy(c, "carrier") })
	r.Any("/carriers/:id/rates", func(c *gin.Context) { h.proxy(c, "carrier") })
	r.Any("/carriers/:id/tracking", func(c *gin.Context) { h.proxy(c, "carrier") })
	r.Any("/rates", func(c *gin.Context) { h.proxy(c, "rate") })
	r.Any("/rates/compare", func(c *gin.Context) { h.proxy(c, "rate") })
	r.Any("/labels", func(c *gin.Context) { h.proxy(c, "label") })
	r.Any("/labels/:id", func(c *gin.Context) { h.proxy(c, "label") })
	r.Any("/labels/:id/download", func(c *gin.Context) { h.proxy(c, "label") })
	r.Any("/tracking", func(c *gin.Context) { h.proxy(c, "tracking") })
	r.Any("/tracking/:shipment_id", func(c *gin.Context) { h.proxy(c, "tracking") })
	r.Any("/addresses/validate", func(c *gin.Context) { h.proxy(c, "address") })
	r.Any("/addresses/pickup-locations", func(c *gin.Context) { h.proxy(c, "address") })
	r.Any("/addresses/drop-locations", func(c *gin.Context) { h.proxy(c, "address") })
	r.Any("/billing/invoices", func(c *gin.Context) { h.proxy(c, "billing") })
	r.Any("/billing/invoices/:id", func(c *gin.Context) { h.proxy(c, "billing") })
	r.Any("/billing/payments", func(c *gin.Context) { h.proxy(c, "billing") })
	r.Any("/returns", func(c *gin.Context) { h.proxy(c, "return") })
	r.Any("/returns/:id", func(c *gin.Context) { h.proxy(c, "return") })
	r.Any("/returns/:id/labels", func(c *gin.Context) { h.proxy(c, "return") })
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
