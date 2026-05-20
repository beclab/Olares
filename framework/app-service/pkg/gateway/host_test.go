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

func TestNormalizeHostOrLogicalPattern_AcceptsExact(t *testing.T) {
	cases := map[string]string{
		"abc.shared.example.com":    "abc.shared.example.com",
		"ABC.shared.EXAMPLE.com":    "abc.shared.example.com",
		"ABC.shared.EXAMPLE.com:80": "abc.shared.example.com",
	}
	for in, want := range cases {
		got, err := NormalizeHostOrLogicalPattern(in)
		if err != nil {
			t.Fatalf("NormalizeHostOrLogicalPattern(%q): unexpected error %v", in, err)
		}
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	}
}

func TestNormalizeHostOrLogicalPattern_Logical(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"v2 happy path", "01234567.*.olares.com", "01234567.*.olares.com"},
		{"uppercase v2", "01234567.*.OLARES.COM", "01234567.*.olares.com"},
		{"v2 deeper domain", "deadbeef.*.olares.com", "deadbeef.*.olares.com"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NormalizeHostOrLogicalPattern(tc.in)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestNormalizeHostOrLogicalPattern_Rejects(t *testing.T) {
	bad := []string{
		"",
		"http://01234567.*.olares.com",
		"01234567.*.olares.com/foo",
		"*.olares.com",
		"abc.*.def.*.com",
		"01234567.*", // only two labels
		"01234567.*x.olares.com",
	}
	for _, in := range bad {
		if _, err := NormalizeHostOrLogicalPattern(in); err == nil {
			t.Fatalf("NormalizeHostOrLogicalPattern(%q) = nil error, want failure", in)
		}
	}
}

func TestIsLogicalHostPattern(t *testing.T) {
	yes := []string{"01234567.*.olares.com", "deadbeef.*.example.io"}
	no := []string{"abc.example.com", "example.com", ""}
	for _, s := range yes {
		if !IsLogicalHostPattern(s) {
			t.Fatalf("IsLogicalHostPattern(%q) = false, want true", s)
		}
	}
	for _, s := range no {
		if IsLogicalHostPattern(s) {
			t.Fatalf("IsLogicalHostPattern(%q) = true, want false", s)
		}
	}
}

func TestNormalizeHostOrLogicalPatterns(t *testing.T) {
	in := []string{"ABC.example.com", "01234567.*.olares.com", "abc.example.com"}
	want := []string{"abc.example.com", "01234567.*.olares.com"}
	got, err := NormalizeHostOrLogicalPatterns(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("got %v, want %v", got, want)
	}
}
