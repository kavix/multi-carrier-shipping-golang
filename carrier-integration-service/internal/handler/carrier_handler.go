package handler

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shipping/carrier-integration-service/internal/client"
	"github.com/shipping/carrier-integration-service/internal/service"
)

type CarrierHandler struct {
	svc *service.CarrierService
}

func NewCarrierHandler(svc *service.CarrierService) *CarrierHandler {
	return &CarrierHandler{svc: svc}
}

func (h *CarrierHandler) RegisterCarrier(c *gin.Context) {
	var req struct {
		Name      string `json:"name" binding:"required"`
		Code      string `json:"code" binding:"required"`
		APIKey    string `json:"api_key" binding:"required"`
		APISecret string `json:"api_secret" binding:"required"`
		BaseURL   string `json:"base_url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	carrier, err := h.svc.RegisterCarrier(c.Request.Context(), req.Name, req.Code, req.APIKey, req.APISecret, req.BaseURL)
	if err != nil {
		// Map DB unique-constraint errors to HTTP 409 Conflict
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			c.JSON(http.StatusConflict, gin.H{"error": "carrier with this code already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, carrier)
}

func (h *CarrierHandler) GetRates(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	weightStr := c.Query("weight")
	if from == "" || to == "" || weightStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from, to, and weight are required"})
		return
	}
	weight, _ := strconv.ParseFloat(weightStr, 64)
	rates, err := h.svc.GetCarrierRates(c.Request.Context(), from, to, weight)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rates)
}

func (h *CarrierHandler) GetTracking(c *gin.Context) {
	carrierCode := c.Query("carrier")
	trackingNumber := c.Query("tracking_number")
	if carrierCode == "" || trackingNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "carrier and tracking_number are required"})
		return
	}
	info, err := h.svc.GetTracking(c.Request.Context(), carrierCode, trackingNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, info)
}

func (h *CarrierHandler) GetPickupLocations(c *gin.Context) {
	carrierCode := c.Query("carrier")
	address := c.Query("address")
	limitStr := c.DefaultQuery("limit", "10")
	if carrierCode == "" || address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "carrier and address are required"})
		return
	}
	limit, _ := strconv.Atoi(limitStr)
	locations, err := h.svc.GetPickupLocations(c.Request.Context(), carrierCode, address, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, locations)
}

func (h *CarrierHandler) GetDropLocations(c *gin.Context) {
	carrierCode := c.Query("carrier")
	address := c.Query("address")
	limitStr := c.DefaultQuery("limit", "10")
	if carrierCode == "" || address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "carrier and address are required"})
		return
	}
	limit, _ := strconv.Atoi(limitStr)
	locations, err := h.svc.GetDropLocations(c.Request.Context(), carrierCode, address, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, locations)
}

func (h *CarrierHandler) ValidatePostalCode(c *gin.Context) {
	carrierCode := c.Query("carrier")
	countryCode := c.Query("country")
	postalCode := c.Query("postal_code")
	if carrierCode == "" || countryCode == "" || postalCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "carrier, country, and postal_code are required"})
		return
	}
	valid, err := h.svc.ValidatePostalCode(c.Request.Context(), carrierCode, countryCode, postalCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"valid": valid})
}

// fedexAddressDTO is the wire shape of a FedEx address used by the
// create-shipment endpoint. We translate it to the client's FedExAddress
// before calling the service.
type fedexAddressDTO struct {
	StreetLines     []string `json:"street_lines"`
	City            string   `json:"city"`
	StateOrProvince string   `json:"state_or_province_code"`
	PostalCode      string   `json:"postal_code"`
	CountryCode     string   `json:"country_code"`
}

type fedexContactDTO struct {
	PersonName  string `json:"person_name"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
	CompanyName string `json:"company_name"`
}

// fedexCreateShipmentRequest is the wire DTO accepted by POST
// /carriers/fedex/create-shipment. It is converted into a
// client.CreateShipmentInput before being passed to the service layer.
type fedexCreateShipmentRequest struct {
	AccountNumber                  string            `json:"account_number"`
	ServiceType                    string            `json:"service_type"`
	PackagingType                  string            `json:"packaging_type"`
	Weight                         float64           `json:"weight"`
	WeightUnits                    string            `json:"weight_units"`
	Sender                         fedexAddressDTO   `json:"sender"`
	SenderContact                  fedexContactDTO   `json:"sender_contact"`
	Recipient                      fedexAddressDTO   `json:"recipient"`
	RecipientContact               fedexContactDTO   `json:"recipient_contact"`
	IsInternational                bool    `json:"is_international"`
	TotalCustomsValue              float64 `json:"total_customs_value"`
	TotalCustomsCurrency           string  `json:"total_customs_currency"`
	CommodityDescription           string  `json:"commodity_description"`
	CommodityCountryOfManufacture  string  `json:"commodity_country_of_manufacture"`
	CommodityQuantity              int     `json:"commodity_quantity"`
	CommodityUnitPrice             float64 `json:"commodity_unit_price"`
}

func (d fedexAddressDTO) toClient() client.FedExAddress {
	return client.FedExAddress{
		StreetLines:     d.StreetLines,
		City:            d.City,
		StateOrProvince: d.StateOrProvince,
		PostalCode:      d.PostalCode,
		CountryCode:     d.CountryCode,
	}
}

func (d fedexContactDTO) toClient() client.FedExContact {
	return client.FedExContact{
		PersonName:  d.PersonName,
		PhoneNumber: d.PhoneNumber,
		Email:       d.Email,
		CompanyName: d.CompanyName,
	}
}

func (r fedexCreateShipmentRequest) toClient() client.CreateShipmentInput {
	return client.CreateShipmentInput{
		AccountNumber:                 r.AccountNumber,
		ServiceType:                   r.ServiceType,
		PackagingType:                 r.PackagingType,
		Weight:                        r.Weight,
		WeightUnits:                   r.WeightUnits,
		Sender:                        r.Sender.toClient(),
		SenderContact:                 r.SenderContact.toClient(),
		Recipient:                     r.Recipient.toClient(),
		RecipientContact:              r.RecipientContact.toClient(),
		IsInternational:               r.IsInternational,
		TotalCustomsValue:             r.TotalCustomsValue,
		TotalCustomsCurrency:          r.TotalCustomsCurrency,
		CommodityDescription:          r.CommodityDescription,
		CommodityCountryOfManufacture: r.CommodityCountryOfManufacture,
		CommodityQuantity:             r.CommodityQuantity,
		CommodityUnitPrice:            r.CommodityUnitPrice,
	}
}

// CreateFedExShipment accepts a JSON payload, hands it to the service layer
// to call FedEx's /ship/v1/shipments endpoint, and returns the resulting
// tracking number plus the base64-encoded label PDF.
func (h *CarrierHandler) CreateFedExShipment(c *gin.Context) {
	var input fedexCreateShipmentRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fallback to default sandbox account if none is provided in the incoming request payload
	if input.AccountNumber == "" {
		// Match the Python script account number
		input.AccountNumber = "740561073"
	}

	result, err := h.svc.CreateFedExShipment(c.Request.Context(), input.toClient())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the tracking number and label encoded back to base64 for safe JSON transit
	c.JSON(http.StatusOK, gin.H{
		"tracking_number": result.TrackingNumber,
		"label_pdf_b64":   base64.StdEncoding.EncodeToString(result.LabelPDF),
		"label_format":    result.LabelFormat,
		"service_type":    result.ServiceType,
	})
}

func (h *CarrierHandler) Routes(r *gin.Engine) {
	r.POST("/carriers", h.RegisterCarrier)
	r.GET("/carriers/rates", h.GetRates)
	r.GET("/carriers/tracking", h.GetTracking)
	r.GET("/carriers/pickup-locations", h.GetPickupLocations)
	r.GET("/carriers/drop-locations", h.GetDropLocations)
	r.GET("/carriers/validate-postal-code", h.ValidatePostalCode)
	r.POST("/carriers/fedex/create-shipment", h.CreateFedExShipment)
}
