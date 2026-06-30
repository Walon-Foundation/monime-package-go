package monime

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPayout_Create(t *testing.T) {
	var gotMethod, gotPath, gotIdem string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotIdem = r.Method, r.URL.Path, r.Header.Get("Idempotency-Key")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"result":{"id":"po_1","status":"pending","amount":{"currency":"SLE","value":1000}}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.Payout().Create(context.Background(), CreatePayoutParams{
		Amount:        1000,
		SourceAccount: "fac-1",
		Destination:   PayoutDestination{Type: "momo", ProviderID: "m17", PhoneNumber: "076000000"},
		Metadata:      map[string]any{"orderId": "ord-1"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "po_1" || got.Status != "pending" {
		t.Fatalf("unexpected response: %+v", got)
	}
	if gotMethod != http.MethodPost || gotPath != "/payouts" {
		t.Fatalf("method/path = %s %s", gotMethod, gotPath)
	}
	if gotIdem == "" {
		t.Error("expected Idempotency-Key header")
	}

	// Amount must be forwarded as-is (NOT scaled to minor units).
	amount, _ := gotBody["amount"].(map[string]any)
	if amount["value"].(float64) != 1000 {
		t.Errorf("amount.value = %v, want 1000 (no scaling)", amount["value"])
	}
	if amount["currency"] != "SLE" {
		t.Errorf("amount.currency = %v", amount["currency"])
	}

	source, _ := gotBody["source"].(map[string]any)
	if source["financialAccountId"] != "fac-1" {
		t.Errorf("source.financialAccountId = %v", source["financialAccountId"])
	}

	dest, _ := gotBody["destination"].(map[string]any)
	if dest["type"] != "momo" || dest["providerId"] != "m17" || dest["phoneNumber"] != "076000000" {
		t.Errorf("destination = %v", dest)
	}
	// Union-specific fields for other types must be omitted.
	if _, ok := dest["accountNumber"]; ok {
		t.Error("accountNumber should be omitted for momo")
	}
	if _, ok := dest["walletId"]; ok {
		t.Error("walletId should be omitted for momo")
	}

	meta, _ := gotBody["metadata"].(map[string]any)
	if meta["orderId"] != "ord-1" {
		t.Errorf("metadata = %v", meta)
	}
}

func TestPayout_Create_DefaultsMetadata(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"result":{"id":"po_2"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if _, err := c.Payout().Create(context.Background(), CreatePayoutParams{
		Amount:        50,
		SourceAccount: "fac-1",
		Destination:   PayoutDestination{Type: "bank", ProviderID: "slb001", AccountNumber: "12345"},
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	meta, ok := gotBody["metadata"].(map[string]any)
	if !ok || len(meta) != 0 {
		t.Errorf("metadata should default to empty object, got %v", gotBody["metadata"])
	}
}

func TestPayout_Create_ValidationShortCircuits(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("server should not be called when validation fails")
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	// Missing source account.
	if _, err := c.Payout().Create(context.Background(), CreatePayoutParams{
		Amount:      100,
		Destination: PayoutDestination{Type: "momo", ProviderID: "m17"},
	}); err == nil {
		t.Fatal("expected validation error for missing source account")
	}
	// Invalid destination type passes struct validation but fails union check.
	if _, err := c.Payout().Create(context.Background(), CreatePayoutParams{
		Amount:        100,
		SourceAccount: "fac-1",
		Destination:   PayoutDestination{Type: "card", ProviderID: "x"},
	}); err == nil {
		t.Fatal("expected validation error for invalid destination type")
	}
}

func TestPayout_Retrieve(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/payouts/po_9" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"result":{"id":"po_9","status":"completed","fees":[{"code":"f1","amount":{"currency":"SLE","value":10}}]}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.Payout().Retrieve(context.Background(), "po_9")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "po_9" || got.Status != "completed" {
		t.Fatalf("unexpected: %+v", got)
	}
	if len(got.Fees) != 1 || got.Fees[0].Code != "f1" || got.Fees[0].Amount.Value != 10 {
		t.Fatalf("fees = %+v", got.Fees)
	}
}

func TestPayout_Retrieve_RequiresID(t *testing.T) {
	c := newTestClient(t, httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})))
	if _, err := c.Payout().Retrieve(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty id")
	}
}

func TestPayout_List(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/payouts" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"result":[{"id":"po_1"},{"id":"po_2"}],"pagination":{"count":2,"next":"cur"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.Payout().List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Result) != 2 || got.Pagination.Count != 2 || got.Pagination.Next != "cur" {
		t.Fatalf("unexpected list: %+v", got)
	}
}

func TestPayout_Update(t *testing.T) {
	var gotMethod, gotPath, gotIdem string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotIdem = r.Method, r.URL.Path, r.Header.Get("Idempotency-Key")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"result":{"id":"po_1","status":"pending"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.Payout().Update(context.Background(), "po_1", map[string]any{"metadata": map[string]any{"k": "v"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPatch || gotPath != "/payouts/po_1" {
		t.Errorf("method/path = %s %s, want PATCH /payouts/po_1", gotMethod, gotPath)
	}
	if gotIdem == "" {
		t.Error("expected Idempotency-Key header on update")
	}
	meta, _ := gotBody["metadata"].(map[string]any)
	if meta["k"] != "v" {
		t.Errorf("body metadata = %v", gotBody["metadata"])
	}
	if got.ID != "po_1" {
		t.Errorf("unexpected response: %+v", got)
	}
}

func TestPayout_Update_Validation(t *testing.T) {
	c := newTestClient(t, httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("server should not be called when validation fails")
	})))
	if _, err := c.Payout().Update(context.Background(), "", map[string]any{"a": 1}); err == nil {
		t.Fatal("expected error for empty id")
	}
	if _, err := c.Payout().Update(context.Background(), "po_1", nil); err == nil {
		t.Fatal("expected error for empty params")
	}
}

func TestPayout_Delete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/payouts/po_1" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if err := c.Payout().Delete(context.Background(), "po_1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPayout_Delete_RequiresID(t *testing.T) {
	c := newTestClient(t, httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})))
	if err := c.Payout().Delete(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty id")
	}
}
