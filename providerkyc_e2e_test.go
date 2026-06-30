package monime

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProviderKyc_Retrieve(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_, _ = w.Write([]byte(`{"result":{"account":{"id":"acc_1","name":"Main","holderName":"John Doe","metadata":{"k":"v"}},"provider":{"id":"m17","type":"momo","name":"Orange Money"}}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.ProviderKyc().Retrieve(context.Background(), "m17")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodGet || gotPath != "/provider-kyc/m17" {
		t.Fatalf("method/path = %s %s", gotMethod, gotPath)
	}
	if got.Provider.ID != "m17" || got.Provider.Type != "momo" || got.Provider.Name != "Orange Money" {
		t.Fatalf("unexpected provider: %+v", got.Provider)
	}
	if got.Account.ID != "acc_1" || got.Account.HolderName != "John Doe" {
		t.Fatalf("unexpected account: %+v", got.Account)
	}
	if got.Account.Metadata["k"] != "v" {
		t.Fatalf("unexpected metadata: %+v", got.Account.Metadata)
	}
}

func TestProviderKyc_Retrieve_RequiresID_NoCall(t *testing.T) {
	c := newTestClient(t, httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("server should not be called for empty id")
	})))
	if _, err := c.ProviderKyc().Retrieve(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty id")
	}
}
