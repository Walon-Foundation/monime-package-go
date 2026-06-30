package monime

import (
	"context"
	"errors"
	"testing"
)

func TestBankService_Retrieve_RequiresProviderID(t *testing.T) {
	svc := &BankService{}
	_, err := svc.Retrieve(context.Background(), "")
	if err == nil {
		t.Fatal("expected validation error for empty providerID")
	}
	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
}

func TestMomoService_Retrieve_RequiresProviderID(t *testing.T) {
	svc := &MomoService{}
	_, err := svc.Retrieve(context.Background(), "")
	if err == nil {
		t.Fatal("expected validation error for empty providerID")
	}
	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
}
