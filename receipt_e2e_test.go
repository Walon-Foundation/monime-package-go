package monime

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReceipt_Retrieve(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_, _ = w.Write([]byte(`{"result":{"orderNumber":"ord_9","status":"not_redeemed","orderAmount":{"currency":"SLE","value":10000},"entitlements":[{"key":"ticket","limit":2,"current":0,"remaining":2,"exhausted":false}]}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.Receipt().Retrieve(context.Background(), "ord_9")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodGet || gotPath != "/receipts/ord_9" {
		t.Fatalf("method/path = %s %s", gotMethod, gotPath)
	}
	if got.OrderNumber != "ord_9" || got.Status != "not_redeemed" {
		t.Fatalf("unexpected response: %+v", got)
	}
	if got.OrderAmount.Value != 10000 {
		t.Errorf("orderAmount.value = %d, want 10000", got.OrderAmount.Value)
	}
	if len(got.Entitlements) != 1 || got.Entitlements[0].Key != "ticket" || got.Entitlements[0].Remaining != 2 {
		t.Errorf("unexpected entitlements: %+v", got.Entitlements)
	}
}

func TestReceipt_Redeem(t *testing.T) {
	var gotMethod, gotPath, gotIdem string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotIdem = r.Method, r.URL.Path, r.Header.Get("Idempotency-Key")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"result":{"orderNumber":"ord_9","status":"fully_redeemed"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.Receipt().Redeem(context.Background(), "ord_9", map[string]any{
		"entitlements": []map[string]any{{"key": "ticket", "units": 1}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPost || gotPath != "/receipts/ord_9/redeem" {
		t.Fatalf("method/path = %s %s", gotMethod, gotPath)
	}
	if gotIdem == "" {
		t.Error("expected Idempotency-Key header")
	}
	if got.Status != "fully_redeemed" {
		t.Fatalf("unexpected response: %+v", got)
	}
	ents, _ := gotBody["entitlements"].([]any)
	if len(ents) != 1 {
		t.Fatalf("expected one entitlement in body, got %+v", gotBody)
	}
	first, _ := ents[0].(map[string]any)
	if first["key"] != "ticket" {
		t.Errorf("entitlement key = %v, want ticket", first["key"])
	}
}
