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

// Receipt statuses.
const (
	ReceiptStatusNotRedeemed       = "not_redeemed"
	ReceiptStatusPartiallyRedeemed = "partially_redeemed"
	ReceiptStatusFullyRedeemed     = "fully_redeemed"
)

// RedeemReceiptParams holds the inputs for POST /receipts/{orderNumber}/redeem:
// either RedeemAll, or the entitlements to redeem and how many units of each.
type RedeemReceiptParams struct {
	RedeemAll    bool
	Entitlements []ReceiptRedemptionItem `validate:"dive"`
	Metadata     map[string]any
}

// ReceiptRedemptionItem redeems Units (at least 1; zero means 1) of the
// entitlement named Key.
type ReceiptRedemptionItem struct {
	Key   string `json:"key" validate:"required,min=1,max=100"`
	Units int    `json:"units,omitempty" validate:"omitempty,min=1"`
}

// ReceiptRedemption is what a redemption returns: whether it was applied, and
// the receipt as it now stands.
type ReceiptRedemption struct {
	Redeem  bool    `json:"redeem"`
	Receipt Receipt `json:"receipt"`
}

// Redeem redeems one or more entitlements from a receipt. When processing a
// bulk redemption, if any entitlement is already exhausted the entire
// operation is rejected.
func (s *ReceiptService) Redeem(ctx context.Context, orderNumber string, params RedeemReceiptParams) (*ReceiptRedemption, error) {
	if orderNumber == "" || len(orderNumber) > 20 {
		return nil, newValidationError("orderNumber is required and is at most 20 characters")
	}
	if !params.RedeemAll && len(params.Entitlements) == 0 {
		return nil, newValidationError("set RedeemAll or list the entitlements to redeem")
	}
	if err := validateStruct(params); err != nil {
		return nil, err
	}
	if err := validateMetadata(params.Metadata, false); err != nil {
		return nil, err
	}
	body := map[string]any{}
	if params.RedeemAll {
		body["redeemAll"] = true
	}
	if len(params.Entitlements) > 0 {
		body["entitlements"] = params.Entitlements
	}
	if params.Metadata != nil {
		body["metadata"] = params.Metadata
	}

	key, err := idempotencyKey(ctx)
	if err != nil {
		return nil, err
	}

	var out ReceiptRedemption
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
