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
	AccountName string   `validate:"required"`
	Currency    Currency `validate:"required,oneof=USD SLE"`
	Description string
	Metadata    map[string]any
}

// FinancialAccount is the financial-account resource returned by the API.
type FinancialAccount struct {
	ID          string                  `json:"id"`
	UVAN        string                  `json:"uvan"`
	Name        string                  `json:"name"`
	Currency    string                  `json:"currency"`
	Reference   string                  `json:"reference"`
	Description string                  `json:"description"`
	Balance     FinancialAccountBalance `json:"balance"`
	CreateTime  string                  `json:"createTime"`
	UpdateTime  string                  `json:"updateTime"`
	Metadata    map[string]any          `json:"metadata,omitempty"`
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

	metadata := params.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	body := map[string]any{
		"name":        params.AccountName,
		"currency":    string(params.Currency),
		"description": params.Description,
		"metadata":    metadata,
	}

	key, err := generateIdempotencyKey()
	if err != nil {
		return nil, &Error{Message: err.Error()}
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

// Retrieve fetches a single financial account by id.
func (s *FinancialAccountService) Retrieve(ctx context.Context, financialAccountID string) (*FinancialAccount, error) {
	if financialAccountID == "" {
		return nil, newValidationError("financialAccountID is required")
	}
	var out FinancialAccount
	if err := s.client.do(ctx, requestOptions{
		method: http.MethodGet,
		path:   fmt.Sprintf("%s/%s", financialAccountPath, financialAccountID),
		out:    &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns all financial accounts.
func (s *FinancialAccountService) List(ctx context.Context) (*FinancialAccountList, error) {
	var out FinancialAccountList
	if err := s.client.do(ctx, requestOptions{
		method:  http.MethodGet,
		path:    financialAccountPath,
		out:     &out,
		rawBody: true,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update applies a partial update to a financial account. The fields accepted
// are those documented for PATCH /financial-accounts/{id} (name, reference,
// description, metadata).
func (s *FinancialAccountService) Update(ctx context.Context, financialAccountID string, params map[string]any) (*FinancialAccount, error) {
	if financialAccountID == "" {
		return nil, newValidationError("financialAccountID is required")
	}
	if len(params) == 0 {
		return nil, newValidationError("at least one field is required to update")
	}

	key, err := generateIdempotencyKey()
	if err != nil {
		return nil, &Error{Message: err.Error()}
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
