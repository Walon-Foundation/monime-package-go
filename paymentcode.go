package monime

import (
	"context"
	"fmt"
	"net/http"
)

// PaymentCodeService provides access to the payment-codes API.
type PaymentCodeService struct {
	client *Client
}

// PaymentCode returns the payment-codes service.
func (c *Client) PaymentCode() *PaymentCodeService {
	return &PaymentCodeService{client: c}
}

const paymentCodePath = "/payment-codes"

// CreatePaymentCodeParams holds the inputs for creating a USSD payment code.
type CreatePaymentCodeParams struct {
	PaymentName string `validate:"required"`
	Amount      int64  `validate:"gt=0"`
	Name        string `validate:"required"`
	PhoneNumber string `validate:"required"`
	// FinancialAccountID is optional; when empty the space's main account is used.
	FinancialAccountID string
}

// PaymentCode is the payment-code resource returned by the API.
type PaymentCode struct {
	ID                     string                     `json:"id"`
	Mode                   string                     `json:"mode"`
	Status                 string                     `json:"status"`
	Name                   string                     `json:"name"`
	Amount                 Amount                     `json:"amount"`
	Enable                 bool                       `json:"enable"`
	ExpireTime             string                     `json:"expireTime"`
	Customer               PaymentCodeCustomer        `json:"customer"`
	USSDCode               string                     `json:"ussdCode"`
	Reference              string                     `json:"reference"`
	AuthorizedProviders    []string                   `json:"authorizedProviders"`
	AuthorizedPhoneNumber  string                     `json:"authorizedPhoneNumber"`
	RecurrentPaymentTarget PaymentCodeRecurrentTarget `json:"recurrentPaymentTarget"`
	FinancialAccountID     string                     `json:"financialAccountId"`
	ProcessedPaymentData   PaymentCodeProcessedData   `json:"processedPaymentData"`
	CreateTime             string                     `json:"createTime"`
	UpdateTime             string                     `json:"updateTime"`
	OwnershipGraph         PaymentCodeOwnershipGraph  `json:"ownershipGraph"`
	Metadata               map[string]any             `json:"metadata,omitempty"`
}

type PaymentCodeCustomer struct {
	Name string `json:"name"`
}

type PaymentCodeRecurrentTarget struct {
	ExpectedPaymentCount int    `json:"expectedPaymentCount"`
	ExpectedPaymentTotal Amount `json:"expectedPaymentTotal"`
}

type PaymentCodeProcessedData struct {
	Amount                        Amount                 `json:"amount"`
	OrderID                       string                 `json:"orderId"`
	PaymentID                     string                 `json:"paymentId"`
	OrderNumber                   string                 `json:"orderNumber"`
	ChannelData                   PaymentCodeChannelData `json:"channelData"`
	FinancialTransactionReference string                 `json:"financialTransactionReference"`
}

type PaymentCodeChannelData struct {
	ProviderID string `json:"providerId"`
	AccountID  string `json:"accountId"`
	Reference  string `json:"reference"`
}

type PaymentCodeOwnershipGraph struct {
	Owner PaymentCodeOwner `json:"owner"`
}

type PaymentCodeOwner struct {
	ID    string            `json:"id"`
	Type  string            `json:"type"`
	Owner *PaymentCodeOwner `json:"owner,omitempty"`
}

// PaymentCodeList is the paginated payment-code list response.
type PaymentCodeList struct {
	Result     []PaymentCode `json:"result"`
	Pagination Pagination    `json:"pagination"`
}

// Create creates a new recurrent USSD payment code.
func (s *PaymentCodeService) Create(ctx context.Context, params CreatePaymentCodeParams) (*PaymentCode, error) {
	if err := validateStruct(params); err != nil {
		return nil, err
	}

	minor := params.Amount * 100
	body := map[string]any{
		"name":                  params.PaymentName,
		"mode":                  "recurrent",
		"enable":                true,
		"amount":                Amount{Currency: "SLE", Value: minor},
		"duration":              "1h30m",
		"customer":              map[string]any{"name": params.Name},
		"reference":             "",
		"authorizedPhoneNumber": params.PhoneNumber,
		"recurrentPaymentTarget": map[string]any{
			"expectedPaymentCount": 1,
			"expectedPaymentTotal": Amount{Currency: "SLE", Value: minor},
		},
		"metadata": map[string]any{},
	}
	if params.FinancialAccountID != "" {
		body["financialAccountId"] = params.FinancialAccountID
	}

	key, err := generateIdempotencyKey()
	if err != nil {
		return nil, &Error{Message: err.Error()}
	}

	var out PaymentCode
	if err := s.client.do(ctx, requestOptions{
		method:         http.MethodPost,
		path:           paymentCodePath,
		body:           body,
		idempotencyKey: key,
		out:            &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// Retrieve fetches a single payment code by id.
func (s *PaymentCodeService) Retrieve(ctx context.Context, paymentCodeID string) (*PaymentCode, error) {
	if paymentCodeID == "" {
		return nil, newValidationError("paymentCodeID is required")
	}
	var out PaymentCode
	if err := s.client.do(ctx, requestOptions{
		method: http.MethodGet,
		path:   fmt.Sprintf("%s/%s", paymentCodePath, paymentCodeID),
		out:    &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns all payment codes.
func (s *PaymentCodeService) List(ctx context.Context) (*PaymentCodeList, error) {
	var out PaymentCodeList
	if err := s.client.do(ctx, requestOptions{
		method:  http.MethodGet,
		path:    paymentCodePath,
		out:     &out,
		rawBody: true,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update applies a partial update to a payment code. The fields accepted are
// those documented for PATCH /payment-codes/{id} (e.g. name, enable, metadata).
func (s *PaymentCodeService) Update(ctx context.Context, paymentCodeID string, params map[string]any) (*PaymentCode, error) {
	if paymentCodeID == "" {
		return nil, newValidationError("paymentCodeID is required")
	}
	if len(params) == 0 {
		return nil, newValidationError("at least one field is required to update")
	}

	key, err := generateIdempotencyKey()
	if err != nil {
		return nil, &Error{Message: err.Error()}
	}

	var out PaymentCode
	if err := s.client.do(ctx, requestOptions{
		method:         http.MethodPatch,
		path:           fmt.Sprintf("%s/%s", paymentCodePath, paymentCodeID),
		body:           params,
		idempotencyKey: key,
		out:            &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a payment code by id.
func (s *PaymentCodeService) Delete(ctx context.Context, paymentCodeID string) error {
	if paymentCodeID == "" {
		return newValidationError("paymentCodeID is required")
	}
	return s.client.do(ctx, requestOptions{
		method: http.MethodDelete,
		path:   fmt.Sprintf("%s/%s", paymentCodePath, paymentCodeID),
	})
}
