package monime

import (
	"context"
	"errors"
	"testing"
)

func TestPayment_Retrieve_Validation(t *testing.T) {
	svc := &PaymentService{}
	_, err := svc.Retrieve(context.Background(), "")
	if err == nil {
		t.Fatal("expected validation error for empty id")
	}
	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
}

func TestPayment_Update_Validation(t *testing.T) {
	svc := &PaymentService{}
	tests := []struct {
		name    string
		id      string
		params  map[string]any
		wantErr bool
	}{
		{name: "empty id", id: "", params: map[string]any{"name": "x"}, wantErr: true},
		{name: "nil params", id: "pay_1", params: nil, wantErr: true},
		{name: "empty params", id: "pay_1", params: map[string]any{}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Update(context.Background(), tt.id, tt.params)
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
