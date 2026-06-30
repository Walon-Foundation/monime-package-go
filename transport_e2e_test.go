package monime

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestClient returns a Client pointed at srv with dummy credentials.
func newTestClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c, err := New(
		WithSpaceID("test-space"),
		WithAccessToken("test-token"),
		WithVersion(Version20250823),
		WithBaseURL(srv.URL),
		WithHTTPClient(srv.Client()),
	)
	if err != nil {
		t.Fatalf("failed to build test client: %v", err)
	}
	return c
}

type sampleResult struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestDo_SendsStandardHeaders(t *testing.T) {
	var gotAuth, gotSpace, gotVersion, gotIdem, gotContentType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotSpace = r.Header.Get("Monime-Space-Id")
		gotVersion = r.Header.Get("Monime-Version")
		gotIdem = r.Header.Get("Idempotency-Key")
		gotContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result":{"id":"1","name":"ok"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	var out sampleResult
	err := c.do(context.Background(), requestOptions{
		method:         http.MethodPost,
		path:           "/things",
		body:           map[string]string{"a": "b"},
		idempotencyKey: "idem-123",
		out:            &out,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotAuth != "Bearer test-token" {
		t.Errorf("Authorization = %q", gotAuth)
	}
	if gotSpace != "test-space" {
		t.Errorf("Monime-Space-Id = %q", gotSpace)
	}
	if gotVersion != string(Version20250823) {
		t.Errorf("Monime-Version = %q", gotVersion)
	}
	if gotIdem != "idem-123" {
		t.Errorf("Idempotency-Key = %q", gotIdem)
	}
	if gotContentType != "application/json" {
		t.Errorf("Content-Type = %q", gotContentType)
	}
}

func TestDo_UnwrapsResultEnvelope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"result":{"id":"abc","name":"widget"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	var out sampleResult
	if err := c.do(context.Background(), requestOptions{method: http.MethodGet, path: "/x", out: &out}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.ID != "abc" || out.Name != "widget" {
		t.Fatalf("unwrapped result = %+v", out)
	}
}

func TestDo_DecodesUnwrappedBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":"raw","name":"direct"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	var out sampleResult
	if err := c.do(context.Background(), requestOptions{method: http.MethodGet, path: "/x", out: &out}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.ID != "raw" {
		t.Fatalf("decoded body = %+v", out)
	}
}

func TestDo_RawBodyKeepsPagination(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"success":true,"result":[{"id":"1"},{"id":"2"}],"pagination":{"count":2,"next":"cur"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	var out struct {
		Result     []sampleResult `json:"result"`
		Pagination Pagination     `json:"pagination"`
	}
	err := c.do(context.Background(), requestOptions{method: http.MethodGet, path: "/x", out: &out, rawBody: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Result) != 2 {
		t.Fatalf("result len = %d, want 2", len(out.Result))
	}
	if out.Pagination.Count != 2 || out.Pagination.Next != "cur" {
		t.Fatalf("pagination = %+v", out.Pagination)
	}
}

func TestDo_NoContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if err := c.do(context.Background(), requestOptions{method: http.MethodDelete, path: "/x/1"}); err != nil {
		t.Fatalf("unexpected error on 204: %v", err)
	}
}

func TestDo_AuthenticationError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("x-request-id", "req-401")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"invalid token"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.do(context.Background(), requestOptions{method: http.MethodGet, path: "/x", out: &sampleResult{}})

	var authErr *AuthenticationError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected *AuthenticationError, got %T (%v)", err, err)
	}
	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected unwrap to *Error, got %T", err)
	}
	if apiErr.Status != http.StatusUnauthorized {
		t.Errorf("status = %d", apiErr.Status)
	}
	if apiErr.RequestID != "req-401" {
		t.Errorf("requestID = %q", apiErr.RequestID)
	}
	if apiErr.Message != "invalid token" {
		t.Errorf("message = %q", apiErr.Message)
	}
}

func TestDo_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.do(context.Background(), requestOptions{method: http.MethodGet, path: "/x", out: &sampleResult{}})

	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *Error, got %T", err)
	}
	if apiErr.Status != http.StatusInternalServerError || apiErr.Message != "boom" {
		t.Fatalf("unexpected error: %+v", apiErr)
	}

	var authErr *AuthenticationError
	if errors.As(err, &authErr) {
		t.Fatal("500 should not be an AuthenticationError")
	}
}

func TestDo_NetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close() // server is now down -> connection refused

	c, err := New(WithSpaceID("s"), WithAccessToken("t"), WithBaseURL(url))
	if err != nil {
		t.Fatalf("client build: %v", err)
	}
	err = c.do(context.Background(), requestOptions{method: http.MethodGet, path: "/x", out: &sampleResult{}})
	if err == nil {
		t.Fatal("expected network error")
	}
}
