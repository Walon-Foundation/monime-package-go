package monime

import (
	"errors"
	"testing"
)

func TestCreatePaymentCodeParams_Validation(t *testing.T) {
	svc := &PaymentCodeService{}
	tests := []struct {
		name    string
		params  CreatePaymentCodeParams
		wantErr bool
	}{
		{name: "name only (the API defaults the rest)", params: CreatePaymentCodeParams{Name: "Lunch"}, wantErr: false},
		{name: "full", params: CreatePaymentCodeParams{
			Name: "Lunch", Mode: PaymentCodeModeRecurrent, Amount: &Amount{Value: 2500}, CustomerName: "John",
			Reference: "ord-1", AuthorizedProviders: []string{"m17", "m18"}, AuthorizedPhoneNumber: "+23276000000",
			FinancialAccountID: "fac-1",
		}, wantErr: false},
		{name: "name too short", params: CreatePaymentCodeParams{Name: "ab"}, wantErr: true},
		{name: "missing name", params: CreatePaymentCodeParams{Amount: &Amount{Value: 100}}, wantErr: true},
		{name: "unknown mode", params: CreatePaymentCodeParams{Name: "Lunch", Mode: "weekly"}, wantErr: true},
		{name: "unknown provider", params: CreatePaymentCodeParams{Name: "Lunch", AuthorizedProviders: []string{"m99"}}, wantErr: true},
		{name: "account id without fac- prefix", params: CreatePaymentCodeParams{Name: "Lunch", FinancialAccountID: "acc-1"}, wantErr: true},
		{name: "reference over 64", params: CreatePaymentCodeParams{Name: "Lunch", Reference: string(make([]byte, 65))}, wantErr: true},
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
	_ = svc
}
