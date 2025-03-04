package ukvat_test

import (
	"os"
	"testing"

	"github.com/creativefabrica/vat"
	"github.com/creativefabrica/vat/ukvat"
	"github.com/stretchr/testify/assert"
)

func TestClient_Validate(t *testing.T) {
	tests := []struct {
		name      string
		creds     ukvat.ClientCredentials
		vatNumber vat.IDNumber
		wantErr   error
	}{
		{
			name:      "Missing credentials",
			creds:     ukvat.ClientCredentials{},
			vatNumber: vat.MustParse("GB123456789"),
			wantErr:   vat.ErrServiceUnavailable,
		},
		{
			name: "Valid VAT number",
			creds: ukvat.ClientCredentials{
				ID:     os.Getenv("UKVAT_API_CLIENT_ID"),
				Secret: os.Getenv("UKVAT_API_CLIENT_SECRET"),
			},
			vatNumber: vat.MustParse("GB146295999727"),
			wantErr:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := ukvat.NewClient(
				tt.creds,
				ukvat.WithBaseURL(ukvat.TestServiceBaseURL),
			)
			err := c.Validate(t.Context(), tt.vatNumber)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
