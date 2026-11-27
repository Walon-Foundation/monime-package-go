package monime

import (
	"context"
	"fmt"
	"net/http"
)

// FinancialAccountService provides access to the financial-accounts API.
type FinancialAccountService struct {
	client *Client
}

// FinancialAccount returns the financial-accounts service.
func (c *Client) FinancialAccount() *FinancialAccountService {
	return &FinancialAccountService{client: c}
}

const financialAccountPath = "/financial-accounts"

// Currency is an ISO 4217 currency code supported when creating a financial
// account.
type Currency string

const (
	CurrencyUSD Currency = "USD"
	CurrencySLE Currency = "SLE"
)

// CreateFinancialAccountParams holds the inputs for creating a financial account.
type CreateFinancialAccountParams struct {
	AccountName string   `validate:"required,max=100"`
	Currency    Currency `validate:"required,oneof=USD SLE"`
	// Reference is your own identifier for the account. It must be unique within
	// the Space, so creating a second account with the same reference fails.
	Reference   string `validate:"max=64"`
	Description string `validate:"max=150"`
	Metadata    map[string]any
}

// FinancialAccount is the financial-account resource returned by the API.
type FinancialAccount struct {
	ID          string `json:"id"`
	UVAN        string `json:"uvan"`
	Name        string `json:"name"`
	Currency    string `json:"currency"`
	Reference   string `json:"reference"`
	Description string `json:"description"`
	// Balance is filled only when it was asked for (WithFinancialAccountBalance(true));
	// otherwise the API returns null and it is zero.
	Balance    FinancialAccountBalance `json:"balance"`
	CreateTime string                  `json:"createTime"`
	UpdateTime string                  `json:"updateTime"`
	Metadata   map[string]any          `json:"metadata,omitempty"`
}

// FinancialAccountBalance holds the balance details of a financial account.
type FinancialAccountBalance struct {
	Available FinancialAccountAvailable `json:"available"`
}

// FinancialAccountAvailable holds the available funds, expressed in minor units.
type FinancialAccountAvailable struct {
	Currency string `json:"currency"`
	Value    int64  `json:"value"`
}

// FinancialAccountList is the paginated financial-account list response.
type FinancialAccountList struct {
	Result     []FinancialAccount `json:"result"`
	Pagination Pagination         `json:"pagination"`
}

// Create creates a new financial account.
func (s *FinancialAccountService) Create(ctx context.Context, params CreateFinancialAccountParams) (*FinancialAccount, error) {
	if err := validateStruct(params); err != nil {
		return nil, err
	}

	if err := validateMetadata(params.Metadata, false); err != nil {
		return nil, err
	}

	body := map[string]any{
		"name":     params.AccountName,
		"currency": string(params.Currency),
	}
	putIfSet(body, "reference", params.Reference)
	putIfSet(body, "description", params.Description)
	if params.Metadata != nil {
		body["metadata"] = params.Metadata
	}

	key, err := idempotencyKey(ctx)
	if err != nil {
		return nil, err
	}

	var out FinancialAccount
	if err := s.client.do(ctx, requestOptions{
		method:         http.MethodPost,
		path:           financialAccountPath,
		body:           body,
		idempotencyKey: key,
		out:            &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// Retrieve fetches a single financial account by id. Pass
// WithFinancialAccountBalance(true) to include its balance; the API leaves the
// balance out otherwise. Other list options do not apply here.
func (s *FinancialAccountService) Retrieve(ctx context.Context, financialAccountID string, opts ...ListOption) (*FinancialAccount, error) {
	if financialAccountID == "" {
		return nil, newValidationError("financialAccountID is required")
	}
	query, err := buildListQuery(opts)
	if err != nil {
		return nil, err
	}
	query.Del("limit")
	query.Del("after")

	var out FinancialAccount
	if err := s.client.do(ctx, requestOptions{
		method: http.MethodGet,
		path:   fmt.Sprintf("%s/%s", financialAccountPath, financialAccountID),
		query:  query,
		out:    &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns a page of financial accounts. Use WithLimit and WithAfter to
// page through the collection, and WithFinancialAccountUvan,
// WithFinancialAccountReference or WithFinancialAccountBalance to narrow it.
func (s *FinancialAccountService) List(ctx context.Context, opts ...ListOption) (*FinancialAccountList, error) {
	query, err := buildListQuery(opts)
	if err != nil {
		return nil, err
	}

	var out FinancialAccountList
	if err := s.client.do(ctx, requestOptions{
		method:  http.MethodGet,
		path:    financialAccountPath,
		query:   query,
		out:     &out,
		rawBody: true,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// financialAccountUpdateFields are the fields PATCH /financial-accounts/{id}
// accepts.
var financialAccountUpdateFields = map[string]updateField{
	"name":        {maxLen: 100},
	"reference":   {maxLen: 64},
	"description": {maxLen: 150},
	"metadata":    {},
}

// Update applies a partial update to a financial account. The fields accepted
// are those documented for PATCH /financial-accounts/{id}: name (100
// characters), reference (64), description (150) and metadata, where a nil or
// "" value removes a key.
func (s *FinancialAccountService) Update(ctx context.Context, financialAccountID string, params map[string]any) (*FinancialAccount, error) {
	if financialAccountID == "" {
		return nil, newValidationError("financialAccountID is required")
	}
	if err := validateUpdate(params, financialAccountUpdateFields); err != nil {
		return nil, err
	}

	key, err := idempotencyKey(ctx)
	if err != nil {
		return nil, err
	}

	var out FinancialAccount
	if err := s.client.do(ctx, requestOptions{
		method:         http.MethodPatch,
		path:           fmt.Sprintf("%s/%s", financialAccountPath, financialAccountID),
		body:           params,
		idempotencyKey: key,
		out:            &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}
