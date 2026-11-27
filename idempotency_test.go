package monime

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// The same vectors are pinned in the TypeScript SDK, so both derive the same key
// from the same reference.
func TestIdempotencyKeyFor_Vectors(t *testing.T) {
	for ref, want := range map[string]string{
		"payout:line-42:1": "61d2f738-ba9c-57c7-bf2e-03fd90e25927",
		"account:org-7":    "17443694-14ee-5556-83ab-601bf47c3684",
		"a":                "71089080-09e5-5194-b507-0949de5547cd",
	} {
		if got := IdempotencyKeyFor(ref); got != want {
			t.Errorf("IdempotencyKeyFor(%q) = %s, want %s", ref, got, want)
		}
	}
}

func TestIdempotencyKey_CallerKeyIsSentAndReused(t *testing.T) {
	var keys []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		keys = append(keys, r.Header.Get("Idempotency-Key"))
		_, _ = w.Write([]byte(`{"result":{"id":"pyt-1","status":"pending"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	key := IdempotencyKeyFor("payout:line-42:1")
	ctx := WithIdempotencyKey(context.Background(), key)
	params := CreatePayoutParams{
		Amount:        1000,
		SourceAccount: "fac-1",
		Destination:   PayoutDestination{Type: "momo", ProviderID: "m17", PhoneNumber: "+23276123456"},
	}
	for i := 0; i < 2; i++ {
		if _, err := c.Payout().Create(ctx, params); err != nil {
			t.Fatal(err)
		}
	}
	if len(keys) != 2 || keys[0] != key || keys[1] != key {
		t.Fatalf("a retry must send the caller's key both times, got %v", keys)
	}
}

func TestIdempotencyKey_GeneratedWhenNotGiven(t *testing.T) {
	var keys []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		keys = append(keys, r.Header.Get("Idempotency-Key"))
		_, _ = w.Write([]byte(`{"result":{"id":"fac-1"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	for i := 0; i < 2; i++ {
		if _, err := c.FinancialAccount().Create(context.Background(), CreateFinancialAccountParams{AccountName: "A", Currency: CurrencySLE}); err != nil {
			t.Fatal(err)
		}
	}
	if keys[0] == "" || keys[0] == keys[1] {
		t.Fatalf("expected two different generated keys, got %v", keys)
	}
}

func TestIdempotencyKey_RejectsBadKeys(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("server should not be called with an invalid idempotency key")
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	for _, key := range []string{"", strings.Repeat("k", 65)} {
		ctx := WithIdempotencyKey(context.Background(), key)
		_, err := c.FinancialAccount().Create(ctx, CreateFinancialAccountParams{AccountName: "A", Currency: CurrencySLE})
		var vErr *ValidationError
		if !errors.As(err, &vErr) {
			t.Errorf("key of length %d: want *ValidationError, got %v", len(key), err)
		}
	}
}

func TestNew_DefaultsMonimeVersion(t *testing.T) {
	t.Setenv("MONIME_VERSION", "")
	c, err := New(WithSpaceID("spc-1"), WithAccessToken("tok"))
	if err != nil {
		t.Fatal(err)
	}
	if c.version != DefaultVersion || DefaultVersion != "caph.2025-08-23" {
		t.Fatalf("version = %q, want %q", c.version, DefaultVersion)
	}
}
