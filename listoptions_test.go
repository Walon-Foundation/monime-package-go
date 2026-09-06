package monime

import (
	"errors"
	"testing"
)

func TestBuildListQuery(t *testing.T) {
	tests := []struct {
		name string
		opts []ListOption
		want string
	}{
		{
			name: "no options sends nothing",
			opts: nil,
			want: "",
		},
		{
			name: "paging",
			opts: []ListOption{WithLimit(50), WithAfter("cur-1")},
			want: "after=cur-1&limit=50",
		},
		{
			name: "empty cursor is dropped",
			opts: []ListOption{WithAfter("")},
			want: "",
		},
		{
			name: "nil option is skipped",
			opts: []ListOption{nil, WithLimit(1)},
			want: "limit=1",
		},
		{
			name: "payout filters",
			opts: []ListOption{
				WithPayoutStatus("completed"),
				WithPayoutSourceAccount("fac-1"),
				WithPayoutSourceTransactionReference("src-ref"),
				WithPayoutDestinationTransactionReference("dst-ref"),
			},
			want: "destinationTransactionReference=dst-ref&sourceFinancialAccountId=fac-1&sourceTransactionReference=src-ref&status=completed",
		},
		{
			name: "payment code filters use the documented parameter names",
			opts: []ListOption{
				WithPaymentCodeUssdCode("*715*1#"),
				WithPaymentCodeMode("one_time"),
				WithPaymentCodeStatus("expired"),
			},
			want: "mode=one_time&status=expired&ussd_code=%2A715%2A1%23",
		},
		{
			name: "financial account filters",
			opts: []ListOption{
				WithFinancialAccountUvan("uvan-1"),
				WithFinancialAccountReference("ref-1"),
				WithFinancialAccountBalance(true),
			},
			want: "reference=ref-1&uvan=uvan-1&withBalance=true",
		},
		{
			name: "withBalance false is still sent",
			opts: []ListOption{WithFinancialAccountBalance(false)},
			want: "withBalance=false",
		},
		{
			name: "financial transaction filters",
			opts: []ListOption{
				WithFinancialTransactionAccountID("fac-1"),
				WithFinancialTransactionReference("ref-1"),
				WithFinancialTransactionType("debit"),
			},
			want: "financialAccountId=fac-1&reference=ref-1&type=debit",
		},
		{
			name: "internal transfer filters",
			opts: []ListOption{
				WithInternalTransferStatus("pending"),
				WithInternalTransferSourceAccount("fac-1"),
				WithInternalTransferDestinationAccount("fac-2"),
				WithInternalTransferTransactionReference("ftr-1"),
			},
			want: "destinationFinancialAccountId=fac-2&financialTransactionReference=ftr-1&sourceFinancialAccountId=fac-1&status=pending",
		},
		{
			name: "payment filters",
			opts: []ListOption{
				WithPaymentOrderNumber("ord-1"),
				WithPaymentFinancialAccountID("fac-1"),
				WithPaymentFinancialTransactionReference("ftr-1"),
			},
			want: "financialAccountId=fac-1&financialTransactionReference=ftr-1&orderNumber=ord-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildListQuery(tt.opts)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Encode() != tt.want {
				t.Fatalf("query = %q, want %q", got.Encode(), tt.want)
			}
		})
	}
}

func TestBuildListQuery_Validation(t *testing.T) {
	tests := []struct {
		name string
		opt  ListOption
	}{
		{"limit below range", WithLimit(0)},
		{"limit above range", WithLimit(51)},
		{"negative limit", WithLimit(-1)},
		{"unknown payout status", WithPayoutStatus("refunded")},
		{"unknown payment code mode", WithPaymentCodeMode("oneTime")},
		{"unknown payment code status", WithPaymentCodeStatus("paid")},
		{"unknown transaction type", WithFinancialTransactionType("transfer")},
		{"unknown internal transfer status", WithInternalTransferStatus("done")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := buildListQuery([]ListOption{tt.opt})
			if err == nil {
				t.Fatal("expected a validation error")
			}
			var valErr *ValidationError
			if !errors.As(err, &valErr) {
				t.Fatalf("expected *ValidationError, got %T", err)
			}
		})
	}
}

func TestBuildListQuery_ReportsFirstFailure(t *testing.T) {
	_, err := buildListQuery([]ListOption{
		WithPayoutStatus("bogus"),
		WithLimit(999),
	})
	if err == nil {
		t.Fatal("expected a validation error")
	}
	if got := err.Error(); got != "monime: status must be one of pending, processing, failed, completed (status 400)" {
		t.Fatalf("error = %q, want the first failure", got)
	}
}
