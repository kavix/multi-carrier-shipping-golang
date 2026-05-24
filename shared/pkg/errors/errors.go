package errors

import "errors"

var (
	ErrNotFound       = errors.New("resource not found")
	ErrConflict       = errors.New("resource already exists")
	ErrValidation     = errors.New("validation failed")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrCarrierAPI     = errors.New("carrier API error")
	ErrRateNotFound   = errors.New("shipping rate not found")
	ErrAddressInvalid = errors.New("address validation failed")
)
