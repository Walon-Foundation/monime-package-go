package monime

import (
	"errors"
	"testing"
)

func TestCreateInternalTransferParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  CreateInternalTransferParams
		wantErr bool
	}{
		{
			name:    "valid",
			params:  CreateInternalTransferParams{SourceAccount: "acc_1", DestinationAccount: "acc_2", Amount: 100},
			wantErr: false,
		},
		{
			name:    "valid with description and metadata",
			params:  CreateInternalTransferParams{SourceAccount: "acc_1", DestinationAccount: "acc_2", Amount: 50, Description: "rent", Metadata: map[string]any{"k": "v"}},
			wantErr: false,
		},
		{
			name:    "zero amount",
			params:  CreateInternalTransferParams{SourceAccount: "acc_1", DestinationAccount: "acc_2", Amount: 0},
			wantErr: true,
		},
		{
			name:    "negative amount",
			params:  CreateInternalTransferParams{SourceAccount: "acc_1", DestinationAccount: "acc_2", Amount: -10},
			wantErr: true,
		},
		{
			name:    "missing source account",
			params:  CreateInternalTransferParams{DestinationAccount: "acc_2", Amount: 100},
			wantErr: true,
		},
		{
			name:    "missing destination account",
			params:  CreateInternalTransferParams{SourceAccount: "acc_1", Amount: 100},
			wantErr: true,
		},
		{
			name:    "missing both accounts",
			params:  CreateInternalTransferParams{Amount: 100},
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
