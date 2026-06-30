package monime

import (
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("missing space id returns error", func(t *testing.T) {
		_, err := New(WithAccessToken("tok"))
		if err == nil {
			t.Fatal("expected error when space id is missing")
		}
	})

	t.Run("missing access token returns error", func(t *testing.T) {
		_, err := New(WithSpaceID("space"))
		if err == nil {
			t.Fatal("expected error when access token is missing")
		}
	})

	t.Run("succeeds with both credentials", func(t *testing.T) {
		c, err := New(WithSpaceID("space"), WithAccessToken("tok"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.spaceID != "space" || c.accessToken != "tok" {
			t.Fatalf("credentials not stored: %+v", c)
		}
		if c.baseURL != defaultBaseURL {
			t.Fatalf("baseURL = %q, want %q", c.baseURL, defaultBaseURL)
		}
	})

	t.Run("falls back to environment variables", func(t *testing.T) {
		t.Setenv("MONIME_SPACE_ID", "env-space")
		t.Setenv("MONIME_ACCESS_TOKEN", "env-tok")
		t.Setenv("MONIME_VERSION", string(Version20250823))

		c, err := New()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.spaceID != "env-space" || c.accessToken != "env-tok" {
			t.Fatalf("env credentials not applied: %+v", c)
		}
		if c.version != Version20250823 {
			t.Fatalf("version = %q, want %q", c.version, Version20250823)
		}
	})

	t.Run("options override environment", func(t *testing.T) {
		t.Setenv("MONIME_SPACE_ID", "env-space")
		t.Setenv("MONIME_ACCESS_TOKEN", "env-tok")

		c, err := New(WithSpaceID("opt-space"), WithAccessToken("opt-tok"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.spaceID != "opt-space" || c.accessToken != "opt-tok" {
			t.Fatalf("options did not override env: %+v", c)
		}
	})

	t.Run("WithBaseURL trims trailing slash", func(t *testing.T) {
		c, err := New(WithSpaceID("s"), WithAccessToken("t"), WithBaseURL("http://localhost:8080/"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.baseURL != "http://localhost:8080" {
			t.Fatalf("baseURL = %q, want trimmed", c.baseURL)
		}
	})

	t.Run("config error is a *Error", func(t *testing.T) {
		_, err := New()
		var apiErr *Error
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})
}
