package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/creativefabrica/vat"
	"github.com/creativefabrica/vat/ukvat"
	"github.com/creativefabrica/vat/vies"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	vatIN, err := vat.Parse("NL822010690B01")
	if err != nil {
		logger.Error("Invalid VAT number", "error", err)
		os.Exit(1)

		return
	}

	logger.Info("Parsed VAT number", "country_code", vatIN.CountryCode, "number", vatIN.Number)

	vat.MustParse("NL822010690B01")

	httpClient := &http.Client{}
	validator := vat.NewValidator(
		vat.WithViesClient(vies.NewClient(
			vies.WithHTTPClient(httpClient),
		)),
		vat.WithUKVATClient(ukvat.NewClient(
			ukvat.ClientCredentials{
				Secret: os.Getenv("UKVAT_API_CLIENT_SECRET"),
				ID:     os.Getenv("UKVAT_API_CLIENT_ID"),
			},
			ukvat.WithBaseURL(ukvat.TestServiceBaseURL),
			ukvat.WithHTTPClient(httpClient),
		)),
	)

	err = validator.Validate(context.Background(), "GB146295999727")
	if err != nil {
		logger.Error("Invalid VAT number", "error", err)
		os.Exit(1)

		return
	}

	logger.Info("VAT number is valid", "vat_number", "GB146295999727")
}
