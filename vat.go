package vat

import (
	"context"
)

type Validator struct {
	viesClient  ValidationClient
	ukVATClient ValidationClient
	abnClient   ValidationClient
}

type ValidatorOption func(*Validator)

func WithViesClient(client ValidationClient) ValidatorOption {
	return func(v *Validator) {
		v.viesClient = client
	}
}

func WithUKVATClient(client ValidationClient) ValidatorOption {
	return func(v *Validator) {
		v.ukVATClient = client
	}
}

func WithANBClient(client ValidationClient) ValidatorOption {
	return func(v *Validator) {
		v.abnClient = client
	}
}

func NewValidator(options ...ValidatorOption) *Validator {
	v := &Validator{}
	for _, option := range options {
		option(v)
	}

	return v
}

// Validate checks the format of a VAT number, and its existence only if the respective client is present.
func (v *Validator) Validate(ctx context.Context, vatNumber string) error {
	id, err := Parse(vatNumber)
	if err != nil {
		return err
	}

	switch id.CountryCode {
	case "AU":
		if v.abnClient == nil {
			return nil
		}

		return v.abnClient.Validate(ctx, id)
	case "GB":
		if v.ukVATClient == nil {
			return nil
		}

		return v.ukVATClient.Validate(ctx, id)
	default:
		if v.viesClient == nil {
			return nil
		}

		return v.viesClient.Validate(ctx, id)
	}
}
