package controller

import (
	"reflect"
	"regexp"
	"testing"
)

func TestParseLogicalPattern(t *testing.T) {
	good := map[string]LogicalPattern{
		"01234567.*.olares.com": {Hash8: "01234567", PlatformDomain: "olares.com"},
		"deadbeef.*.example.io": {Hash8: "deadbeef", PlatformDomain: "example.io"},
		"abcdef01.*.a.b.c":      {Hash8: "abcdef01", PlatformDomain: "a.b.c"},
	}
	for in, want := range good {
		got, ok := ParseLogicalPattern(in)
		if !ok || got != want {
			t.Fatalf("ParseLogicalPattern(%q) = %+v ok=%v, want %+v", in, got, ok, want)
		}
	}
	bad := []string{
		"",
		"abc.example.com",         // no wildcard
		"abc.*.com",               // hash too short
		"01234567x.*.olares.com",  // hash 9
		"0123456g.*.olares.com",   // non-hex
		"01234567.*",              // missing domain
		"01234567.*x.olares.com",  // wildcard not on a standalone label
		"01234567.*.olares.*",     // extra wildcard
	}
	for _, in := range bad {
		if _, ok := ParseLogicalPattern(in); ok {
			t.Fatalf("ParseLogicalPattern(%q) ok=true, want false", in)
		}
	}

	// Mixed-case inputs are normalised, not rejected.
	if got, ok := ParseLogicalPattern("01234567.*.OLARES.COM"); !ok || got.PlatformDomain != "olares.com" {
		t.Fatalf("ParseLogicalPattern uppercase domain: got %+v ok=%v", got, ok)
	}
}

func TestHostRegexValue(t *testing.T) {
	p := LogicalPattern{Hash8: "01234567", PlatformDomain: "olares.com"}
	re := HostRegexValue(p)
	want := `^01234567\.[a-z0-9]([-a-z0-9]*[a-z0-9])?\.olares\.com$`
	if re != want {
		t.Fatalf("HostRegexValue = %q, want %q", re, want)
	}

	compiled := regexp.MustCompile(re)
	cases := map[string]bool{
		"01234567.brucedai.olares.com":   true,
		"01234567.alice.olares.com":      true,
		"01234567.bob.olares.com":        true,
		"01234567.user-name.olares.com":  true,
		"01234567.x.olares.com":          true,
		"01234567.X.olares.com":          false, // upper-case viewer
		"01234567..olares.com":           false, // empty viewer
		"01234568.brucedai.olares.com":   false, // wrong hash
		"01234567.brucedai.example.com":  false, // wrong domain
		"01234567.brucedai.olares-com":   false,
		"01234567.brucedai.olares.comm":  false,
		"01234567.alice.bob.olares.com":  false, // multi-label viewer
		"01234567.foo_bar.olares.com":    false, // underscore
		"01234567.-foo.olares.com":       false,
		"01234567.foo-.olares.com":       false,
	}
	for in, want := range cases {
		if got := compiled.MatchString(in); got != want {
			t.Fatalf("regex.MatchString(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestMaterializeHostnames(t *testing.T) {
	got := MaterializeHostnames([]string{
		"01234567.*.olares.com",
		"deadbeef.*.olares.com",       // same wildcard host -> dedup
		"abc.shared.example.com",
		"abc.shared.example.com",
	})
	want := []any{"*.olares.com", "abc.shared.example.com"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("MaterializeHostnames: got %v, want %v", got, want)
	}
}

func TestMaterializeHostHeaders(t *testing.T) {
	got := MaterializeHostHeaders([]string{
		"01234567.*.olares.com",
		"deadbeef.*.olares.com",
		"01234567.*.olares.com", // dedup by regex value
		"abc.shared.example.com",
	})
	if len(got) != 2 {
		t.Fatalf("expected 2 header matches, got %v", got)
	}
	values := []string{got[0]["value"].(string), got[1]["value"].(string)}
	wantRE := []string{
		HostRegexValue(LogicalPattern{Hash8: "01234567", PlatformDomain: "olares.com"}),
		HostRegexValue(LogicalPattern{Hash8: "deadbeef", PlatformDomain: "olares.com"}),
	}
	if !reflect.DeepEqual(values, wantRE) {
		t.Fatalf("regex values mismatch: got %v want %v", values, wantRE)
	}
}

func TestMaterializeHostHeaders_ExactOnly(t *testing.T) {
	got := MaterializeHostHeaders([]string{"abc.example.com"})
	if len(got) != 0 {
		t.Fatalf("exact-only hosts must not produce header matches: %v", got)
	}
}
