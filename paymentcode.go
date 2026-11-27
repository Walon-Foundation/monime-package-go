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

// Payment code modes and statuses.
const (
	PaymentCodeModeOneTime   = "one_time"
	PaymentCodeModeRecurrent = "recurrent"

	PaymentCodeStatusPending    = "pending"
	PaymentCodeStatusCancelled  = "cancelled"
	PaymentCodeStatusProcessing = "processing"
	PaymentCodeStatusExpired    = "expired"
	PaymentCodeStatusCompleted  = "completed"
)

// CreatePaymentCodeParams holds the inputs for creating a USSD payment code,
// as POST /v1/payment-codes defines them. Every field but Name is optional and
// is left out of the request when unset, so the API's defaults apply (mode
// one_time, enabled, a 10-minute duration).
type CreatePaymentCodeParams struct {
	// Name is the payment code's own name, 3 to 64 characters.
	Name string `validate:"required,min=3,max=64"`
	// Mode is PaymentCodeModeOneTime (the API's default) or
	// PaymentCodeModeRecurrent.
	Mode string `validate:"omitempty,oneof=one_time recurrent"`
	// Enable defaults to true when nil.
	Enable *bool
	// Amount is in minor units (cents), sent as-is. Currency defaults to SLE.
	Amount *Amount
	// Duration is how long the code stays valid, e.g. "10m" (the default) or
	// "1h30m".
	Duration string
	// CustomerName is shown to the payer, up to 100 characters.
	CustomerName string `validate:"max=100"`
	// Reference is your own identifier for the code, up to 64 characters.
	Reference string `validate:"max=64"`
	// AuthorizedProviders limits who can pay: PayoutProviderOrangeMoney ("m17")
	// and/or PayoutProviderAfricellMoney ("m18").
	AuthorizedProviders []string `validate:"omitempty,min=1,max=10,dive,oneof=m17 m18"`
	// AuthorizedPhoneNumber, when set, is the only number allowed to pay.
	AuthorizedPhoneNumber string
	// RecurrentPaymentTarget applies only to PaymentCodeModeRecurrent.
	RecurrentPaymentTarget *PaymentCodeRecurrentTarget
	// FinancialAccountID is the account the money goes to; when empty the
	// Space's main account is used.
	FinancialAccountID string `validate:"omitempty,max=64,startswith=fac-"`
	Metadata           map[string]any
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

// PaymentCodeRecurrentTarget is when a recurrent code is complete: after a
// number of payments, a total, or both.
type PaymentCodeRecurrentTarget struct {
	ExpectedPaymentCount int     `json:"expectedPaymentCount,omitempty"`
	ExpectedPaymentTotal *Amount `json:"expectedPaymentTotal,omitempty"`
}

type PaymentCodeProcessedData struct {
	Amount                        Amount                 `json:"amount"`
	OrderID                       string                 `json:"orderId"`
	PaymentID                     string                 `json:"paymentId"`
	OrderNumber                   string                 `json:"orderNumber"`
	ChannelData                   PaymentCodeChannelData `json:"channelData"`
	FinancialTransactionReference string                 `json:"financialTransactionReference"`
	Metadata                      map[string]any         `json:"metadata,omitempty"`
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
	ID       string            `json:"id"`
	Type     string            `json:"type"`
	Metadata map[string]any    `json:"metadata,omitempty"`
	Owner    *PaymentCodeOwner `json:"owner,omitempty"`
}

// PaymentCodeList is the paginated payment-code list response.
type PaymentCodeList struct {
	Result     []PaymentCode `json:"result"`
	Pagination Pagination    `json:"pagination"`
}

// normalizeAmount fills in the default currency and checks the value.
func normalizeAmount(field string, a *Amount) (*Amount, error) {
	if a == nil {
		return nil, nil
	}
	if a.Value <= 0 {
		return nil, newValidationError(field + ".value must be greater than zero (minor units)")
	}
	currency, err := currencyOrDefault(Currency(a.Currency))
	if err != nil {
		return nil, err
	}
	return &Amount{Currency: string(currency), Value: a.Value}, nil
}

// Create creates a USSD payment code. Amounts are in minor units: SLE 25.00 is
// Value 2500.
func (s *PaymentCodeService) Create(ctx context.Context, params CreatePaymentCodeParams) (*PaymentCode, error) {
	if err := validateStruct(params); err != nil {
		return nil, err
	}
	if params.RecurrentPaymentTarget != nil && params.Mode != PaymentCodeModeRecurrent {
		return nil, newValidationError("recurrentPaymentTarget applies only to mode recurrent")
	}
	if err := validateMetadata(params.Metadata, false); err != nil {
		return nil, err
	}
	amount, err := normalizeAmount("amount", params.Amount)
	if err != nil {
		return nil, err
	}

	body := map[string]any{"name": params.Name}
	putIfSet(body, "mode", params.Mode)
	if params.Enable != nil {
		body["enable"] = *params.Enable
	}
	if amount != nil {
		body["amount"] = amount
	}
	putIfSet(body, "duration", params.Duration)
	if params.CustomerName != "" {
		body["customer"] = map[string]any{"name": params.CustomerName}
	}
	putIfSet(body, "reference", params.Reference)
	if len(params.AuthorizedProviders) > 0 {
		body["authorizedProviders"] = params.AuthorizedProviders
	}
	putIfSet(body, "authorizedPhoneNumber", params.AuthorizedPhoneNumber)
	if target := params.RecurrentPaymentTarget; target != nil {
		if target.ExpectedPaymentCount < 0 {
			return nil, newValidationError("recurrentPaymentTarget.expectedPaymentCount must be at least 1")
		}
		total, err := normalizeAmount("recurrentPaymentTarget.expectedPaymentTotal", target.ExpectedPaymentTotal)
		if err != nil {
			return nil, err
		}
		body["recurrentPaymentTarget"] = PaymentCodeRecurrentTarget{ExpectedPaymentCount: target.ExpectedPaymentCount, ExpectedPaymentTotal: total}
	}
	putIfSet(body, "financialAccountId", params.FinancialAccountID)
	if params.Metadata != nil {
		body["metadata"] = params.Metadata
	}

	key, err := idempotencyKey(ctx)
	if err != nil {
		return nil, err
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

// List returns a page of payment codes. Use WithLimit and WithAfter to page
// through the collection, and WithPaymentCodeStatus, WithPaymentCodeMode or
// WithPaymentCodeUssdCode to narrow it.
func (s *PaymentCodeService) List(ctx context.Context, opts ...ListOption) (*PaymentCodeList, error) {
	query, err := buildListQuery(opts)
	if err != nil {
		return nil, err
	}

	var out PaymentCodeList
	if err := s.client.do(ctx, requestOptions{
		method:  http.MethodGet,
		path:    paymentCodePath,
		query:   query,
		out:     &out,
		rawBody: true,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// paymentCodeUpdateFields are the fields PATCH /payment-codes/{id} accepts.
var paymentCodeUpdateFields = map[string]updateField{
	"name":                   {maxLen: 64},
	"amount":                 {},
	"duration":               {},
	"enable":                 {},
	"customer":               {},
	"reference":              {maxLen: 64},
	"authorizedProviders":    {},
	"authorizedPhoneNumber":  {},
	"recurrentPaymentTarget": {},
	"financialAccountId":     {maxLen: 64},
	"metadata":               {},
}

// Update applies a partial update to a payment code. The fields accepted are
// those documented for PATCH /payment-codes/{id}: name, amount (minor units),
// duration, enable, customer, reference, authorizedProviders,
// authorizedPhoneNumber, recurrentPaymentTarget, financialAccountId and
// metadata.
func (s *PaymentCodeService) Update(ctx context.Context, paymentCodeID string, params map[string]any) (*PaymentCode, error) {
	if paymentCodeID == "" {
		return nil, newValidationError("paymentCodeID is required")
	}
	if err := validateUpdate(params, paymentCodeUpdateFields); err != nil {
		return nil, err
	}

	key, err := idempotencyKey(ctx)
	if err != nil {
		return nil, err
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
