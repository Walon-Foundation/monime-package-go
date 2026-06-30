package monime

import (
	"context"
	"fmt"
	"net/http"
)

// PayoutService provides access to the payouts API.
type PayoutService struct {
	client *Client
}

// Payout returns the payouts service.
func (c *Client) Payout() *PayoutService {
	return &PayoutService{client: c}
}

const payoutPath = "/payouts"

// PayoutDestination is the discriminated destination of a payout. Type selects
// which of the union-specific fields applies: "momo" uses PhoneNumber, "bank"
// uses AccountNumber, and "wallet" uses WalletID. ProviderID identifies the
// concrete provider (e.g. m17, slb001, dw001).
type PayoutDestination struct {
	Type                 string `json:"type"`
	ProviderID           string `json:"providerId"`
	PhoneNumber          string `json:"phoneNumber,omitempty"`
	AccountNumber        string `json:"accountNumber,omitempty"`
	WalletID             string `json:"walletId,omitempty"`
	TransactionReference string `json:"transactionReference,omitempty"`
}

// CreatePayoutParams holds the inputs for creating a payout.
type CreatePayoutParams struct {
	Amount        int64             `validate:"gt=0"`
	SourceAccount string            `validate:"required"`
	Destination   PayoutDestination `validate:"-"`
	Metadata      map[string]any
}

// PayoutSource is the origin financial account of a payout.
type PayoutSource struct {
	FinancialAccountID   string `json:"financialAccountId"`
	TransactionReference string `json:"transactionReference"`
}

// PayoutFee is a single fee applied to a payout.
type PayoutFee struct {
	Code   string `json:"code"`
	Amount Amount `json:"amount"`
}

// PayoutFailureDetail describes why a payout failed, when applicable.
type PayoutFailureDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// PayoutOwner is a node in a payout's ownership graph.
type PayoutOwner struct {
	ID    string       `json:"id"`
	Type  string       `json:"type"`
	Owner *PayoutOwner `json:"owner,omitempty"`
}

// PayoutOwnershipGraph traces the ownership of a payout.
type PayoutOwnershipGraph struct {
	Owner PayoutOwner `json:"owner"`
}

// Payout is the payout resource returned by the API.
type Payout struct {
	ID             string               `json:"id"`
	Status         string               `json:"status"`
	Amount         Amount               `json:"amount"`
	Source         PayoutSource         `json:"source"`
	Destination    PayoutDestination    `json:"destination"`
	Fees           []PayoutFee          `json:"fees"`
	FailureDetail  PayoutFailureDetail  `json:"failureDetail"`
	CreateTime     string               `json:"createTime"`
	UpdateTime     string               `json:"updateTime"`
	OwnershipGraph PayoutOwnershipGraph `json:"ownershipGraph"`
	Metadata       map[string]any       `json:"metadata,omitempty"`
}

// PayoutList is the paginated payout list response.
type PayoutList struct {
	Result     []Payout   `json:"result"`
	Pagination Pagination `json:"pagination"`
}

// validatePayoutDestination enforces the discriminated-union shape: Type must be
// one of momo/bank/wallet and ProviderID must be present.
func validatePayoutDestination(d PayoutDestination) error {
	switch d.Type {
	case "momo", "bank", "wallet":
	default:
		return newValidationError("destination.type must be one of momo, bank, wallet")
	}
	if d.ProviderID == "" {
		return newValidationError("destination.providerId is required")
	}
	return nil
}

// Create creates a new payout. Unlike payment codes, the amount is sent as-is in
// minor units and is not multiplied (matching the Monime API and the TS SDK).
func (s *PayoutService) Create(ctx context.Context, params CreatePayoutParams) (*Payout, error) {
	if err := validateStruct(params); err != nil {
		return nil, err
	}
	if err := validatePayoutDestination(params.Destination); err != nil {
		return nil, err
	}

	metadata := params.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}

	body := map[string]any{
		"amount":      Amount{Currency: "SLE", Value: params.Amount},
		"source":      map[string]any{"financialAccountId": params.SourceAccount},
		"destination": params.Destination,
		"metadata":    metadata,
	}

	key, err := generateIdempotencyKey()
	if err != nil {
		return nil, &Error{Message: err.Error()}
	}

	var out Payout
	if err := s.client.do(ctx, requestOptions{
		method:         http.MethodPost,
		path:           payoutPath,
		body:           body,
		idempotencyKey: key,
		out:            &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// Retrieve fetches a single payout by id.
func (s *PayoutService) Retrieve(ctx context.Context, payoutID string) (*Payout, error) {
	if payoutID == "" {
		return nil, newValidationError("payoutID is required")
	}
	var out Payout
	if err := s.client.do(ctx, requestOptions{
		method: http.MethodGet,
		path:   fmt.Sprintf("%s/%s", payoutPath, payoutID),
		out:    &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns all payouts.
func (s *PayoutService) List(ctx context.Context) (*PayoutList, error) {
	var out PayoutList
	if err := s.client.do(ctx, requestOptions{
		method:  http.MethodGet,
		path:    payoutPath,
		out:     &out,
		rawBody: true,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update applies a partial update to a payout. The fields accepted are those
// documented for PATCH /payouts/{id} (notably metadata); fields can only be
// modified while the payout has not yet been processed.
func (s *PayoutService) Update(ctx context.Context, payoutID string, params map[string]any) (*Payout, error) {
	if payoutID == "" {
		return nil, newValidationError("payoutID is required")
	}
	if len(params) == 0 {
		return nil, newValidationError("at least one field is required to update")
	}

	key, err := generateIdempotencyKey()
	if err != nil {
		return nil, &Error{Message: err.Error()}
	}

	var out Payout
	if err := s.client.do(ctx, requestOptions{
		method:         http.MethodPatch,
		path:           fmt.Sprintf("%s/%s", payoutPath, payoutID),
		body:           params,
		idempotencyKey: key,
		out:            &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete deletes or cancels a payout by id.
func (s *PayoutService) Delete(ctx context.Context, payoutID string) error {
	if payoutID == "" {
		return newValidationError("payoutID is required")
	}
	return s.client.do(ctx, requestOptions{
		method: http.MethodDelete,
		path:   fmt.Sprintf("%s/%s", payoutPath, payoutID),
	})
}
