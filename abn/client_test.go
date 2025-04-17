package abn_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/creativefabrica/vat"
	"github.com/creativefabrica/vat/abn"
)

func TestClient_Validate(t *testing.T) {
	tests := []struct {
		name      string
		guid      string
		vatNumber vat.IDNumber
		wantErr   error
	}{
		{
			name:      "Missing credentials",
			guid:      "",
			vatNumber: vat.MustParse("AU51824753556"),
			wantErr:   vat.ErrServiceUnavailable,
		},
		{
			name:      "Valid VAT number",
			guid:      os.Getenv("ABN_API_AUTH_GUID"),
			vatNumber: vat.MustParse("AU51824753556"),
			wantErr:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := abn.NewClient(tt.guid)
			err := c.Validate(t.Context(), tt.vatNumber)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
