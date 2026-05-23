package tracking

import "github.com/example/multi-carrier-shipping-golang/backend/internal/model"

func TrackShipment(trackingNumber string) model.TrackingResponse {
	if trackingNumber == "" {
		return model.TrackingResponse{
			TrackingNumber: trackingNumber,
			Status:         "unknown",
			LastLocation:   "not available",
			History:        []string{"no tracking identifier provided"},
		}
	}

	return model.TrackingResponse{
		TrackingNumber: trackingNumber,
		Status:         "in transit",
		LastLocation:   "Hub A - Los Angeles, CA",
		History: []string{
			"Label created",
			"Picked up by carrier",
			"Arrived at Hub A",
			"Departed Hub A",
		},
	}
}
