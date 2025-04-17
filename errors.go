package vat

import "errors"

var (
	ErrInvalidFormat      = errors.New("invalid vat number format")
	ErrNotFound           = errors.New("vat number not found")
	ErrServiceUnavailable = errors.New("validation service unavailable")
	ErrInvalidCountryCode = errors.New("invalid country code")
)
