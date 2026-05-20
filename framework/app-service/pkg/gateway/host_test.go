package gateway

import (
	"errors"
	"reflect"
	"testing"
)

func TestNormalizeHostPattern(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "lowercase no port", in: "abc.shared.example.com", want: "abc.shared.example.com"},
		{name: "trim spaces", in: "  abc.shared.example.com  ", want: "abc.shared.example.com"},
		{name: "strip port", in: "abc.shared.example.com:11434", want: "abc.shared.example.com"},
		{name: "uppercase", in: "ABC.SHARED.EXAMPLE.com", want: "abc.shared.example.com"},
		{name: "mixed case + port", in: "ABc.Shared.Example.COM:8080", want: "abc.shared.example.com"},
		{name: "single label", in: "ollama", want: "ollama"},
		{name: "with digits", in: "a1b2c3d40.shared.example.com", want: "a1b2c3d40.shared.example.com"},

		{name: "empty", in: "", wantErr: true},
		{name: "whitespace only", in: "  \t", wantErr: true},
		{name: "scheme http", in: "http://abc.shared.example.com", wantErr: true},
		{name: "scheme https", in: "https://abc.shared.example.com:443", wantErr: true},
		{name: "with path", in: "abc.shared.example.com/foo", wantErr: true},
		{name: "with query", in: "abc.shared.example.com?x=1", wantErr: true},
		{name: "with fragment", in: "abc.shared.example.com#h", wantErr: true},
		{name: "missing host before port", in: ":443", wantErr: true},
		{name: "missing port after colon", in: "abc.shared.example.com:", wantErr: true},
		{name: "non-numeric port", in: "abc.shared.example.com:abc", wantErr: true},
		{name: "leading dash", in: "-abc.example.com", wantErr: true},
		{name: "trailing dash", in: "abc.example.com-", wantErr: true},
		{name: "underscore", in: "abc_d.example.com", wantErr: true},
		{name: "wildcard star", in: "*.example.com", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NormalizeHostPattern(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("NormalizeHostPattern(%q) = %q, want error", tc.in, got)
				}
				if !errors.Is(err, ErrInvalidHostPattern) {
					t.Fatalf("error is not ErrInvalidHostPattern: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeHostPattern(%q) unexpected error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Fatalf("NormalizeHostPattern(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNormalizeHostPatterns(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		got, err := NormalizeHostPatterns(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
	t.Run("dedup preserves order", func(t *testing.T) {
		in := []string{"A.example.com:80", "a.EXAMPLE.com", "b.example.com"}
		want := []string{"a.example.com", "b.example.com"}
		got, err := NormalizeHostPatterns(in)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})
	t.Run("aborts on first invalid", func(t *testing.T) {
		in := []string{"valid.example.com", "http://x"}
		if _, err := NormalizeHostPatterns(in); err == nil {
			t.Fatalf("expected error from invalid second entry")
		}
	})
}
