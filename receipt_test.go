package monime

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReceipt_Retrieve_RequiresOrderNumber(t *testing.T) {
	c := newTestClient(t, httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("server should not be called when validation fails")
	})))
	_, err := c.Receipt().Retrieve(context.Background(), "")
	if err == nil {
		t.Fatal("expected validation error for empty order number")
	}
	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
}

func TestReceipt_Redeem_Validation(t *testing.T) {
	c := newTestClient(t, httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("server should not be called when validation fails")
	})))

	if _, err := c.Receipt().Redeem(context.Background(), "", map[string]any{"redeemAll": true}); err == nil {
		t.Fatal("expected validation error for empty order number")
	}
	_, err := c.Receipt().Redeem(context.Background(), "ord_1", nil)
	if err == nil {
		t.Fatal("expected validation error for empty body")
	}
	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
}
