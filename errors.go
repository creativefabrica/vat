package vat

import "errors"

var (
	ErrInvalidFormat          = errors.New("invalid VAT number format")
	ErrNotFound               = errors.New("VAT number not found")
	ErrServiceUnavailable     = errors.New("validation service unavailable")
	ErrInvalidCountryCode     = errors.New("invalid country code")
	ErrUKVATClientNotProvided = errors.New("UK VAT service not provided")
	ErrViesClientNotProvided  = errors.New("VIES service not provided")
)
