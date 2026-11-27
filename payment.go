package monime

import (
	"context"
	"fmt"
	"net/http"
)

// PaymentService provides access to the payments API.
type PaymentService struct {
	client *Client
}

// Payment returns the payments service.
func (c *Client) Payment() *PaymentService {
	return &PaymentService{client: c}
}

const paymentPath = "/payments"

// Payment is the payment resource returned by the API.
type Payment struct {
	ID                            string                `json:"id"`
	Status                        string                `json:"status"`
	Amount                        Amount                `json:"amount"`
	Channel                       PaymentChannel        `json:"channel"`
	Name                          string                `json:"name"`
	Reference                     string                `json:"reference"`
	OrderNumber                   string                `json:"orderNumber"`
	FinancialAccountID            string                `json:"financialAccountId"`
	FinancialTransactionReference string                `json:"financialTransactionReference"`
	Fees                          []PaymentFee          `json:"fees"`
	CreateTime                    string                `json:"createTime"`
	UpdateTime                    string                `json:"updateTime"`
	OwnershipGraph                PaymentOwnershipGraph `json:"ownershipGraph"`
	Metadata                      map[string]any        `json:"metadata,omitempty"`
}

// Payment statuses. The API has no failed state for payments.
const (
	PaymentStatusPending    = "pending"
	PaymentStatusProcessing = "processing"
	PaymentStatusCompleted  = "completed"
)

// PaymentChannel describes the channel a payment was made through. Type is
// bank, card, momo or wallet, and selects which fields are set: bank uses
// AccountNumber, momo PhoneNumber, wallet WalletID (each with Provider and
// Reference), and card Scheme and Last4.
type PaymentChannel struct {
	Type          string         `json:"type"`
	Provider      string         `json:"provider,omitempty"`
	Reference     string         `json:"reference,omitempty"`
	AccountNumber string         `json:"accountNumber,omitempty"`
	PhoneNumber   string         `json:"phoneNumber,omitempty"`
	WalletID      string         `json:"walletId,omitempty"`
	Scheme        string         `json:"scheme,omitempty"` // visa or mastercard
	Last4         string         `json:"last4,omitempty"`
	Fingerprint   string         `json:"fingerprint,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
}

// PaymentFee is a single fee applied to a payment.
type PaymentFee struct {
	Code     string         `json:"code"`
	Amount   Amount         `json:"amount"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// PaymentOwnershipGraph describes the ownership chain of a payment.
type PaymentOwnershipGraph struct {
	Owner PaymentOwner `json:"owner"`
}

// PaymentOwner is the owner of a payment within the ownership graph.
type PaymentOwner struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Metadata map[string]any `json:"metadata,omitempty"`
	Owner    *PaymentOwner  `json:"owner,omitempty"`
}

// PaymentList is the paginated payment list response.
type PaymentList struct {
	Result     []Payment  `json:"result"`
	Pagination Pagination `json:"pagination"`
}

// Retrieve fetches a single payment by id.
func (s *PaymentService) Retrieve(ctx context.Context, paymentID string) (*Payment, error) {
	if paymentID == "" {
		return nil, newValidationError("paymentID is required")
	}
	var out Payment
	if err := s.client.do(ctx, requestOptions{
		method: http.MethodGet,
		path:   fmt.Sprintf("%s/%s", paymentPath, paymentID),
		out:    &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns a page of payments. Use WithLimit and WithAfter to page through
// the collection, and WithPaymentOrderNumber, WithPaymentFinancialAccountID or
// WithPaymentFinancialTransactionReference to narrow it.
func (s *PaymentService) List(ctx context.Context, opts ...ListOption) (*PaymentList, error) {
	query, err := buildListQuery(opts)
	if err != nil {
		return nil, err
	}

	var out PaymentList
	if err := s.client.do(ctx, requestOptions{
		method:  http.MethodGet,
		path:    paymentPath,
		query:   query,
		out:     &out,
		rawBody: true,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update applies a partial update to a payment. The fields accepted are those
// documented for PATCH /payments/{id}.
func (s *PaymentService) Update(ctx context.Context, paymentID string, params map[string]any) (*Payment, error) {
	if paymentID == "" {
		return nil, newValidationError("paymentID is required")
	}
	// PATCH /payments/{id} accepts name (100 characters) and metadata only.
	if err := validateUpdate(params, map[string]updateField{"name": {maxLen: 100}, "metadata": {}}); err != nil {
		return nil, err
	}

	key, err := idempotencyKey(ctx)
	if err != nil {
		return nil, err
	}

	var out Payment
	if err := s.client.do(ctx, requestOptions{
		method:         http.MethodPatch,
		path:           fmt.Sprintf("%s/%s", paymentPath, paymentID),
		body:           params,
		idempotencyKey: key,
		out:            &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}
