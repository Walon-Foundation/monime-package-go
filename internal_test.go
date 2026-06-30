package monime

import (
	"errors"
	"testing"
)

func TestGenerateIdempotencyKey(t *testing.T) {
	a, err := generateIdempotencyKey()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a) != 40 { // 20 random bytes hex-encoded
		t.Fatalf("key length = %d, want 40", len(a))
	}
	b, _ := generateIdempotencyKey()
	if a == b {
		t.Fatal("expected unique keys")
	}
}

type validatable struct {
	Name   string `validate:"required"`
	Amount int    `validate:"gt=0"`
}

func TestValidateStruct(t *testing.T) {
	t.Run("valid input passes", func(t *testing.T) {
		if err := validateStruct(validatable{Name: "ok", Amount: 5}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("invalid input returns *ValidationError", func(t *testing.T) {
		err := validateStruct(validatable{Name: "", Amount: 0})
		if err == nil {
			t.Fatal("expected validation error")
		}
		var valErr *ValidationError
		if !errors.As(err, &valErr) {
			t.Fatalf("expected *ValidationError, got %T", err)
		}
		var apiErr *Error
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected unwrap to *Error, got %T", err)
		}
		if apiErr.Status != 400 {
			t.Fatalf("status = %d, want 400", apiErr.Status)
		}
	})
}
