package monime

import (
	"errors"
	"testing"
)

func boolPtr(b bool) *bool { return &b }

func TestCreateWebhookParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  CreateWebhookParams
		wantErr bool
	}{
		{
			name:    "valid minimal",
			params:  CreateWebhookParams{Name: "hook", URL: "https://example.com/hook", Events: []string{"payment.succeeded"}},
			wantErr: false,
		},
		{
			name: "valid full",
			params: CreateWebhookParams{
				Name:               "hook",
				URL:                "https://example.com/hook",
				Events:             []string{"payment.succeeded"},
				Enabled:            boolPtr(true),
				APIRelease:         "caph",
				Headers:            map[string]string{"X-Token": "abc"},
				AlertEmails:        []string{"ops@example.com"},
				VerificationMethod: &WebhookVerificationMethod{Type: "HS256", Secret: "whsec_0123456789abcdef0123456789abcdef"},
				Metadata:           map[string]any{"team": "payments"},
			},
			wantErr: false,
		},
		{
			name:    "missing name",
			params:  CreateWebhookParams{URL: "https://example.com/hook", Events: []string{"e"}},
			wantErr: true,
		},
		{
			name:    "missing url",
			params:  CreateWebhookParams{Name: "hook", Events: []string{"e"}},
			wantErr: true,
		},
		{
			name:    "invalid url",
			params:  CreateWebhookParams{Name: "hook", URL: "not-a-url", Events: []string{"e"}},
			wantErr: true,
		},
		{
			name:    "empty events",
			params:  CreateWebhookParams{Name: "hook", URL: "https://example.com/hook", Events: []string{}},
			wantErr: true,
		},
		{
			name:    "nil events",
			params:  CreateWebhookParams{Name: "hook", URL: "https://example.com/hook"},
			wantErr: true,
		},
		{
			name:    "invalid alert email",
			params:  CreateWebhookParams{Name: "hook", URL: "https://example.com/hook", Events: []string{"e"}, AlertEmails: []string{"nope"}},
			wantErr: true,
		},
		{
			name:    "invalid verification type",
			params:  CreateWebhookParams{Name: "hook", URL: "https://example.com/hook", Events: []string{"e"}, VerificationMethod: &WebhookVerificationMethod{Type: "RS256", Secret: "s"}},
			wantErr: true,
		},
		{
			name:    "missing verification secret",
			params:  CreateWebhookParams{Name: "hook", URL: "https://example.com/hook", Events: []string{"e"}, VerificationMethod: &WebhookVerificationMethod{Type: "HS256"}},
			wantErr: true,
		},
		{
			name:    "HS256 secret shorter than 32 characters",
			params:  CreateWebhookParams{Name: "hook", URL: "https://example.com/hook", Events: []string{"e"}, VerificationMethod: &WebhookVerificationMethod{Type: "HS256", Secret: "too-short"}},
			wantErr: true,
		},
		{
			name:    "ES256 needs no secret",
			params:  CreateWebhookParams{Name: "hook", URL: "https://example.com/hook", Events: []string{"e"}, VerificationMethod: &WebhookVerificationMethod{Type: "ES256"}},
			wantErr: false,
		},
		{
			name:    "ES256 with a secret",
			params:  CreateWebhookParams{Name: "hook", URL: "https://example.com/hook", Events: []string{"e"}, VerificationMethod: &WebhookVerificationMethod{Type: "ES256", Secret: "whsec_0123456789abcdef0123456789abcdef"}},
			wantErr: true,
		},
		{
			name:    "unknown api release",
			params:  CreateWebhookParams{Name: "hook", URL: "https://example.com/hook", Events: []string{"e"}, APIRelease: "v2"},
			wantErr: true,
		},
		{
			name:    "three alert emails",
			params:  CreateWebhookParams{Name: "hook", URL: "https://example.com/hook", Events: []string{"e"}, AlertEmails: []string{"a@x.io", "b@x.io", "c@x.io"}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStruct(tt.params)
			if tt.wantErr && err == nil {
				t.Fatal("expected validation error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}
			if tt.wantErr {
				var valErr *ValidationError
				if !errors.As(err, &valErr) {
					t.Fatalf("expected *ValidationError, got %T", err)
				}
			}
		})
	}
}

func TestUpdateWebhookParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  UpdateWebhookParams
		wantErr bool
	}{
		{
			name:    "valid empty (all optional)",
			params:  UpdateWebhookParams{},
			wantErr: false,
		},
		{
			name:    "valid single field",
			params:  UpdateWebhookParams{Enabled: boolPtr(false)},
			wantErr: false,
		},
		{
			name:    "valid url",
			params:  UpdateWebhookParams{URL: "https://example.com/new"},
			wantErr: false,
		},
		{
			name:    "invalid url",
			params:  UpdateWebhookParams{URL: "nope"},
			wantErr: true,
		},
		{
			name:    "nil events omitted",
			params:  UpdateWebhookParams{Events: nil},
			wantErr: false, // nil slice is omitted -> treated as not provided
		},
		{
			name:    "valid events",
			params:  UpdateWebhookParams{Events: []string{"e"}},
			wantErr: false,
		},
		{
			name:    "empty events rejected",
			params:  UpdateWebhookParams{Events: []string{}},
			wantErr: true, // mirrors zod min(1): an empty array fails
		},
		{
			name:    "invalid alert email",
			params:  UpdateWebhookParams{AlertEmails: []string{"bad"}},
			wantErr: true,
		},
		{
			name:    "invalid api release",
			params:  UpdateWebhookParams{APIRelease: "v2"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStruct(tt.params)
			if tt.wantErr && err == nil {
				t.Fatal("expected validation error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}
		})
	}
}
