package vat_test

import (
	"testing"

	"github.com/creativefabrica/vat"
	"github.com/creativefabrica/vat/vattest"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestValidator_Validate(t *testing.T) {
	validationClientMock := vattest.NewMockValidationClient(gomock.NewController(t))
	validator := vat.NewValidator(
		vat.WithUKVATClient(validationClientMock),
		vat.WithViesClient(validationClientMock),
	)

	t.Run("valid VAT number", func(t *testing.T) {
		ctx := t.Context()
		id := vat.MustParse("NL822010690B01")
		validationClientMock.EXPECT().Validate(ctx, id).Return(nil)
		err := validator.Validate(ctx, id.String())
		assert.NoError(t, err)
	})

	t.Run("invalid VAT number", func(t *testing.T) {
		ctx := t.Context()
		id := vat.MustParse("NL822010690B02")
		validationClientMock.EXPECT().Validate(ctx, id).Return(vat.ErrInvalidFormat)
		err := validator.Validate(ctx, id.String())
		assert.ErrorIs(t, err, vat.ErrInvalidFormat)
	})

	t.Run("service unavailable", func(t *testing.T) {
		ctx := t.Context()
		id := vat.MustParse("NL822010690B03")
		validationClientMock.EXPECT().Validate(ctx, id).Return(vat.ErrServiceUnavailable)
		err := validator.Validate(ctx, id.String())
		assert.ErrorIs(t, err, vat.ErrServiceUnavailable)
	})

	t.Run("not found", func(t *testing.T) {
		ctx := t.Context()
		id := vat.MustParse("NL822010690B04")
		validationClientMock.EXPECT().Validate(ctx, id).Return(vat.ErrNotFound)
		err := validator.Validate(ctx, id.String())
		assert.ErrorIs(t, err, vat.ErrNotFound)
	})

	t.Run("invalid country code", func(t *testing.T) {
		ctx := t.Context()
		err := validator.Validate(ctx, "AR822010690B05")
		assert.ErrorIs(t, err, vat.ErrInvalidCountryCode)
	})

	t.Run("invalid VAT number length", func(t *testing.T) {
		ctx := t.Context()
		err := validator.Validate(ctx, "NL")
		assert.ErrorIs(t, err, vat.ErrInvalidFormat)
	})
}
