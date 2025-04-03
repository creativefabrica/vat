package ukvat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/creativefabrica/vat"
)

// API Documentation:
// https://developer.service.hmrc.gov.uk/api-documentation/docs/api/service/vat-registered-companies-api/2.0/oas/page
const (
	ServiceBaseURL     = "https://api.service.hmrc.gov.uk"
	TestServiceBaseURL = "https://test-api.service.hmrc.gov.uk"
)

type authToken struct {
	Value     string `json:"access_token"`
	ExpiresIn int64  `json:"expires_in"`
}

type ClientCredentials struct {
	Secret string
	ID     string
}

type Client struct {
	httpClient  *http.Client
	baseURL     string
	credentials ClientCredentials
	token       string
	expiry      time.Time
	mutex       sync.Mutex
}

func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.baseURL = url
	}
}

type ClientOption func(*Client)

func NewClient(creds ClientCredentials, options ...ClientOption) *Client {
	c := &Client{
		httpClient:  http.DefaultClient,
		baseURL:     ServiceBaseURL,
		credentials: creds,
	}
	for _, option := range options {
		option(c)
	}

	return c
}

func (c *Client) Authenticate(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	data := url.Values{}
	data.Set("client_secret", c.credentials.Secret)
	data.Set("client_id", c.credentials.ID)
	data.Set("grant_type", "client_credentials")
	data.Set("scope", "read:vat")

	url := c.baseURL + "/oauth/token"

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewBufferString(data.Encode()),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		type errorRes struct {
			Code        string `json:"code"`
			Description string `json:"error_description"`
		}

		var errRes errorRes

		err = json.NewDecoder(res.Body).Decode(&errRes)
		if err != nil {
			errRes.Description = "Failed to decode error response"
		}

		return errors.Join(
			vat.ErrServiceUnavailable,
			fmt.Errorf("failed to authenticate with UK VAT API: %s", errRes.Description),
		)
	}

	var token authToken

	err = json.NewDecoder(res.Body).Decode(&token)
	if err != nil {
		return errors.Join(
			vat.ErrServiceUnavailable,
			fmt.Errorf("failed to decode UK VAT API response: %w", err),
		)
	}

	c.token = token.Value
	c.expiry = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)

	return nil
}

func (c *Client) Validate(ctx context.Context, id vat.IDNumber) error {
	// Check if token needs to be refreshed
	c.mutex.Lock()
	needsAuth := time.Now().After(c.expiry.Add(-2 * time.Minute))
	c.mutex.Unlock()

	if needsAuth {
		err := c.Authenticate(ctx)
		if err != nil {
			return err
		}
	}

	// Use the token to make the request
	c.mutex.Lock()
	token := c.token
	c.mutex.Unlock()

	url := fmt.Sprintf("%s/organisations/vat/check-vat-number/lookup/%s", c.baseURL, id.Number)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return errors.Join(vat.ErrServiceUnavailable, err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.hmrc.2.0+json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Join(vat.ErrServiceUnavailable, err)
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusUnauthorized:
		return errors.Join(
			vat.ErrServiceUnavailable,
			errors.New("unauthorized request to UK VAT API"),
		)
	case http.StatusBadRequest:
		return vat.ErrInvalidFormat
	case http.StatusNotFound:
		return vat.ErrNotFound
	}

	if res.StatusCode != http.StatusOK {
		return errors.Join(
			vat.ErrServiceUnavailable,
			fmt.Errorf("unexpected status code from UK VAT API: %d", res.StatusCode),
		)
	}

	// If we receive a valid 200 response from this API, it means the VAT number exists and is valid
	return nil
}
