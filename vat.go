package vat

import (
	"context"
)

type Validator struct {
	viesClient  ValidationClient
	ukVATClient ValidationClient
}

type ValidatorOption func(*Validator)

func WithViesClient(viesService ValidationClient) ValidatorOption {
	return func(v *Validator) {
		v.viesClient = viesService
	}
}

func WithUKVATClient(ukVATService ValidationClient) ValidatorOption {
	return func(v *Validator) {
		v.ukVATClient = ukVATService
	}
}

func NewValidator(options ...ValidatorOption) *Validator {
	v := &Validator{}
	for _, option := range options {
		option(v)
	}
	return v
}

// Validate checks a VAT number using the appropriate service based on the country code.
func (v *Validator) Validate(ctx context.Context, vatNumber string) error {
	id, err := Parse(vatNumber)
	if err != nil {
		return err
	}

	switch id.CountryCode {
	case "GB":
		if v.ukVATClient == nil {
			return ErrUKVATClientNotProvided
		}

		return v.ukVATClient.Validate(ctx, id)
	default:
		if v.viesClient == nil {
			return ErrViesClientNotProvided
		}

		return v.viesClient.Validate(ctx, id)
	}
}
