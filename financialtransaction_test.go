package monime

import (
	"context"
	"errors"
	"testing"
)

func TestFinancialTransaction_Retrieve_RequiresID(t *testing.T) {
	svc := &FinancialTransactionService{}
	_, err := svc.Retrieve(context.Background(), "")
	if err == nil {
		t.Fatal("expected validation error for empty id")
	}
	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
}
