package vies

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"io"
	"net/http"

	"github.com/creativefabrica/vat"
)

const ServiceBaseURL = "https://ec.europa.eu/taxation_customs/vies/services/checkVatService"

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

type ClientOption func(*Client)

func NewClient(options ...ClientOption) *Client {
	c := &Client{
		httpClient: http.DefaultClient,
		baseURL:    ServiceBaseURL,
	}
	for _, option := range options {
		option(c)
	}

	return c
}

// Validate returns whether the given VAT number is valid or not
func (c *Client) Validate(ctx context.Context, id vat.IDNumber) error {
	envelope, err := c.buildEnvelope(id)
	if err != nil {
		return errors.Join(vat.ErrServiceUnavailable, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ServiceBaseURL, bytes.NewBufferString(envelope))
	if err != nil {
		return errors.Join(vat.ErrServiceUnavailable, err)
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Join(vat.ErrServiceUnavailable, err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	xmlRes, err := io.ReadAll(res.Body)
	if err != nil {
		return errors.Join(vat.ErrServiceUnavailable, err)
	}

	// check if response contains "INVALID_INPUT" string
	if bytes.Contains(xmlRes, []byte("INVALID_INPUT")) {
		return vat.ErrInvalidFormat
	}

	// check if response contains "MS_UNAVAILABLE" string
	if bytes.Contains(xmlRes, []byte("MS_UNAVAILABLE")) {
		return vat.ErrServiceUnavailable
	} else if bytes.Contains(xmlRes, []byte("MS_MAX_CONCURRENT_REQ")) {
		return errors.Join(vat.ErrServiceUnavailable, errors.New("max concurrent requests limit hit"))
	}

	var resEnv struct {
		XMLName xml.Name `xml:"Envelope"`
		Soap    struct {
			XMLName xml.Name `xml:"Body"`
			Soap    struct {
				XMLName xml.Name `xml:"checkVatResponse"`
				Valid   bool     `xml:"valid"`
			}
		}
	}
	err = xml.Unmarshal(xmlRes, &resEnv)
	if err != nil {
		return errors.Join(vat.ErrServiceUnavailable, err) // assume if response data doesn't match the struct, the service is down
	}

	valid := resEnv.Soap.Soap.Valid
	if !valid {
		return vat.ErrNotFound
	}

	return nil
}

type reqEnvelope struct {
	XMLName      xml.Name `xml:"soapenv:Envelope"`
	XMLNSSoapEnv string   `xml:"xmlns:soapenv,attr"`
	Header       struct{} `xml:"soapenv:Header"`
	Body         struct {
		CheckVat struct {
			XMLNS       string `xml:"xmlns,attr"`
			CountryCode string `xml:"countryCode"`
			VATNumber   string `xml:"vatNumber"`
		} `xml:"checkVat"`
	} `xml:"soapenv:Body"`
}

func (c *Client) buildEnvelope(id vat.IDNumber) (string, error) {
	data := reqEnvelope{
		XMLNSSoapEnv: "http://schemas.xmlsoap.org/soap/envelope/",
	}
	data.Body.CheckVat.XMLNS = "urn:ec.europa.eu:taxud:vies:services:checkVat:types"
	data.Body.CheckVat.CountryCode = id.CountryCode
	data.Body.CheckVat.VATNumber = id.Number

	output, err := xml.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(output), nil
}
