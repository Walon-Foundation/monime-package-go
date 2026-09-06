package monime

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
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

func TestDo_PrefersMonimeRequestIDHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Monime-Request-Id", "mon-req-1")
		w.Header().Set("x-request-id", "legacy")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.do(context.Background(), requestOptions{method: http.MethodGet, path: "/x", out: &sampleResult{}})

	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *Error, got %T", err)
	}
	if apiErr.RequestID != "mon-req-1" {
		t.Fatalf("requestID = %q, want mon-req-1", apiErr.RequestID)
	}
}

func TestDo_ParsesErrorEnvelope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Monime-Request-Id", "req-409")
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"success":false,"messages":[],"error":{"code":409,` +
			`"reason":"idempotency_key_in_use",` +
			`"message":"Conflict: Idempotency key reused with a non-identical request.",` +
			`"details":["dup"]}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.do(context.Background(), requestOptions{method: http.MethodPost, path: "/x", out: &sampleResult{}})

	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *Error, got %T", err)
	}
	if apiErr.Message != "Conflict: Idempotency key reused with a non-identical request." {
		t.Errorf("message = %q", apiErr.Message)
	}
	if apiErr.Reason != "idempotency_key_in_use" {
		t.Errorf("reason = %q", apiErr.Reason)
	}
	if apiErr.Code != 409 || apiErr.Status != http.StatusConflict {
		t.Errorf("code/status = %d/%d", apiErr.Code, apiErr.Status)
	}
	if apiErr.RequestID != "req-409" {
		t.Errorf("requestID = %q", apiErr.RequestID)
	}
	details, ok := apiErr.Details.([]any)
	if !ok || len(details) != 1 || details[0] != "dup" {
		t.Errorf("details = %#v, want the envelope's error.details", apiErr.Details)
	}
}

func TestDo_RateLimitError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "3")
		w.Header().Set("Monime-Rate-Limit", "endpoint-limit")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"success":false,"messages":[],"error":{"code":429,` +
			`"reason":"too_many_requests","message":"Too many requests sent in a short period","details":[]}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.do(context.Background(), requestOptions{method: http.MethodGet, path: "/x", out: &sampleResult{}})

	var rateErr *RateLimitError
	if !errors.As(err, &rateErr) {
		t.Fatalf("expected *RateLimitError, got %T (%v)", err, err)
	}
	if rateErr.RetryAfter != 3*time.Second {
		t.Errorf("retryAfter = %v, want 3s", rateErr.RetryAfter)
	}
	if rateErr.Limit != "endpoint-limit" {
		t.Errorf("limit = %q", rateErr.Limit)
	}

	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatal("RateLimitError should unwrap to *Error")
	}
	if apiErr.Reason != "too_many_requests" || apiErr.Status != http.StatusTooManyRequests {
		t.Errorf("unwrapped = %+v", apiErr)
	}
}

func TestDo_RateLimitErrorWithoutRetryAfter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.do(context.Background(), requestOptions{method: http.MethodGet, path: "/x", out: &sampleResult{}})

	var rateErr *RateLimitError
	if !errors.As(err, &rateErr) {
		t.Fatalf("expected *RateLimitError, got %T", err)
	}
	if rateErr.RetryAfter != 0 {
		t.Errorf("retryAfter = %v, want 0 when the header is absent", rateErr.RetryAfter)
	}
}

func TestDo_EncodesQueryParameters(t *testing.T) {
	var gotRawQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRawQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"result":{"id":"1"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	var out sampleResult
	err := c.do(context.Background(), requestOptions{
		method: http.MethodGet,
		path:   "/x",
		query:  url.Values{"limit": {"25"}, "after": {"cur sor"}},
		out:    &out,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotRawQuery != "after=cur+sor&limit=25" {
		t.Fatalf("raw query = %q", gotRawQuery)
	}
}

func TestDo_OmitsEmptyQuery(t *testing.T) {
	var gotURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.String()
		_, _ = w.Write([]byte(`{"result":{"id":"1"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	var out sampleResult
	if err := c.do(context.Background(), requestOptions{
		method: http.MethodGet,
		path:   "/x",
		query:  url.Values{},
		out:    &out,
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotURL != "/x" {
		t.Fatalf("url = %q, want /x with no trailing ?", gotURL)
	}
}
