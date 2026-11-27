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
			// The API pays from the Space's main account when source is omitted.
			name: "no source account is allowed",
			params: CreatePayoutParams{
				Amount:      100,
				Destination: PayoutDestination{Type: "momo", ProviderID: "m17"},
			},
			wantErr: false,
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
		{name: "momo ok", dest: PayoutDestination{Type: "momo", ProviderID: "m17", PhoneNumber: "+23276123456"}, wantErr: false},
		{name: "bank ok", dest: PayoutDestination{Type: "bank", ProviderID: "slb001", AccountNumber: "0012345"}, wantErr: false},
		{name: "wallet ok without wallet id", dest: PayoutDestination{Type: "wallet", ProviderID: "dw001"}, wantErr: false},
		{name: "momo without phone", dest: PayoutDestination{Type: "momo", ProviderID: "m17"}, wantErr: true},
		{name: "bank without account", dest: PayoutDestination{Type: "bank", ProviderID: "slb001"}, wantErr: true},
		{name: "invalid type", dest: PayoutDestination{Type: "card", ProviderID: "x"}, wantErr: true},
		{name: "empty type", dest: PayoutDestination{ProviderID: "m17"}, wantErr: true},
		{name: "missing provider", dest: PayoutDestination{Type: "momo"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := payoutDestinationBody(tt.dest)
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

func TestPayoutDestinationBody_SendsOnlyTheVariantFields(t *testing.T) {
	body, err := payoutDestinationBody(PayoutDestination{
		Type: "momo", ProviderID: "m18", PhoneNumber: "+23276123456",
		AccountNumber: "ignored", WalletID: "ignored", TransactionReference: "read-only",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(body) != 3 || body["phoneNumber"] != "+23276123456" {
		t.Fatalf("momo destination must carry type, providerId and phoneNumber only, got %v", body)
	}
}
