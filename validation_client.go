package vat

import "context"

type ValidationClient interface {
	Validate(ctx context.Context, id IDNumber) error
}
