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
	SourceAccount string `validate:"required"`
	// DestinationAccount is the financial account id to credit.
	DestinationAccount string `validate:"required"`
	// Amount is the value forwarded as amount.value (currency is fixed to SLE).
	Amount int64 `validate:"gt=0"`
	// Description is optional free-form context (up to 150 characters).
	Description string
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
	ID    string                 `json:"id"`
	Type  string                 `json:"type"`
	Owner *InternalTransferOwner `json:"owner,omitempty"`
}

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

	description := params.Description
	metadata := params.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}

	body := map[string]any{
		"amount":                      Amount{Currency: "SLE", Value: params.Amount},
		"sourceFinancialAccount":      map[string]any{"id": params.SourceAccount},
		"destinationFinancialAccount": map[string]any{"id": params.DestinationAccount},
		"description":                 description,
		"metadata":                    metadata,
	}

	key, err := generateIdempotencyKey()
	if err != nil {
		return nil, &Error{Message: err.Error()}
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

// List returns all internal transfers.
func (s *InternalTransferService) List(ctx context.Context) (*InternalTransferList, error) {
	var out InternalTransferList
	if err := s.client.do(ctx, requestOptions{
		method:  http.MethodGet,
		path:    internalTransferPath,
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
// description and metadata. This method has no TypeScript SDK counterpart; it
// is added from the official API documentation.
func (s *InternalTransferService) Update(ctx context.Context, internalTransferID string, params map[string]any) (*InternalTransfer, error) {
	if internalTransferID == "" {
		return nil, newValidationError("internalTransferID is required")
	}
	if len(params) == 0 {
		return nil, newValidationError("at least one field is required to update")
	}

	key, err := generateIdempotencyKey()
	if err != nil {
		return nil, &Error{Message: err.Error()}
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
