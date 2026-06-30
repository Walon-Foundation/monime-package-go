package monime

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebhook_Create(t *testing.T) {
	var gotMethod, gotPath, gotIdem string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotIdem = r.Method, r.URL.Path, r.Header.Get("Idempotency-Key")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"result":{"id":"wh_1","name":"hook","url":"https://example.com/hook","enabled":true,"events":["payment.succeeded"]}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.Webhook().Create(context.Background(), CreateWebhookParams{
		Name:               "hook",
		URL:                "https://example.com/hook",
		Events:             []string{"payment.succeeded"},
		Enabled:            boolPtr(true),
		VerificationMethod: &WebhookVerificationMethod{Type: "HS256", Secret: "supersecretvalue"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "wh_1" || got.Name != "hook" || !got.Enabled {
		t.Fatalf("unexpected response: %+v", got)
	}
	if gotMethod != http.MethodPost || gotPath != "/webhooks" {
		t.Fatalf("method/path = %s %s", gotMethod, gotPath)
	}
	if gotIdem == "" {
		t.Error("expected Idempotency-Key header")
	}
	if gotBody["name"] != "hook" || gotBody["url"] != "https://example.com/hook" {
		t.Errorf("body name/url = %v %v", gotBody["name"], gotBody["url"])
	}
	events, _ := gotBody["events"].([]any)
	if len(events) != 1 || events[0] != "payment.succeeded" {
		t.Errorf("body events = %v", gotBody["events"])
	}
	if gotBody["enabled"] != true {
		t.Errorf("body enabled = %v", gotBody["enabled"])
	}
	vm, _ := gotBody["verificationMethod"].(map[string]any)
	if vm["type"] != "HS256" || vm["secret"] != "supersecretvalue" {
		t.Errorf("body verificationMethod = %v", gotBody["verificationMethod"])
	}
}

func TestWebhook_Create_OmitsEmptyOptionals(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"result":{"id":"wh_2"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if _, err := c.Webhook().Create(context.Background(), CreateWebhookParams{
		Name:   "hook",
		URL:    "https://example.com/hook",
		Events: []string{"e"},
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, k := range []string{"enabled", "apiRelease", "headers", "alertEmails", "verificationMethod", "metadata"} {
		if _, ok := gotBody[k]; ok {
			t.Errorf("%s should be omitted when empty", k)
		}
	}
}

func TestWebhook_Create_ValidationShortCircuits(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("server should not be called when validation fails")
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.Webhook().Create(context.Background(), CreateWebhookParams{Name: ""})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestWebhook_Retrieve(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/webhooks/wh_9" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"result":{"id":"wh_9","enabled":false}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.Webhook().Retrieve(context.Background(), "wh_9")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "wh_9" || got.Enabled {
		t.Fatalf("unexpected: %+v", got)
	}
}

func TestWebhook_Retrieve_RequiresID(t *testing.T) {
	c := newTestClient(t, httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})))
	if _, err := c.Webhook().Retrieve(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty id")
	}
}

func TestWebhook_List(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/webhooks" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"result":[{"id":"wh_1"},{"id":"wh_2"}],"pagination":{"count":2,"next":"tok"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.Webhook().List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Result) != 2 || got.Pagination.Count != 2 || got.Pagination.Next != "tok" {
		t.Fatalf("unexpected list: %+v", got)
	}
}

func TestWebhook_Update(t *testing.T) {
	var gotMethod, gotPath, gotIdem string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotIdem = r.Method, r.URL.Path, r.Header.Get("Idempotency-Key")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"result":{"id":"wh_1","enabled":false}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.Webhook().Update(context.Background(), "wh_1", UpdateWebhookParams{Enabled: boolPtr(false)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPatch || gotPath != "/webhooks/wh_1" {
		t.Errorf("method/path = %s %s", gotMethod, gotPath)
	}
	if gotIdem == "" {
		t.Error("expected Idempotency-Key header")
	}
	if got.Enabled {
		t.Error("expected enable=false in response")
	}
	if gotBody["enabled"] != false {
		t.Errorf("body enabled = %v", gotBody["enabled"])
	}
}

func TestWebhook_Update_RequiresID(t *testing.T) {
	c := newTestClient(t, httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})))
	if _, err := c.Webhook().Update(context.Background(), "", UpdateWebhookParams{Enabled: boolPtr(true)}); err == nil {
		t.Fatal("expected error for empty id")
	}
}

func TestWebhook_Update_ValidationShortCircuits(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("server should not be called when validation fails")
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if _, err := c.Webhook().Update(context.Background(), "wh_1", UpdateWebhookParams{URL: "not-a-url"}); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestWebhook_Delete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/webhooks/wh_1" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if err := c.Webhook().Delete(context.Background(), "wh_1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWebhook_Delete_RequiresID(t *testing.T) {
	c := newTestClient(t, httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})))
	if err := c.Webhook().Delete(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty id")
	}
}
