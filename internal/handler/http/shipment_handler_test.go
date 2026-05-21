package http

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kavindus/multi-carrier-shipping-golang/internal/domain"
	"github.com/kavindus/multi-carrier-shipping-golang/internal/repository"
	"github.com/kavindus/multi-carrier-shipping-golang/internal/service"
)

func TestShipmentHandler(t *testing.T) {
	// Set Gin to Test Mode to avoid logging verbosity
	gin.SetMode(gin.TestMode)

	repo := repository.NewMemoryShipmentRepository()
	svc := service.NewShipmentService(repo)
	handler := NewShipmentHandler(svc)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError})) // Silent logger for tests
	router := NewRouter(handler, logger)

	t.Run("Create Shipment - Success", func(t *testing.T) {
		reqBody, _ := json.Marshal(CreateShipmentRequest{
			Carrier:        "DHL",
			TrackingNumber: "TEST-TRK-100",
			Weight:         3.2,
			Origin:         "Chicago",
			Destination:    "Houston",
		})

		req := httptest.NewRequest("POST", "/api/v1/shipments", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Errorf("expected status Created (201), got %d", rr.Code)
		}

		var createdShipment domain.Shipment
		if err := json.NewDecoder(rr.Body).Decode(&createdShipment); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if createdShipment.ID == "" {
			t.Errorf("expected non-empty ID")
		}
		if createdShipment.Carrier != "DHL" {
			t.Errorf("expected carrier 'DHL', got %s", createdShipment.Carrier)
		}
	})

	t.Run("Create Shipment - Missing Carrier (400)", func(t *testing.T) {
		reqBody, _ := json.Marshal(CreateShipmentRequest{
			Carrier:        "",
			TrackingNumber: "TEST-TRK-200",
			Weight:         1.0,
		})

		req := httptest.NewRequest("POST", "/api/v1/shipments", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status BadRequest (400), got %d", rr.Code)
		}

		var errResp ErrorResponse
		_ = json.NewDecoder(rr.Body).Decode(&errResp)
		if errResp.Error != domain.ErrCarrierRequired.Error() {
			t.Errorf("expected error message '%s', got '%s'", domain.ErrCarrierRequired.Error(), errResp.Error)
		}
	})

	t.Run("Get Shipment - Success", func(t *testing.T) {
		// Populate mock data via service first
		created, err := svc.CreateShipment(context.Background(), "UPS", "TEST-TRK-300", 4.0, "Miami", "Denver")
		if err != nil {
			t.Fatalf("failed to setup shipment: %v", err)
		}

		// Make GET request
		req := httptest.NewRequest("GET", "/api/v1/shipments/"+created.ID, nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status OK (200), got %d", rr.Code)
		}

		var retrieved domain.Shipment
		if err := json.NewDecoder(rr.Body).Decode(&retrieved); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if retrieved.ID != created.ID {
			t.Errorf("expected ID '%s', got '%s'", created.ID, retrieved.ID)
		}
	})

	t.Run("Get Shipment - Not Found (404)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/shipments/non-existent-uuid", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("expected status NotFound (404), got %d", rr.Code)
		}
	})
}
