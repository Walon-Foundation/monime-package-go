package monime

import (
	"errors"
	"testing"
)

func TestCreateUssdOtpParams_Validation(t *testing.T) {
	svc := &UssdOtpService{}
	tests := []struct {
		name    string
		params  CreateUssdOtpParams
		wantErr bool
	}{
		{
			name:    "valid",
			params:  CreateUssdOtpParams{PhoneNumber: "076000000"},
			wantErr: false,
		},
		{
			name:    "missing phone number",
			params:  CreateUssdOtpParams{},
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
