package vat

import "context"

//go:generate go tool mockgen -destination=vattest/mock_validation_client.gen.go -package=vattest github.com/pcriv/vat ValidationClient
type ValidationClient interface {
	Validate(ctx context.Context, id IDNumber) error
}
