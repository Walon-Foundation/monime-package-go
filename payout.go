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

// Provider ids the API's payout schema lists. Monime adds providers over time,
// so other ids are passed through; read the live list with
// FinancialProvider().ListBanks / ListMomos.
const (
	PayoutProviderSLCB          = "slb001" // Sierra Leone Commercial Bank
	PayoutProviderGTBank        = "slb004" // Guaranty Trust Bank
	PayoutProviderAccessBank    = "slb007" // Access Bank
	PayoutProviderOrangeMoney   = "m17"
	PayoutProviderAfricellMoney = "m18"
	PayoutProviderMonimeWallet  = "dw001"
)

// Payout statuses.
const (
	PayoutStatusPending    = "pending"
	PayoutStatusProcessing = "processing"
	PayoutStatusCompleted  = "completed"
	PayoutStatusFailed     = "failed"
)

// Payout failure codes, from failureDetail.code.
const (
	PayoutFailureUnknown                       = "unknown"
	PayoutFailureFundInsufficient              = "fund_insufficient"
	PayoutFailureAuthorizationFailed           = "authorization_failed"
	PayoutFailureProviderUnknown               = "provider_unknown"
	PayoutFailureProviderAccountBlocked        = "provider_account_blocked"
	PayoutFailureProviderAccountMissing        = "provider_account_missing"
	PayoutFailureProviderAccountQuotaExhausted = "provider_account_quota_exhausted"
)

// CreatePayoutParams holds the inputs for creating a payout.
type CreatePayoutParams struct {
	// Amount is in minor units (cents), sent as-is.
	Amount int64 `validate:"gt=0"`
	// Currency defaults to SLE; the API also supports USD.
	Currency Currency
	// SourceAccount is the financial account to pay from. Leave it empty to pay
	// from the Space's main account, as the API does when source is omitted.
	SourceAccount string
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
	Code     string         `json:"code"`
	Amount   Amount         `json:"amount"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// PayoutFailureDetail describes why a payout failed, when applicable.
type PayoutFailureDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// PayoutOwner is a node in a payout's ownership graph.
type PayoutOwner struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Metadata map[string]any `json:"metadata,omitempty"`
	Owner    *PayoutOwner   `json:"owner,omitempty"`
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

// payoutDestinationBody validates the destination's discriminated union and
// returns only the fields its type allows. The API rejects unknown properties on
// each variant, and transactionReference is read-only, so nothing else is sent:
// bank needs accountNumber, momo needs phoneNumber, wallet takes an optional
// walletId.
func payoutDestinationBody(d PayoutDestination) (map[string]any, error) {
	if d.ProviderID == "" {
		return nil, newValidationError("destination.providerId is required")
	}
	body := map[string]any{"type": d.Type, "providerId": d.ProviderID}
	switch d.Type {
	case "bank":
		if d.AccountNumber == "" {
			return nil, newValidationError("destination.accountNumber is required for a bank payout")
		}
		body["accountNumber"] = d.AccountNumber
	case "momo":
		if d.PhoneNumber == "" {
			return nil, newValidationError("destination.phoneNumber is required for a momo payout")
		}
		body["phoneNumber"] = d.PhoneNumber
	case "wallet":
		putIfSet(body, "walletId", d.WalletID)
	default:
		return nil, newValidationError("destination.type must be one of momo, bank, wallet")
	}
	return body, nil
}

// Create creates a new payout. The amount is in minor units and sent as-is.
// Pass a key with WithIdempotencyKey so a retry can't pay twice.
func (s *PayoutService) Create(ctx context.Context, params CreatePayoutParams) (*Payout, error) {
	if err := validateStruct(params); err != nil {
		return nil, err
	}
	destination, err := payoutDestinationBody(params.Destination)
	if err != nil {
		return nil, err
	}
	currency, err := currencyOrDefault(params.Currency)
	if err != nil {
		return nil, err
	}
	if err := validateMetadata(params.Metadata, false); err != nil {
		return nil, err
	}

	body := map[string]any{
		"amount":      Amount{Currency: string(currency), Value: params.Amount},
		"destination": destination,
	}
	if params.SourceAccount != "" {
		body["source"] = map[string]any{"financialAccountId": params.SourceAccount}
	}
	if params.Metadata != nil {
		body["metadata"] = params.Metadata
	}

	key, err := idempotencyKey(ctx)
	if err != nil {
		return nil, err
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

// List returns a page of payouts. Use WithLimit and WithAfter to page through
// the collection, and WithPayoutStatus, WithPayoutSourceAccount,
// WithPayoutSourceTransactionReference or
// WithPayoutDestinationTransactionReference to narrow it.
func (s *PayoutService) List(ctx context.Context, opts ...ListOption) (*PayoutList, error) {
	query, err := buildListQuery(opts)
	if err != nil {
		return nil, err
	}

	var out PayoutList
	if err := s.client.do(ctx, requestOptions{
		method:  http.MethodGet,
		path:    payoutPath,
		query:   query,
		out:     &out,
		rawBody: true,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update applies a partial update to a payout. PATCH /payouts/{id} accepts only
// metadata, and only before the payout is processed.
func (s *PayoutService) Update(ctx context.Context, payoutID string, params map[string]any) (*Payout, error) {
	if payoutID == "" {
		return nil, newValidationError("payoutID is required")
	}
	if err := validateUpdate(params, map[string]updateField{"metadata": {}}); err != nil {
		return nil, err
	}

	key, err := idempotencyKey(ctx)
	if err != nil {
		return nil, err
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
