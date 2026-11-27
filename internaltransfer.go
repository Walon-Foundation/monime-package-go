package monime

import (
	"context"
	"fmt"
	"net/http"
)

// InternalTransferService provides access to the internal-transfers API.
type InternalTransferService struct {
	client *Client
}

// InternalTransfer returns the internal-transfers service.
func (c *Client) InternalTransfer() *InternalTransferService {
	return &InternalTransferService{client: c}
}

const internalTransferPath = "/internal-transfers"

// CreateInternalTransferParams holds the inputs for creating an internal
// transfer between two financial accounts. It mirrors the zod validator in the
// TypeScript SDK: source and destination accounts are required and the amount
// must be greater than zero.
type CreateInternalTransferParams struct {
	// SourceAccount is the financial account id to debit.
	SourceAccount string `validate:"required,max=64"`
	// DestinationAccount is the financial account id to credit.
	DestinationAccount string `validate:"required,max=64"`
	// Amount is in minor units, forwarded as amount.value.
	Amount int64 `validate:"gt=0"`
	// Currency defaults to SLE; the API also supports USD.
	Currency Currency
	// Description is optional free-form context (up to 150 characters).
	Description string `validate:"max=150"`
	// Metadata is optional key-value context attached to the transfer.
	Metadata map[string]any
}

// InternalTransfer is the internal-transfer resource returned by the API.
type InternalTransfer struct {
	ID                            string                         `json:"id"`
	Status                        string                         `json:"status"`
	Amount                        Amount                         `json:"amount"`
	SourceFinancialAccount        InternalTransferAccount        `json:"sourceFinancialAccount"`
	DestinationFinancialAccount   InternalTransferAccount        `json:"destinationFinancialAccount"`
	FinancialTransactionReference string                         `json:"financialTransactionReference"`
	Description                   string                         `json:"description"`
	FailureDetail                 InternalTransferFailureDetail  `json:"failureDetail"`
	OwnershipGraph                InternalTransferOwnershipGraph `json:"ownershipGraph"`
	CreateTime                    string                         `json:"createTime"`
	UpdateTime                    string                         `json:"updateTime"`
	Metadata                      map[string]any                 `json:"metadata,omitempty"`
}

// InternalTransferAccount identifies a financial account by id.
type InternalTransferAccount struct {
	ID string `json:"id"`
}

// InternalTransferFailureDetail describes why a transfer failed, when applicable.
type InternalTransferFailureDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// InternalTransferOwnershipGraph captures the ownership chain of a transfer.
type InternalTransferOwnershipGraph struct {
	Owner InternalTransferOwner `json:"owner"`
}

// InternalTransferOwner is a node in the ownership graph.
type InternalTransferOwner struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Metadata map[string]any         `json:"metadata,omitempty"`
	Owner    *InternalTransferOwner `json:"owner,omitempty"`
}

// Internal transfer statuses and failure codes.
const (
	InternalTransferStatusPending    = "pending"
	InternalTransferStatusProcessing = "processing"
	InternalTransferStatusCompleted  = "completed"
	InternalTransferStatusFailed     = "failed"

	InternalTransferFailureUnknown          = "unknown"
	InternalTransferFailureFundInsufficient = "fund_insufficient"
)

// InternalTransferList is the paginated internal-transfer list response.
type InternalTransferList struct {
	Result     []InternalTransfer `json:"result"`
	Pagination Pagination         `json:"pagination"`
}

// Create creates a new internal transfer between financial accounts.
func (s *InternalTransferService) Create(ctx context.Context, params CreateInternalTransferParams) (*InternalTransfer, error) {
	if err := validateStruct(params); err != nil {
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
		"amount":                      Amount{Currency: string(currency), Value: params.Amount},
		"sourceFinancialAccount":      map[string]any{"id": params.SourceAccount},
		"destinationFinancialAccount": map[string]any{"id": params.DestinationAccount},
	}
	putIfSet(body, "description", params.Description)
	if params.Metadata != nil {
		body["metadata"] = params.Metadata
	}

	key, err := idempotencyKey(ctx)
	if err != nil {
		return nil, err
	}

	var out InternalTransfer
	if err := s.client.do(ctx, requestOptions{
		method:         http.MethodPost,
		path:           internalTransferPath,
		body:           body,
		idempotencyKey: key,
		out:            &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns a page of internal transfers. Use WithLimit and WithAfter to
// page through the collection, and WithInternalTransferStatus,
// WithInternalTransferSourceAccount, WithInternalTransferDestinationAccount or
// WithInternalTransferTransactionReference to narrow it.
func (s *InternalTransferService) List(ctx context.Context, opts ...ListOption) (*InternalTransferList, error) {
	query, err := buildListQuery(opts)
	if err != nil {
		return nil, err
	}

	var out InternalTransferList
	if err := s.client.do(ctx, requestOptions{
		method:  http.MethodGet,
		path:    internalTransferPath,
		query:   query,
		out:     &out,
		rawBody: true,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// Retrieve fetches a single internal transfer by id.
func (s *InternalTransferService) Retrieve(ctx context.Context, internalTransferID string) (*InternalTransfer, error) {
	if internalTransferID == "" {
		return nil, newValidationError("internalTransferID is required")
	}
	var out InternalTransfer
	if err := s.client.do(ctx, requestOptions{
		method: http.MethodGet,
		path:   fmt.Sprintf("%s/%s", internalTransferPath, internalTransferID),
		out:    &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update applies a partial update to an internal transfer. The documented
// PATCH /internal-transfers/{id} endpoint accepts the mutable fields
// description (up to 255 characters) and metadata.
func (s *InternalTransferService) Update(ctx context.Context, internalTransferID string, params map[string]any) (*InternalTransfer, error) {
	if internalTransferID == "" {
		return nil, newValidationError("internalTransferID is required")
	}
	if err := validateUpdate(params, map[string]updateField{"description": {maxLen: 255}, "metadata": {}}); err != nil {
		return nil, err
	}

	key, err := idempotencyKey(ctx)
	if err != nil {
		return nil, err
	}

	var out InternalTransfer
	if err := s.client.do(ctx, requestOptions{
		method:         http.MethodPatch,
		path:           fmt.Sprintf("%s/%s", internalTransferPath, internalTransferID),
		body:           params,
		idempotencyKey: key,
		out:            &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes or cancels an internal transfer by id.
func (s *InternalTransferService) Delete(ctx context.Context, internalTransferID string) error {
	if internalTransferID == "" {
		return newValidationError("internalTransferID is required")
	}
	return s.client.do(ctx, requestOptions{
		method: http.MethodDelete,
		path:   fmt.Sprintf("%s/%s", internalTransferPath, internalTransferID),
	})
}
