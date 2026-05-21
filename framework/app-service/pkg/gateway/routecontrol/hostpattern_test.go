package routecontrol

import (
	"regexp"
	"testing"
)

func TestParseLogicalPattern(t *testing.T) {
	cases := []struct {
		in   string
		ok   bool
		hash string
		dom  string
	}{
		{"01234567.*.olares.com", true, "01234567", "olares.com"},
		{"ABcdef00.*.olares.com", true, "abcdef00", "olares.com"},
		{"shortbad.*.olares.com", false, "", ""},
		{"0123456.*.olares.com", false, "", ""},
		{"0123456g.*.olares.com", false, "", ""},
		{"01234567.viewer.olares.com", false, "", ""},
		{"01234567.*.", false, "", ""},
		{"01234567.*.*.com", false, "", ""},
	}
	for _, tc := range cases {
		got, ok := ParseLogicalPattern(tc.in)
		if ok != tc.ok {
			t.Fatalf("%q: ok=%v want %v", tc.in, ok, tc.ok)
		}
		if !ok {
			continue
		}
		if got.Hash8 != tc.hash || got.PlatformDomain != tc.dom {
			t.Fatalf("%q: got %+v", tc.in, got)
		}
	}
}

func TestHTTPRouteHostnames_DedupAndOrder(t *testing.T) {
	out := HTTPRouteHostnames([]string{
		"01234567.*.olares.com",
		"abcdef00.*.olares.com",
		"01234567.*.olares.com",
		"verbatim.olares.com",
		"VERBATIM.olares.com",
	})
	if len(out) != 2 {
		t.Fatalf("want 2 entries, got %d: %v", len(out), out)
	}
	if out[0].(string) != "*.olares.com" {
		t.Fatalf("first entry: %v", out[0])
	}
	if out[1].(string) != "verbatim.olares.com" {
		t.Fatalf("second entry: %v", out[1])
	}
}

func TestHTTPRouteHostHeaderMatches_OnlyLogical(t *testing.T) {
	out := HTTPRouteHostHeaderMatches([]string{
		"01234567.*.olares.com",
		"verbatim.olares.com",
		"01234567.*.olares.com",
		"abcdef00.*.olares.com",
	})
	if len(out) != 2 {
		t.Fatalf("want 2 host headers, got %d: %v", len(out), out)
	}
	re := regexp.MustCompile(`^\^01234567\\\.\[a-z0-9\]\(\[-a-z0-9\]\*\[a-z0-9\]\)\?\\\.olares\\\.com\$$`)
	if !re.MatchString(out[0]["value"].(string)) {
		t.Fatalf("regex value mismatch: %s", out[0]["value"])
	}
}

func TestHostRegexValue_MatchesViewerLabel(t *testing.T) {
	p, ok := ParseLogicalPattern("01234567.*.olares.com")
	if !ok {
		t.Fatal("parse failed")
	}
	re := regexp.MustCompile(HostRegexValue(p))
	for _, host := range []string{
		"01234567.alice.olares.com",
		"01234567.bob-1.olares.com",
		"01234567.x.olares.com",
	} {
		if !re.MatchString(host) {
			t.Fatalf("should match %q", host)
		}
	}
	for _, host := range []string{
		"01234566.alice.olares.com",
		"01234567..olares.com",
		"01234567.UPPER.olares.com",
		"01234567.alice.example.com",
	} {
		if re.MatchString(host) {
			t.Fatalf("should NOT match %q", host)
		}
	}
}
