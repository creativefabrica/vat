package vies_test

import (
	"testing"

	"github.com/creativefabrica/vat"
	"github.com/creativefabrica/vat/vies"
	"github.com/stretchr/testify/assert"
)

func Test_Client_Validate(t *testing.T) {
	tests := []struct {
		name      string
		vatNumber vat.IDNumber
		wantErr   error
	}{
		{
			name:      "valid VAT number",
			vatNumber: vat.MustParse("NL822010690B01"),
			wantErr:   nil,
		},
		{
			name:      "non existing VAT number",
			vatNumber: vat.MustParse("NL822010690B02"),
			wantErr:   vat.ErrNotFound,
		},
		{
			name:      "invalid format",
			vatNumber: vat.IDNumber{CountryCode: "XX", Number: "822010690B01"},
			wantErr:   vat.ErrInvalidFormat,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := vies.NewClient()
			err := c.Validate(t.Context(), tt.vatNumber)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
