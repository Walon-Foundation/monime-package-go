package monime

import (
	"context"
	"errors"
	"testing"
)

func TestCountry_Retrieve_Validation(t *testing.T) {
	c, err := New(WithSpaceID("s"), WithAccessToken("t"))
	if err != nil {
		t.Fatalf("client build: %v", err)
	}

	tests := []struct {
		name        string
		countryCode string
	}{
		{"empty", ""},
		{"alpha-3", "SLE"},
		{"single character", "S"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := c.Country().Retrieve(context.Background(), tt.countryCode)
			if err == nil {
				t.Fatal("expected a validation error")
			}
			var valErr *ValidationError
			if !errors.As(err, &valErr) {
				t.Fatalf("expected *ValidationError, got %T", err)
			}
		})
	}
}
