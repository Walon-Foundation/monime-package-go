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

	if _, err := c.Receipt().Redeem(context.Background(), "", RedeemReceiptParams{RedeemAll: true}); err == nil {
		t.Fatal("expected validation error for empty order number")
	}
	if _, err := c.Receipt().Redeem(context.Background(), "123456789012345678901", RedeemReceiptParams{RedeemAll: true}); err == nil {
		t.Fatal("expected validation error for an order number over 20 characters")
	}
	if _, err := c.Receipt().Redeem(context.Background(), "ord_1", RedeemReceiptParams{Entitlements: []ReceiptRedemptionItem{{Key: ""}}}); err == nil {
		t.Fatal("expected validation error for an entitlement without a key")
	}
	_, err := c.Receipt().Redeem(context.Background(), "ord_1", RedeemReceiptParams{})
	if err == nil {
		t.Fatal("expected validation error when nothing is redeemed")
	}
	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
}
