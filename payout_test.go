package monime

import (
	"errors"
	"testing"
)

func TestCreatePayoutParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  CreatePayoutParams
		wantErr bool
	}{
		{
			name: "valid momo",
			params: CreatePayoutParams{
				Amount:        1000,
				SourceAccount: "fac-1",
				Destination:   PayoutDestination{Type: "momo", ProviderID: "m17", PhoneNumber: "076000000"},
			},
			wantErr: false,
		},
		{
			name: "valid bank",
			params: CreatePayoutParams{
				Amount:        500,
				SourceAccount: "fac-1",
				Destination:   PayoutDestination{Type: "bank", ProviderID: "slb001", AccountNumber: "12345"},
			},
			wantErr: false,
		},
		{
			name: "valid wallet",
			params: CreatePayoutParams{
				Amount:        500,
				SourceAccount: "fac-1",
				Destination:   PayoutDestination{Type: "wallet", ProviderID: "dw001", WalletID: "w-1"},
			},
			wantErr: false,
		},
		{
			name: "zero amount",
			params: CreatePayoutParams{
				Amount:        0,
				SourceAccount: "fac-1",
				Destination:   PayoutDestination{Type: "momo", ProviderID: "m17", PhoneNumber: "076"},
			},
			wantErr: true,
		},
		{
			name: "negative amount",
			params: CreatePayoutParams{
				Amount:        -5,
				SourceAccount: "fac-1",
				Destination:   PayoutDestination{Type: "momo", ProviderID: "m17"},
			},
			wantErr: true,
		},
		{
			name: "missing source account",
			params: CreatePayoutParams{
				Amount:      100,
				Destination: PayoutDestination{Type: "momo", ProviderID: "m17"},
			},
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

func TestValidatePayoutDestination(t *testing.T) {
	tests := []struct {
		name    string
		dest    PayoutDestination
		wantErr bool
	}{
		{name: "momo ok", dest: PayoutDestination{Type: "momo", ProviderID: "m17"}, wantErr: false},
		{name: "bank ok", dest: PayoutDestination{Type: "bank", ProviderID: "slb001"}, wantErr: false},
		{name: "wallet ok", dest: PayoutDestination{Type: "wallet", ProviderID: "dw001"}, wantErr: false},
		{name: "invalid type", dest: PayoutDestination{Type: "card", ProviderID: "x"}, wantErr: true},
		{name: "empty type", dest: PayoutDestination{ProviderID: "m17"}, wantErr: true},
		{name: "missing provider", dest: PayoutDestination{Type: "momo"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePayoutDestination(tt.dest)
			if tt.wantErr && err == nil {
				t.Fatal("expected validation error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
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
