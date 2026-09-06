package monime

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// CountryService provides access to the countries API, the read-only directory
// of countries Monime operates in and the currencies they support.
type CountryService struct {
	client *Client
}

// Country returns the countries service.
func (c *Client) Country() *CountryService {
	return &CountryService{client: c}
}

const countryPath = "/countries"

// CountryCurrency describes a currency used in a country. UnitLength is the
// number of minor units per major unit, e.g. 2 for a currency with cents.
type CountryCurrency struct {
	Code       string `json:"code"`
	Unit       string `json:"unit"`
	UnitLength int    `json:"unitLength"`
}

// Country is the country resource returned by the API. Code is the ISO 3166-1
// alpha-2 code and SupportedCurrencies includes Currency's own code.
type Country struct {
	Code                string          `json:"code"`
	Name                string          `json:"name"`
	Currency            CountryCurrency `json:"currency"`
	SupportedCurrencies []string        `json:"supportedCurrencies,omitempty"`
}

// CountryList is the paginated country list response.
type CountryList struct {
	Result     []Country  `json:"result"`
	Pagination Pagination `json:"pagination"`
}

// Retrieve fetches a single country by its ISO 3166-1 alpha-2 code (e.g. "SL").
func (s *CountryService) Retrieve(ctx context.Context, countryCode string) (*Country, error) {
	if len(countryCode) != 2 {
		return nil, newValidationError("countryCode is required and must be an ISO 3166-1 alpha-2 code")
	}

	var out Country
	if err := s.client.do(ctx, requestOptions{
		method: http.MethodGet,
		path:   fmt.Sprintf("%s/%s", countryPath, strings.ToUpper(countryCode)),
		out:    &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns a page of supported countries. Use WithLimit and WithAfter to
// page through the collection.
func (s *CountryService) List(ctx context.Context, opts ...ListOption) (*CountryList, error) {
	query, err := buildListQuery(opts)
	if err != nil {
		return nil, err
	}

	var out CountryList
	if err := s.client.do(ctx, requestOptions{
		method:  http.MethodGet,
		path:    countryPath,
		query:   query,
		out:     &out,
		rawBody: true,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}
