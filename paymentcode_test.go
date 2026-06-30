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
		{
			name:    "valid",
			params:  CreatePaymentCodeParams{PaymentName: "Lunch", Amount: 100, Name: "John", PhoneNumber: "076000000"},
			wantErr: false,
		},
		{
			name:    "valid without financial account id",
			params:  CreatePaymentCodeParams{PaymentName: "Lunch", Amount: 100, Name: "John", PhoneNumber: "076"},
			wantErr: false,
		},
		{
			name:    "zero amount",
			params:  CreatePaymentCodeParams{PaymentName: "Lunch", Amount: 0, Name: "John", PhoneNumber: "076"},
			wantErr: true,
		},
		{
			name:    "missing payment name",
			params:  CreatePaymentCodeParams{Amount: 100, Name: "John", PhoneNumber: "076"},
			wantErr: true,
		},
		{
			name:    "missing name and phone",
			params:  CreatePaymentCodeParams{PaymentName: "Lunch", Amount: 100},
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
	_ = svc
}
