# VAT

![Build](https://github.com/creativefabrica/vat/actions/workflows/ci.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/creativefabrica/vat)](https://goreportcard.com/report/github.com/creativefabrica/vat)
[![GoDoc](https://godoc.org/github.com/creativefabrica/vat?status.svg)](https://godoc.org/github.com/creativefabrica/vat)
[![MIT licensed](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/creativefabrica/vat/master/LICENSE)

Package for parsing and validating VAT Identification numbers

based on https://github.com/Teamwork/vat with some different design choices

## Installing

```shell
go get https://github.com/creativefabrica/vat
```

```shell
import "github.com/creativefabrica/vat"
```

## Usage

Parsing a VAT Number:

```go
vatIN, err := vat.Parse("NL822010690B01")
if err != nil {
    fmt.Printf("Invalid VAT number: %s\n", err)
    return
}
fmt.Printf("Country Code: %s Number: %s\n", vatIN.CountryCode, vatIN.Number)
```

You can also use the `Must` variant if you want to `panic` on error; this is useful on tests:

```go
vatIN := vat.MustParse("INVALID")
```

For validating that a VAT Number actually exists, two different APIs are used:

* EU VAT numbers are looked up using the [VIES VAT validation API](http://ec.europa.eu/taxation_customs/vies/).
* UK VAT numbers are looked up
using the [UK GOV VAT validation API](https://developer.service.hmrc.gov.uk/api-documentation/docs/api/service/vat-registered-companies-api/2.0)
    * Requires [signing up for the UK API](https://developer.service.hmrc.gov.uk/api-documentation/docs/using-the-hub).

You can pass the clients implemented on the `vies` and `ukvat` packages as functional options to the vat Validator initializer:

```go
validator := vat.NewValidator(
    vat.WithViesClient(vies.NewClient()),
    vat.WithUKVATClient(ukvat.NewClient(
        ukvat.ClientCredentials{
            Secret: os.Getenv("UKVAT_API_CLIENT_SECRET"),
            ID:     os.Getenv("UKVAT_API_CLIENT_ID"),
        },
    )),
)

err := validator.Validate(context.Background(), "GB146295999727")
if err != nil {
    return err
}
```

If you only need EU validation and/or UK validation for some reason, you can skip passing the unneeded clients.<br>
In this case the `Validate` function will only validate format using the `Parse` function.

[Full example](/example/main.go)

### Package usage: vies

```go
httpClient := &http.Client{}
client := vies.NewClient(
    // Use this option to provide a custom http client
    vies.WithHTTPClient(httpClient),
    // Use this option to enable retries in case of rate limiting from the VIES API
    vies.WithRetries(3),
)
```

### Package usage: ukvat

> [!IMPORTANT]
> For validating VAT numbers that begin with **GB** you will need to [sign up](https://developer.service.hmrc.gov.uk/api-documentation/docs/using-the-hub) to gain access to the UK government's VAT API.
> Once you have signed up and acquired a client ID and client secret you can provide them on the intitalizer

```go
httpClient := &http.Client{}
client := ukvat.NewClient(
    ukvat.ClientCredentials{
        Secret: os.Getenv("UKVAT_API_CLIENT_SECRET"),
        ID:     os.Getenv("UKVAT_API_CLIENT_ID"),
    },
    // Use this option to provide a custom http client
    ukvat.WithHTTPClient(httpClient),
)
```

> [!NOTE]
> The `ukvat.Client` struct will cache the auth token needed for the validation requests.
> To avoid getting `403` responses when validating VAT numbers, the client will refresh the token 2 minutes before it expires

If you need to hit the sandbox version of the UK VAT API you can use the following option:

```go
ukvat.WithBaseURL(ukvat.TestServiceBaseURL)
```

### Package usage: abn

> [!IMPORTANT]
> For validating Australian VAT numbers (or ABNs) that begin with **AU** you will need to [register](https://abr.business.gov.au/Tools/WebServicesRegister?AcceptLicenceTerms=Y) for an authentication GUID.

```go
httpClient := &http.Client{}
client := abn.NewClient(
    os.Getenv("ABN_API_AUTH_GUID"),
    // Use this option to provide a custom http client
    abn.WithHTTPClient(httpClient),
)
```

### Package usage: vattest

You can use this package to provide a mock validation client to the vat.Validator.
This is useful in tests:

```go
validationClientMock := vattest.NewMockValidationClient(gomock.NewController(t))
validator := vat.NewValidator(
    vat.WithUKVATClient(validationClientMock),
    vat.WithViesClient(validationClientMock),
)
```