package monime

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFinancialAccount_Create(t *testing.T) {
	var gotMethod, gotPath, gotIdem string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotIdem = r.Method, r.URL.Path, r.Header.Get("Idempotency-Key")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"result":{"id":"fac_1","uvan":"UVAN1","name":"Main Wallet","currency":"SLE","balance":{"available":{"currency":"SLE","value":5000}}}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.FinancialAccount().Create(context.Background(), CreateFinancialAccountParams{
		AccountName: "Main Wallet",
		Currency:    CurrencySLE,
		Description: "primary",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "fac_1" || got.UVAN != "UVAN1" || got.Balance.Available.Value != 5000 {
		t.Fatalf("unexpected response: %+v", got)
	}
	if gotMethod != http.MethodPost || gotPath != "/financial-accounts" {
		t.Fatalf("method/path = %s %s", gotMethod, gotPath)
	}
	if gotIdem == "" {
		t.Error("expected Idempotency-Key header")
	}
	if gotBody["name"] != "Main Wallet" {
		t.Errorf("name = %v", gotBody["name"])
	}
	if gotBody["currency"] != "SLE" {
		t.Errorf("currency = %v", gotBody["currency"])
	}
	if gotBody["description"] != "primary" {
		t.Errorf("description = %v", gotBody["description"])
	}
	if _, ok := gotBody["metadata"].(map[string]any); !ok {
		t.Errorf("metadata should default to an object, got %v", gotBody["metadata"])
	}
}

func TestFinancialAccount_Create_ValidationShortCircuits(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("server should not be called when validation fails")
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if _, err := c.FinancialAccount().Create(context.Background(), CreateFinancialAccountParams{
		AccountName: "", Currency: CurrencyUSD,
	}); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestFinancialAccount_Retrieve(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/financial-accounts/fac_9" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"result":{"id":"fac_9","name":"Savings"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.FinancialAccount().Retrieve(context.Background(), "fac_9")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "fac_9" || got.Name != "Savings" {
		t.Fatalf("unexpected: %+v", got)
	}
}

func TestFinancialAccount_Retrieve_RequiresID(t *testing.T) {
	c := newTestClient(t, httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})))
	if _, err := c.FinancialAccount().Retrieve(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty id")
	}
}

func TestFinancialAccount_List(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"result":[{"id":"fac_1"},{"id":"fac_2"}],"pagination":{"count":2,"next":"tok"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.FinancialAccount().List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Result) != 2 || got.Pagination.Count != 2 || got.Pagination.Next != "tok" {
		t.Fatalf("unexpected list: %+v", got)
	}
}

func TestFinancialAccount_Update(t *testing.T) {
	var gotMethod, gotPath, gotIdem string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotIdem = r.Method, r.URL.Path, r.Header.Get("Idempotency-Key")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"result":{"id":"fac_1","name":"Renamed"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.FinancialAccount().Update(context.Background(), "fac_1", map[string]any{"name": "Renamed"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPatch {
		t.Errorf("method = %s, want PATCH", gotMethod)
	}
	if gotPath != "/financial-accounts/fac_1" {
		t.Errorf("path = %s", gotPath)
	}
	if gotIdem == "" {
		t.Error("expected Idempotency-Key header")
	}
	if got.Name != "Renamed" {
		t.Errorf("name = %s", got.Name)
	}
	if gotBody["name"] != "Renamed" {
		t.Errorf("body name = %v", gotBody["name"])
	}
}

func TestFinancialAccount_Update_Validation(t *testing.T) {
	c := newTestClient(t, httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})))
	if _, err := c.FinancialAccount().Update(context.Background(), "", map[string]any{"name": "x"}); err == nil {
		t.Fatal("expected error for empty id")
	}
	if _, err := c.FinancialAccount().Update(context.Background(), "fac_1", nil); err == nil {
		t.Fatal("expected error for empty params")
	}
}
