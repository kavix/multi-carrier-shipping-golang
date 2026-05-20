package domain

import "errors"

// Sentinel errors representing domain-specific failure cases.
// These are translated to HTTP status codes in the transport layer.
var (
	ErrShipmentNotFound      = errors.New("shipment not found")
	ErrShipmentAlreadyExists = errors.New("shipment already exists with this tracking number")
	ErrInvalidShipment       = errors.New("invalid shipment data")
	ErrCarrierRequired       = errors.New("carrier is required")
	ErrTrackingNumberRequired = errors.New("tracking number is required")
	ErrInvalidWeight         = errors.New("shipment weight must be greater than zero")
)
