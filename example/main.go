package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/pcriv/vat"
	"github.com/pcriv/vat/ukvat"
	"github.com/pcriv/vat/vies"
)

func main() {
	vatIN, err := vat.Parse("NL822010690B01")
	if err != nil {
		fmt.Printf("Invalid VAT number: %s\n", err)
		return
	}
	fmt.Printf("Country Code: %s Number: %s\n", vatIN.CountryCode, vatIN.Number)

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
		fmt.Printf("Invalid VAT number: %s\n", err)
		return
	}
	fmt.Println("Valid VAT number")
}
