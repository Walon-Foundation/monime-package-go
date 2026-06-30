package monime

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPayment_Retrieve(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_, _ = w.Write([]byte(`{"result":{"id":"pay_9","status":"completed","amount":{"currency":"SLE","value":10000},"channel":{"type":"momo"},"fees":[{"code":"processing","amount":{"currency":"SLE","value":50}}],"ownershipGraph":{"owner":{"id":"own_1","type":"checkout_session"}}}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.Payment().Retrieve(context.Background(), "pay_9")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodGet || gotPath != "/payments/pay_9" {
		t.Fatalf("method/path = %s %s", gotMethod, gotPath)
	}
	if got.ID != "pay_9" || got.Status != "completed" {
		t.Fatalf("unexpected: %+v", got)
	}
	if got.Channel.Type != "momo" {
		t.Errorf("channel.type = %s, want momo", got.Channel.Type)
	}
	if got.Amount.Value != 10000 {
		t.Errorf("amount.value = %d, want 10000", got.Amount.Value)
	}
	if len(got.Fees) != 1 || got.Fees[0].Code != "processing" || got.Fees[0].Amount.Value != 50 {
		t.Errorf("unexpected fees: %+v", got.Fees)
	}
	if got.OwnershipGraph.Owner.ID != "own_1" || got.OwnershipGraph.Owner.Type != "checkout_session" {
		t.Errorf("unexpected owner: %+v", got.OwnershipGraph.Owner)
	}
}

func TestPayment_Retrieve_RequiresID(t *testing.T) {
	c := newTestClient(t, httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("server should not be called when validation fails")
	})))
	if _, err := c.Payment().Retrieve(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty id")
	}
}

func TestPayment_List(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_, _ = w.Write([]byte(`{"result":[{"id":"pay_1"},{"id":"pay_2"}],"pagination":{"count":2,"next":"abc"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.Payment().List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodGet || gotPath != "/payments" {
		t.Fatalf("method/path = %s %s", gotMethod, gotPath)
	}
	if len(got.Result) != 2 || got.Pagination.Count != 2 || got.Pagination.Next != "abc" {
		t.Fatalf("unexpected list: %+v", got)
	}
}

func TestPayment_Update(t *testing.T) {
	var gotMethod, gotPath, gotIdem string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotIdem = r.Method, r.URL.Path, r.Header.Get("Idempotency-Key")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"result":{"id":"pay_1","name":"Updated"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.Payment().Update(context.Background(), "pay_1", map[string]any{"name": "Updated"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPatch {
		t.Errorf("method = %s, want PATCH", gotMethod)
	}
	if gotPath != "/payments/pay_1" {
		t.Errorf("path = %s", gotPath)
	}
	if gotIdem == "" {
		t.Error("expected Idempotency-Key header")
	}
	if gotBody["name"] != "Updated" {
		t.Errorf("body name = %v", gotBody["name"])
	}
	if got.Name != "Updated" {
		t.Errorf("response name = %s", got.Name)
	}
}

func TestPayment_Update_ValidationShortCircuits(t *testing.T) {
	c := newTestClient(t, httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("server should not be called when validation fails")
	})))
	if _, err := c.Payment().Update(context.Background(), "", map[string]any{"a": 1}); err == nil {
		t.Fatal("expected error for empty id")
	}
	if _, err := c.Payment().Update(context.Background(), "pay_1", nil); err == nil {
		t.Fatal("expected error for empty params")
	}
}
