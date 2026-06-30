package monime

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInternalTransfer_Create(t *testing.T) {
	var gotMethod, gotPath, gotIdem string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotIdem = r.Method, r.URL.Path, r.Header.Get("Idempotency-Key")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"result":{"id":"it_1","status":"pending","amount":{"currency":"SLE","value":100},"sourceFinancialAccount":{"id":"acc_1"},"destinationFinancialAccount":{"id":"acc_2"}}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.InternalTransfer().Create(context.Background(), CreateInternalTransferParams{
		SourceAccount:      "acc_1",
		DestinationAccount: "acc_2",
		Amount:             100,
		Description:        "rent",
		Metadata:           map[string]any{"ref": "x"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "it_1" || got.Status != "pending" {
		t.Fatalf("unexpected response: %+v", got)
	}
	if got.SourceFinancialAccount.ID != "acc_1" || got.DestinationFinancialAccount.ID != "acc_2" {
		t.Fatalf("unexpected accounts: %+v", got)
	}
	if gotMethod != http.MethodPost || gotPath != "/internal-transfers" {
		t.Fatalf("method/path = %s %s", gotMethod, gotPath)
	}
	if gotIdem == "" {
		t.Error("expected Idempotency-Key header")
	}
	// amount is forwarded as-is (not scaled) with currency SLE.
	amount, _ := gotBody["amount"].(map[string]any)
	if amount["value"].(float64) != 100 {
		t.Errorf("amount.value = %v, want 100", amount["value"])
	}
	if amount["currency"] != "SLE" {
		t.Errorf("amount.currency = %v, want SLE", amount["currency"])
	}
	src, _ := gotBody["sourceFinancialAccount"].(map[string]any)
	if src["id"] != "acc_1" {
		t.Errorf("sourceFinancialAccount.id = %v", src["id"])
	}
	dst, _ := gotBody["destinationFinancialAccount"].(map[string]any)
	if dst["id"] != "acc_2" {
		t.Errorf("destinationFinancialAccount.id = %v", dst["id"])
	}
	if gotBody["description"] != "rent" {
		t.Errorf("description = %v", gotBody["description"])
	}
	meta, _ := gotBody["metadata"].(map[string]any)
	if meta["ref"] != "x" {
		t.Errorf("metadata.ref = %v", meta["ref"])
	}
}

func TestInternalTransfer_Create_DefaultsEmptyDescriptionAndMetadata(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"result":{"id":"it_2"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if _, err := c.InternalTransfer().Create(context.Background(), CreateInternalTransferParams{
		SourceAccount: "acc_1", DestinationAccount: "acc_2", Amount: 5,
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["description"] != "" {
		t.Errorf("description = %v, want empty string", gotBody["description"])
	}
	meta, ok := gotBody["metadata"].(map[string]any)
	if !ok || len(meta) != 0 {
		t.Errorf("metadata = %v, want empty object", gotBody["metadata"])
	}
}

func TestInternalTransfer_Create_ValidationShortCircuits(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("server should not be called when validation fails")
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.InternalTransfer().Create(context.Background(), CreateInternalTransferParams{SourceAccount: ""})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestInternalTransfer_Retrieve(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/internal-transfers/it_9" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"result":{"id":"it_9","status":"completed"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.InternalTransfer().Retrieve(context.Background(), "it_9")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "it_9" || got.Status != "completed" {
		t.Fatalf("unexpected: %+v", got)
	}
}

func TestInternalTransfer_Retrieve_RequiresID(t *testing.T) {
	c := newTestClient(t, httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})))
	if _, err := c.InternalTransfer().Retrieve(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty id")
	}
}

func TestInternalTransfer_List(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/internal-transfers" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"result":[{"id":"it_1"},{"id":"it_2"}],"pagination":{"count":2,"next":"cur"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.InternalTransfer().List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Result) != 2 || got.Pagination.Count != 2 || got.Pagination.Next != "cur" {
		t.Fatalf("unexpected list: %+v", got)
	}
}

func TestInternalTransfer_Update(t *testing.T) {
	var gotMethod, gotPath, gotIdem string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotIdem = r.Method, r.URL.Path, r.Header.Get("Idempotency-Key")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"result":{"id":"it_1","description":"updated"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.InternalTransfer().Update(context.Background(), "it_1", map[string]any{"description": "updated"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPatch || gotPath != "/internal-transfers/it_1" {
		t.Errorf("method/path = %s %s", gotMethod, gotPath)
	}
	if gotIdem == "" {
		t.Error("expected Idempotency-Key header")
	}
	if gotBody["description"] != "updated" {
		t.Errorf("body description = %v", gotBody["description"])
	}
	if got.Description != "updated" {
		t.Errorf("got.Description = %v", got.Description)
	}
}

func TestInternalTransfer_Update_Validation(t *testing.T) {
	c := newTestClient(t, httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})))
	if _, err := c.InternalTransfer().Update(context.Background(), "", map[string]any{"a": 1}); err == nil {
		t.Fatal("expected error for empty id")
	}
	if _, err := c.InternalTransfer().Update(context.Background(), "it_1", nil); err == nil {
		t.Fatal("expected error for empty params")
	}
}

func TestInternalTransfer_Delete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/internal-transfers/it_1" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if err := c.InternalTransfer().Delete(context.Background(), "it_1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInternalTransfer_Delete_RequiresID(t *testing.T) {
	c := newTestClient(t, httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})))
	if err := c.InternalTransfer().Delete(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty id")
	}
}
