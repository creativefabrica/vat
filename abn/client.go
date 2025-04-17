package abn

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/creativefabrica/vat"
)

const ServiceBaseURL = "https://abr.business.gov.au/abrxmlsearch/AbrXmlSearch.asmx/"

type Client struct {
	httpClient *http.Client
	baseURL    string
	guid       string
}

func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

type ClientOption func(*Client)

func NewClient(guid string, options ...ClientOption) *Client {
	c := &Client{
		httpClient: http.DefaultClient,
		baseURL:    ServiceBaseURL,
		guid:       guid,
	}
	for _, option := range options {
		option(c)
	}

	return c
}

func (c *Client) Validate(ctx context.Context, id vat.IDNumber) error {
	v := url.Values{}
	v.Add("searchString", id.Number)
	v.Add("includeHistoricalDetails", "N")
	v.Add("authenticationGuid", c.guid)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.baseURL+"/SearchByABNv202001?"+v.Encode(),
		nil,
	)
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

	var resp apiResponse
	err = xml.Unmarshal(body, &resp)
	if err != nil {
		return errors.Join(vat.ErrServiceUnavailable, err)
	}

	if resp.IsException() {
		return resp.Response.Exception.Error()
	}

	return nil
}

type response struct {
	UsageStatement string     `xml:"usageStatement,omitempty"`
	Exception      *exception `xml:"exception,omitempty"`
}

type apiResponse struct {
	Response *response `xml:"response,omitempty"`
}

// IsException will check if a payload response is an exception
func (r *apiResponse) IsException() bool {
	if r.Response.UsageStatement == "" && r.Response.Exception != nil {
		return true
	}
	return false
}

// exception describes an exception and provides an exception code.
// More information about exceptions and there meaning can be found
// here: https://api.gov.au/service/5b639f0f63f18432cd0e1a66/Exceptions#exception-codes-and-descriptions
type exception struct {
	Description string `xml:"exceptionDescription"`
	Code        string `xml:"exceptionCode"`
}

// Error will return a formatted string with information about an API exception
func (e *exception) Error() error {
	switch e.Description {
	case "Search text is not a valid ABN or ACN":
		return vat.ErrInvalidFormat
	default:
		return fmt.Errorf("%w: %s", vat.ErrServiceUnavailable, e.Description)
	}
}
