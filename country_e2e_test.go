package monime

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestCountry_Retrieve(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_, _ = w.Write([]byte(`{"result":{"code":"SL","name":"Sierra Leone",` +
			`"currency":{"code":"SLE","unit":"cent","unitLength":2},` +
			`"supportedCurrencies":["SLE","USD"]}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.Country().Retrieve(context.Background(), "sl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodGet || gotPath != "/countries/SL" {
		t.Fatalf("method/path = %s %s, want GET /countries/SL (uppercased)", gotMethod, gotPath)
	}
	if got.Code != "SL" || got.Name != "Sierra Leone" {
		t.Fatalf("unexpected country: %+v", got)
	}
	if got.Currency.Code != "SLE" || got.Currency.Unit != "cent" || got.Currency.UnitLength != 2 {
		t.Fatalf("currency = %+v", got.Currency)
	}
	if len(got.SupportedCurrencies) != 2 || got.SupportedCurrencies[1] != "USD" {
		t.Fatalf("supportedCurrencies = %v", got.SupportedCurrencies)
	}
}

func TestCountry_List(t *testing.T) {
	var gotMethod, gotPath string
	var gotQuery url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotQuery = r.Method, r.URL.Path, r.URL.Query()
		_, _ = w.Write([]byte(`{"success":true,"result":[{"code":"SL"},{"code":"LR"}],` +
			`"pagination":{"count":2,"next":"cur"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.Country().List(context.Background(), WithLimit(2), WithAfter("cur-0"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodGet || gotPath != "/countries" {
		t.Fatalf("method/path = %s %s", gotMethod, gotPath)
	}
	if gotQuery.Get("limit") != "2" || gotQuery.Get("after") != "cur-0" {
		t.Errorf("query = %v", gotQuery)
	}
	if len(got.Result) != 2 || got.Result[0].Code != "SL" {
		t.Fatalf("unexpected list: %+v", got)
	}
	if got.Pagination.Count != 2 || got.Pagination.Next != "cur" {
		t.Fatalf("pagination = %+v", got.Pagination)
	}
}

func TestCountry_List_InvalidOptionSkipsRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("request should not be sent when an option fails validation")
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if _, err := c.Country().List(context.Background(), WithLimit(0)); err == nil {
		t.Fatal("expected a validation error for limit=0")
	}
}
