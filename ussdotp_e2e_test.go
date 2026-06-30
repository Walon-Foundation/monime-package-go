package monime

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUssdOtp_Create(t *testing.T) {
	var gotMethod, gotPath, gotIdem string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotIdem = r.Method, r.URL.Path, r.Header.Get("Idempotency-Key")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"result":{"id":"otp_1","status":"pending","dialCode":"*123#","authorizedPhoneNumber":"076000000"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.UssdOtp().Create(context.Background(), CreateUssdOtpParams{
		AuthorizedPhoneNumber: "076000000",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "otp_1" || got.Status != "pending" || got.DialCode != "*123#" {
		t.Fatalf("unexpected response: %+v", got)
	}
	if gotMethod != http.MethodPost || gotPath != "/ussd-otps" {
		t.Fatalf("method/path = %s %s", gotMethod, gotPath)
	}
	if gotIdem == "" {
		t.Error("expected Idempotency-Key header")
	}
	// body must send authorizedPhoneNumber (per the API docs).
	if gotBody["authorizedPhoneNumber"] != "076000000" {
		t.Errorf("phoneNumber = %v, want 076000000", gotBody["authorizedPhoneNumber"])
	}
}

func TestUssdOtp_Create_ValidationShortCircuits(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("server should not be called when validation fails")
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.UssdOtp().Create(context.Background(), CreateUssdOtpParams{AuthorizedPhoneNumber: ""})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestUssdOtp_Retrieve(t *testing.T) {
	var gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		if r.URL.Path != "/ussd-otps/otp_9" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"result":{"id":"otp_9","status":"verified"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.UssdOtp().Retrieve(context.Background(), "otp_9")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("method = %s, want GET", gotMethod)
	}
	if got.ID != "otp_9" || got.Status != "verified" {
		t.Fatalf("unexpected: %+v", got)
	}
}

func TestUssdOtp_Retrieve_RequiresID(t *testing.T) {
	c := newTestClient(t, httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})))
	if _, err := c.UssdOtp().Retrieve(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty id")
	}
}

func TestUssdOtp_List(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_, _ = w.Write([]byte(`{"result":[{"id":"otp_1"},{"id":"otp_2"}],"pagination":{"count":2,"next":"cur_3"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.UssdOtp().List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodGet || gotPath != "/ussd-otps" {
		t.Fatalf("method/path = %s %s", gotMethod, gotPath)
	}
	if len(got.Result) != 2 || got.Result[0].ID != "otp_1" {
		t.Fatalf("unexpected list result: %+v", got)
	}
	if got.Pagination.Count != 2 || got.Pagination.Next != "cur_3" {
		t.Fatalf("unexpected pagination: %+v", got.Pagination)
	}
}

func TestUssdOtp_Delete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/ussd-otps/otp_1" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if err := c.UssdOtp().Delete(context.Background(), "otp_1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUssdOtp_Delete_RequiresID(t *testing.T) {
	c := newTestClient(t, httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})))
	if err := c.UssdOtp().Delete(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty id")
	}
}
