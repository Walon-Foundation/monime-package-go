package monime

import (
	"errors"
	"testing"
)

func TestBuildListQuery(t *testing.T) {
	tests := []struct {
		name string
		opts []ListOption
		want string
	}{
		{
			name: "no options sends nothing",
			opts: nil,
			want: "",
		},
		{
			name: "paging",
			opts: []ListOption{WithLimit(50), WithAfter("cur-1")},
			want: "after=cur-1&limit=50",
		},
		{
			name: "empty cursor is dropped",
			opts: []ListOption{WithAfter("")},
			want: "",
		},
		{
			name: "nil option is skipped",
			opts: []ListOption{nil, WithLimit(1)},
			want: "limit=1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildListQuery(tt.opts)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Encode() != tt.want {
				t.Fatalf("query = %q, want %q", got.Encode(), tt.want)
			}
		})
	}
}

func TestBuildListQuery_Validation(t *testing.T) {
	tests := []struct {
		name string
		opt  ListOption
	}{
		{"limit below range", WithLimit(0)},
		{"limit above range", WithLimit(51)},
		{"negative limit", WithLimit(-1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := buildListQuery([]ListOption{tt.opt})
			if err == nil {
				t.Fatal("expected a validation error")
			}
			var valErr *ValidationError
			if !errors.As(err, &valErr) {
				t.Fatalf("expected *ValidationError, got %T", err)
			}
		})
	}
}
