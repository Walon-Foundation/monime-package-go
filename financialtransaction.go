package monime

import (
	"context"
	"fmt"
	"net/http"
)

// FinancialTransactionService provides access to the financial-transactions API.
type FinancialTransactionService struct {
	client *Client
}

// FinancialTransaction returns the financial-transactions service.
func (c *Client) FinancialTransaction() *FinancialTransactionService {
	return &FinancialTransactionService{client: c}
}

const financialTransactionPath = "/financial-transactions"

// FinancialTransaction is the financial-transaction resource returned by the API.
type FinancialTransaction struct {
	ID                  string                                  `json:"id"`
	Type                string                                  `json:"type"`
	Amount              Amount                                  `json:"amount"`
	Timestamp           string                                  `json:"timestamp"`
	Reference           string                                  `json:"reference"`
	FinancialAccount    FinancialTransactionFinancialAccount    `json:"financialAccount"`
	OriginatingReversal FinancialTransactionOriginatingReversal `json:"originatingReversal"`
	OriginatingFee      FinancialTransactionOriginatingFee      `json:"originatingFee"`
	OwnershipGraph      FinancialTransactionOwnershipGraph      `json:"ownershipGraph"`
	Metadata            map[string]any                          `json:"metadata,omitempty"`
}

type FinancialTransactionFinancialAccount struct {
	ID      string                      `json:"id"`
	Balance FinancialTransactionBalance `json:"balance"`
}

type FinancialTransactionBalance struct {
	After Amount `json:"after"`
}

type FinancialTransactionOriginatingReversal struct {
	OriginTxnID  string `json:"originTxnId"`
	OriginTxnRef string `json:"originTxnRef"`
}

type FinancialTransactionOriginatingFee struct {
	Code string `json:"code"`
}

type FinancialTransactionOwnershipGraph struct {
	Owner FinancialTransactionOwner `json:"owner"`
}

type FinancialTransactionOwner struct {
	ID    string                     `json:"id"`
	Type  string                     `json:"type"`
	Owner *FinancialTransactionOwner `json:"owner,omitempty"`
}

// FinancialTransactionList is the paginated financial-transaction list response.
type FinancialTransactionList struct {
	Result     []FinancialTransaction `json:"result"`
	Pagination Pagination             `json:"pagination"`
}

// Retrieve fetches a single financial transaction by id.
func (s *FinancialTransactionService) Retrieve(ctx context.Context, financialTransactionID string) (*FinancialTransaction, error) {
	if financialTransactionID == "" {
		return nil, newValidationError("financialTransactionID is required")
	}
	var out FinancialTransaction
	if err := s.client.do(ctx, requestOptions{
		method: http.MethodGet,
		path:   fmt.Sprintf("%s/%s", financialTransactionPath, financialTransactionID),
		out:    &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns all financial transactions.
func (s *FinancialTransactionService) List(ctx context.Context) (*FinancialTransactionList, error) {
	var out FinancialTransactionList
	if err := s.client.do(ctx, requestOptions{
		method:  http.MethodGet,
		path:    financialTransactionPath,
		out:     &out,
		rawBody: true,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}
