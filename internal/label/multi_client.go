package label

import (
	"context"
	"strings"
)

// MultiCarrierClient implements CarrierService by dispatching to the correct
// carrier-specific client based on the carrier name passed at call time.
type MultiCarrierClient struct {
	fedex *FedExClient
	dhl   *DHLClient
}

// NewMultiCarrierClient constructs a router wrapping both FedEx and DHL clients.
func NewMultiCarrierClient(fedex *FedExClient, dhl *DHLClient) *MultiCarrierClient {
	return &MultiCarrierClient{fedex: fedex, dhl: dhl}
}

// SearchLocations routes the location search to the appropriate carrier backend.
func (m *MultiCarrierClient) SearchLocations(ctx context.Context, carrier, addressStr string) ([]LocationDetail, error) {
	switch strings.ToLower(carrier) {
	case "dhl":
		if m.dhl == nil {
			return nil, ErrUnsupportedCarrier
		}
		return m.dhl.SearchLocations(ctx, carrier, addressStr)
	case "fedex":
		if m.fedex == nil {
			return nil, ErrUnsupportedCarrier
		}
		return m.fedex.SearchLocations(ctx, carrier, addressStr)
	default:
		return nil, ErrUnsupportedCarrier
	}
}

// GenerateLabel routes label generation to the appropriate carrier backend.
func (m *MultiCarrierClient) GenerateLabel(ctx context.Context, shipmentID, carrier string, weight float64, origin, destination string) (*Label, error) {
	switch strings.ToLower(carrier) {
	case "dhl":
		if m.dhl == nil {
			return nil, ErrUnsupportedCarrier
		}
		return m.dhl.GenerateLabel(ctx, shipmentID, carrier, weight, origin, destination)
	case "fedex":
		if m.fedex == nil {
			return nil, ErrUnsupportedCarrier
		}
		return m.fedex.GenerateLabel(ctx, shipmentID, carrier, weight, origin, destination)
	default:
		// For unknown carriers, fall back to the FedEx client as default (existing behaviour)
		if m.fedex != nil {
			return m.fedex.GenerateLabel(ctx, shipmentID, carrier, weight, origin, destination)
		}
		return nil, ErrUnsupportedCarrier
	}
}
