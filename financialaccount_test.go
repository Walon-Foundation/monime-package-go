package monime

import (
	"errors"
	"testing"
)

func TestCreateFinancialAccountParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  CreateFinancialAccountParams
		wantErr bool
	}{
		{
			name:    "valid USD",
			params:  CreateFinancialAccountParams{AccountName: "Main Wallet", Currency: CurrencyUSD},
			wantErr: false,
		},
		{
			name:    "valid SLE",
			params:  CreateFinancialAccountParams{AccountName: "Main Wallet", Currency: CurrencySLE},
			wantErr: false,
		},
		{
			name:    "missing account name",
			params:  CreateFinancialAccountParams{Currency: CurrencyUSD},
			wantErr: true,
		},
		{
			name:    "missing currency",
			params:  CreateFinancialAccountParams{AccountName: "Main Wallet"},
			wantErr: true,
		},
		{
			name:    "bad currency",
			params:  CreateFinancialAccountParams{AccountName: "Main Wallet", Currency: Currency("GBP")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// validateStruct is what Create calls before any network round trip.
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
