package service

import (
	"os"
	"testing"
)

func TestGenerateSampleLabel(t *testing.T) {
	details := map[string]interface{}{
		"shipment_id":    "SHIP-12345678-ABCD",
		"carrier":        "FedEx",
		"sender_name":    "Alice Smith",
		"sender":         "123 Sender Way, San Francisco, CA, 94105",
		"receiver_name":  "Bob Jones",
		"receiver":       "456 Recipient Rd, Seattle, WA, 98101",
		"weight":         4.52,
		"service_type":   "Ground",
		"rate_cost":      24.50,
		"rate_currency":  "USD",
		"rate_service":   "FedEx Ground Home Delivery",
	}

	s := &LabelService{}
	pdfBytes, err := s.generateMarotoPDF(details, "123456789012")
	if err != nil {
		t.Fatalf("failed to generate maroto pdf: %v", err)
	}

	outputPath := "/Users/kavindus/.gemini/antigravity-ide/brain/0eb32012-9a67-4c6c-97e6-e84c32c78846/sample_label.pdf"
	err = os.WriteFile(outputPath, pdfBytes, 0644)
	if err != nil {
		t.Fatalf("failed to write output pdf: %v", err)
	}
	t.Logf("sample label pdf written successfully to: %s", outputPath)
}
