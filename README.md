# VAT

Package for parsing and validating VAT Identification numbers

based on https://github.com/Teamwork/vat with some different design choices

## Installing

```shell
go get https://github.com/pcriv/vat
```

```shell
import "github.com/pcriv/vat"
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

You can also use the `Must` variant if you want to `panic` on error, this is useful on tests:

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
    vat.WithViesService(vies.NewClient()),
    vat.WithUKVATService(ukvat.NewClient(
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

If you only need EU validation or UK validation for some reason, you can skip passing the unneeded client.
If you try to validate a VAT and the respective client is not present either `ErrViesClientNotProvided` or `ErrUKVatClientNotProvided` error will be returned

[Full example](/example/main.go)

### Package usage: vies

```go
httpClient := &http.Client{}
client := vies.NewClient(
    // Use this option to provide a custom http client
    vies.WithHTTPClient(httpClient),
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
> The ukvat.Client structs will cache the auth token needed for the validation requests
> To avoid getting 403s responses when validating VATs the client will refresh the token 1 minute before it expires

If you need to hit the sandbox version of the UK VAT API you can use the following option:

```go
ukvat.WithBaseURL(ukvat.TestServiceBaseURL)
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

## License

Copyright 2025 [Pablo Crivella](https://pcriv.com).
Read [LICENSE](LICENSE) for details.