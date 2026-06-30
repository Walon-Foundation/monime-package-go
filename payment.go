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

// PaymentChannel describes the channel a payment was made through.
type PaymentChannel struct {
	Type string `json:"type"`
}

// PaymentFee is a single fee applied to a payment.
type PaymentFee struct {
	Code   string `json:"code"`
	Amount Amount `json:"amount"`
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

// List returns all payments.
func (s *PaymentService) List(ctx context.Context) (*PaymentList, error) {
	var out PaymentList
	if err := s.client.do(ctx, requestOptions{
		method:  http.MethodGet,
		path:    paymentPath,
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
	if len(params) == 0 {
		return nil, newValidationError("at least one field is required to update")
	}

	key, err := generateIdempotencyKey()
	if err != nil {
		return nil, &Error{Message: err.Error()}
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
