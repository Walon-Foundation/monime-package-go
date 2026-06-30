package monime

import (
	"context"
	"fmt"
	"net/http"
)

// ReceiptService provides access to the receipts API.
type ReceiptService struct {
	client *Client
}

// Receipt returns the receipts service.
func (c *Client) Receipt() *ReceiptService {
	return &ReceiptService{client: c}
}

const receiptPath = "/receipts"

// Receipt is a record of entitlements generated when a payment completes. It
// links a payment (via an order) to redeemable entitlement claims.
type Receipt struct {
	Status       string               `json:"status"`
	OrderName    string               `json:"orderName"`
	OrderNumber  string               `json:"orderNumber"`
	OrderAmount  Amount               `json:"orderAmount"`
	CreateTime   string               `json:"createTime"`
	UpdateTime   string               `json:"updateTime"`
	Entitlements []ReceiptEntitlement `json:"entitlements"`
	Metadata     map[string]string    `json:"metadata,omitempty"`
}

// ReceiptEntitlement is a right granted by a receipt that can be consumed
// through redemption until exhausted.
type ReceiptEntitlement struct {
	Key       string `json:"key"`
	Name      string `json:"name"`
	Limit     int    `json:"limit"`
	Current   int    `json:"current"`
	Remaining int    `json:"remaining"`
	Exhausted bool   `json:"exhausted"`
}

// Retrieve fetches a single receipt by its order number.
func (s *ReceiptService) Retrieve(ctx context.Context, orderNumber string) (*Receipt, error) {
	if orderNumber == "" {
		return nil, newValidationError("orderNumber is required")
	}
	var out Receipt
	if err := s.client.do(ctx, requestOptions{
		method: http.MethodGet,
		path:   fmt.Sprintf("%s/%s", receiptPath, orderNumber),
		out:    &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// Redeem redeems one or more entitlements from a receipt. When processing a
// bulk redemption, if any entitlement is already exhausted the entire
// operation is rejected.
func (s *ReceiptService) Redeem(ctx context.Context, orderNumber string, body map[string]any) (*Receipt, error) {
	if orderNumber == "" {
		return nil, newValidationError("orderNumber is required")
	}
	if len(body) == 0 {
		return nil, newValidationError("body is required to redeem a receipt")
	}

	key, err := generateIdempotencyKey()
	if err != nil {
		return nil, &Error{Message: err.Error()}
	}

	var out Receipt
	if err := s.client.do(ctx, requestOptions{
		method:         http.MethodPost,
		path:           fmt.Sprintf("%s/%s/redeem", receiptPath, orderNumber),
		body:           body,
		idempotencyKey: key,
		out:            &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}
