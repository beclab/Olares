package routecontrol

import "testing"

func TestParseLogicalPattern(t *testing.T) {
	p, ok := ParseLogicalPattern("ab12cd34.*.olares.com")
	if !ok {
		t.Fatal("expected valid logical pattern")
	}
	if p.Hash8 != "ab12cd34" || p.PlatformDomain != "olares.com" {
		t.Errorf("parsed = %+v", p)
	}
	for _, bad := range []string{"a.example.com", "xyz.*.olares.com", "ab12cd34.*.", "ZZ12cd34.*.olares.com"} {
		if _, ok := ParseLogicalPattern(bad); ok {
			t.Errorf("ParseLogicalPattern(%q) should be invalid", bad)
		}
	}
}

func TestHostRegexValue(t *testing.T) {
	got := HostRegexValue(LogicalPattern{Hash8: "ab12cd34", PlatformDomain: "olares.com"})
	want := `^ab12cd34\.[a-z0-9]([-a-z0-9]*[a-z0-9])?\.olares\.com$`
	if got != want {
		t.Errorf("HostRegexValue = %q, want %q", got, want)
	}
}

func TestHTTPRouteHostnamesAndHeaderMatches(t *testing.T) {
	patterns := []string{"ab12cd34.*.olares.com", "exact.example.com", "ab12cd34.*.olares.com"}
	hosts := HTTPRouteHostnames(patterns)
	if len(hosts) != 2 {
		t.Fatalf("hosts = %v, want 2 deduped", hosts)
	}
	if hosts[0] != "*.olares.com" || hosts[1] != "exact.example.com" {
		t.Errorf("hosts = %v", hosts)
	}
	matches := HTTPRouteHostHeaderMatches(patterns)
	if len(matches) != 1 {
		t.Fatalf("header matches = %v, want 1", matches)
	}
	if matches[0]["type"] != "RegularExpression" || matches[0]["name"] != "Host" {
		t.Errorf("header match = %v", matches[0])
	}
}
