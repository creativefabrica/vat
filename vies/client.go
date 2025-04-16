package vies

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/creativefabrica/vat"
)

const ServiceBaseURL = "https://ec.europa.eu/taxation_customs/vies/rest-api/"

type Client struct {
	httpClient *http.Client
	baseURL    string
	retries    int
}

func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

func WithRetries(retries int) ClientOption {
	return func(c *Client) {
		c.retries = retries
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

// response represents the JSON response structure from the VIES API.
type responseError struct {
	Error string `json:"error"`
}

type response struct {
	ActionSucceed *bool           `json:"actionSucceed"`
	ErrorWrappers []responseError `json:"errorWrappers"`
	CountryCode   string          `json:"countryCode"`
	VATNumber     string          `json:"vatNumber"`
	RequestDate   string          `json:"requestDate"`
	Valid         bool            `json:"valid"`
}

// Validate returns whether the given VAT number is valid or not.
func (c *Client) Validate(ctx context.Context, id vat.IDNumber) error {
	if c.retries == 0 {
		return c.validate(ctx, id)
	}

	var err error
	for i := range c.retries {
		err = c.validate(ctx, id)
		if err == nil {
			return nil
		}

		if errors.Is(err, vat.ErrInvalidFormat) || errors.Is(err, vat.ErrNotFound) {
			break
		}

		time.Sleep(time.Duration(i+1) * time.Second)
	}

	if errors.Is(err, errRateLimitExceeded) {
		err = vat.ErrServiceUnavailable
	}

	return err
}

func (c *Client) validate(ctx context.Context, id vat.IDNumber) error {
	payload := map[string]string{
		"countryCode": id.CountryCode,
		"vatNumber":   id.Number,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return errors.Join(vat.ErrServiceUnavailable, err)
	}

	url := c.baseURL + "/check-vat-number"
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return errors.Join(vat.ErrServiceUnavailable, err)
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Join(vat.ErrServiceUnavailable, err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return errors.Join(vat.ErrServiceUnavailable, err)
	}

	if res.StatusCode == http.StatusBadRequest {
		return vat.ErrInvalidFormat
	}

	if res.StatusCode != http.StatusOK {
		return vat.ErrServiceUnavailable
	}

	var resp response
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return errors.Join(vat.ErrServiceUnavailable, err)
	}

	if resp.ActionSucceed != nil && *resp.ActionSucceed == false {
		if len(resp.ErrorWrappers) == 0 {
			return vat.ErrServiceUnavailable
		}

		errorCode := resp.ErrorWrappers[0].Error
		switch errorCode {
		case "INVALID_INPUT":
			return vat.ErrInvalidFormat
		case "MS_UNAVAILABLE":
			return vat.ErrServiceUnavailable
		case "MS_MAX_CONCURRENT_REQ":
			return errRateLimitExceeded
		default:
			return vat.ErrServiceUnavailable
		}
	}

	if !resp.Valid {
		return vat.ErrNotFound
	}

	return nil
}
