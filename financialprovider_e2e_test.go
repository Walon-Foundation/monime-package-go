package monime

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBank_Retrieve(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_, _ = w.Write([]byte(`{"result":{"providerId":"bank_1","name":"Acme Bank","country":"SL","status":{"active":true},"featureSet":{"payout":{"canPayTo":true,"schemes":["a"],"metadata":{}},"payment":{"canPayFrom":true,"metadata":{}},"kycVerification":{"canVerifyAccount":true,"metadata":{}}}}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.FinancialProvider().Bank().Retrieve(context.Background(), "bank_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodGet || gotPath != "/banks/bank_1" {
		t.Fatalf("method/path = %s %s", gotMethod, gotPath)
	}
	if got.ProviderID != "bank_1" || got.Name != "Acme Bank" || !got.Status.Active {
		t.Fatalf("unexpected bank: %+v", got)
	}
	if !got.FeatureSet.Payout.CanPayTo || len(got.FeatureSet.Payout.Schemes) != 1 {
		t.Fatalf("unexpected featureSet: %+v", got.FeatureSet)
	}
	if !got.FeatureSet.KYCVerification.CanVerifyAccount {
		t.Errorf("expected kycVerification.canVerifyAccount true")
	}
}

func TestBank_Retrieve_RequiresID(t *testing.T) {
	c := newTestClient(t, httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})))
	if _, err := c.FinancialProvider().Bank().Retrieve(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty providerID")
	}
}

func TestBank_List(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_, _ = w.Write([]byte(`{"result":[{"providerId":"bank_1"},{"providerId":"bank_2"}],"pagination":{"count":2,"next":"cur"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.FinancialProvider().Bank().List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodGet || gotPath != "/banks" {
		t.Fatalf("method/path = %s %s", gotMethod, gotPath)
	}
	if len(got.Result) != 2 || got.Pagination.Count != 2 || got.Pagination.Next != "cur" {
		t.Fatalf("unexpected list: %+v", got)
	}
}

func TestMomo_Retrieve(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_, _ = w.Write([]byte(`{"result":{"providerId":"momo_1","name":"Orange Money","country":"SL","status":{"active":true},"featureSet":{"payout":{"canPayTo":true,"metadata":{}},"payment":{"canPayFrom":true,"metadata":{}},"kycVerification":{"canVerifyAccount":false,"metadata":{}}}}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.FinancialProvider().Momo().Retrieve(context.Background(), "momo_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodGet || gotPath != "/momos/momo_1" {
		t.Fatalf("method/path = %s %s", gotMethod, gotPath)
	}
	if got.ProviderID != "momo_1" || got.Name != "Orange Money" || !got.Status.Active {
		t.Fatalf("unexpected momo: %+v", got)
	}
	if !got.FeatureSet.Payment.CanPayFrom {
		t.Errorf("expected payment.canPayFrom true")
	}
}

func TestMomo_Retrieve_RequiresID(t *testing.T) {
	c := newTestClient(t, httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})))
	if _, err := c.FinancialProvider().Momo().Retrieve(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty providerID")
	}
}

func TestMomo_List(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_, _ = w.Write([]byte(`{"result":[{"providerId":"momo_1"},{"providerId":"momo_2"},{"providerId":"momo_3"}],"pagination":{"count":3,"next":""}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.FinancialProvider().Momo().List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodGet || gotPath != "/momos" {
		t.Fatalf("method/path = %s %s", gotMethod, gotPath)
	}
	if len(got.Result) != 3 || got.Pagination.Count != 3 {
		t.Fatalf("unexpected list: %+v", got)
	}
}
