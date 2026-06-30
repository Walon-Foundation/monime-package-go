package monime

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFinancialTransaction_Retrieve(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_, _ = w.Write([]byte(`{"result":{"id":"ft_9","type":"credit","amount":{"currency":"SLE","value":5000},"financialAccount":{"id":"fac_1","balance":{"after":{"currency":"SLE","value":12000}}},"ownershipGraph":{"owner":{"id":"own_1","type":"space"}}}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.FinancialTransaction().Retrieve(context.Background(), "ft_9")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodGet || gotPath != "/financial-transactions/ft_9" {
		t.Fatalf("method/path = %s %s", gotMethod, gotPath)
	}
	if got.ID != "ft_9" || got.Type != "credit" {
		t.Fatalf("unexpected response: %+v", got)
	}
	if got.Amount.Value != 5000 {
		t.Errorf("amount.value = %d, want 5000", got.Amount.Value)
	}
	if got.FinancialAccount.Balance.After.Value != 12000 {
		t.Errorf("balance.after.value = %d, want 12000", got.FinancialAccount.Balance.After.Value)
	}
	if got.OwnershipGraph.Owner.ID != "own_1" {
		t.Errorf("ownershipGraph.owner.id = %s, want own_1", got.OwnershipGraph.Owner.ID)
	}
}

func TestFinancialTransaction_Retrieve_RequiresID_E2E(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("server should not be called when id is empty")
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if _, err := c.FinancialTransaction().Retrieve(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty id")
	}
}

func TestFinancialTransaction_List(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"result":[{"id":"ft_1"},{"id":"ft_2"}],"pagination":{"count":2,"next":"ft_3"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.FinancialTransaction().List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/financial-transactions" {
		t.Fatalf("path = %s", gotPath)
	}
	if len(got.Result) != 2 || got.Result[0].ID != "ft_1" {
		t.Fatalf("unexpected result: %+v", got.Result)
	}
	if got.Pagination.Count != 2 || got.Pagination.Next != "ft_3" {
		t.Fatalf("unexpected pagination: %+v", got.Pagination)
	}
}
