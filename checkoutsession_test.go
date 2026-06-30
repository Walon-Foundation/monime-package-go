package monime

import (
	"errors"
	"testing"
)

func TestCreateCheckoutSessionParams_Validation(t *testing.T) {
	valid := CreateCheckoutSessionParams{
		Name:       "Order",
		Amount:     1000,
		Quantity:   1,
		SuccessURL: "https://example.com/success",
		CancelURL:  "https://example.com/cancel",
	}

	tests := []struct {
		name    string
		params  CreateCheckoutSessionParams
		wantErr bool
	}{
		{
			name:    "valid",
			params:  valid,
			wantErr: false,
		},
		{
			name: "valid with optional fields",
			params: CreateCheckoutSessionParams{
				Name: "Order", Amount: 1000, Quantity: 2,
				SuccessURL: "https://example.com/success", CancelURL: "https://example.com/cancel",
				Description: "d", FinancialAccountID: "fac-1", PrimaryColor: "#fff", Images: []string{"https://img"},
			},
			wantErr: false,
		},
		{
			name:    "missing name",
			params:  CreateCheckoutSessionParams{Amount: 1000, Quantity: 1, SuccessURL: "https://e.com/s", CancelURL: "https://e.com/c"},
			wantErr: true,
		},
		{
			name:    "zero amount",
			params:  CreateCheckoutSessionParams{Name: "O", Amount: 0, Quantity: 1, SuccessURL: "https://e.com/s", CancelURL: "https://e.com/c"},
			wantErr: true,
		},
		{
			name:    "zero quantity",
			params:  CreateCheckoutSessionParams{Name: "O", Amount: 1000, Quantity: 0, SuccessURL: "https://e.com/s", CancelURL: "https://e.com/c"},
			wantErr: true,
		},
		{
			name:    "missing success url",
			params:  CreateCheckoutSessionParams{Name: "O", Amount: 1000, Quantity: 1, CancelURL: "https://e.com/c"},
			wantErr: true,
		},
		{
			name:    "invalid success url",
			params:  CreateCheckoutSessionParams{Name: "O", Amount: 1000, Quantity: 1, SuccessURL: "not-a-url", CancelURL: "https://e.com/c"},
			wantErr: true,
		},
		{
			name:    "invalid cancel url",
			params:  CreateCheckoutSessionParams{Name: "O", Amount: 1000, Quantity: 1, SuccessURL: "https://e.com/s", CancelURL: "nope"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStruct(tt.params)
			if tt.wantErr && err == nil {
				t.Fatal("expected validation error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}
			if tt.wantErr {
				var valErr *ValidationError
				if !errors.As(err, &valErr) {
					t.Fatalf("expected *ValidationError, got %T", err)
				}
			}
		})
	}
}
