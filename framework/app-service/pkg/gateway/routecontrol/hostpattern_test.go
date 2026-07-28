package routecontrol

import "testing"

func TestParseLogicalPattern(t *testing.T) {
	p, ok := ParseLogicalPattern("ab12cd34.*.olares.com")
	if !ok {
		t.Fatal("expected valid logical pattern")
	}
	if p.Prefix != "ab12cd34" || p.PlatformDomain != "olares.com" {
		t.Errorf("parsed = %+v", p)
	}
	app, ok := ParseLogicalPattern("e31111940.*.olares.cn")
	if !ok {
		t.Fatal("expected valid application entrance logical pattern")
	}
	if app.Prefix != "e31111940" || app.PlatformDomain != "olares.cn" {
		t.Errorf("application parsed = %+v", app)
	}
	hosts := HTTPRouteHostnames([]string{"e31111940.*.olares.cn"})
	if len(hosts) != 1 || hosts[0] != "*.olares.cn" {
		t.Fatalf("application HTTPRouteHostnames = %v, want [*.olares.cn]", hosts)
	}
	for _, bad := range []string{"a.example.com", "*.olares.com", "ab12cd34.*.", "-bad.*.olares.com"} {
		if _, ok := ParseLogicalPattern(bad); ok {
			t.Errorf("ParseLogicalPattern(%q) should be invalid", bad)
		}
	}
}

func TestHostRegexValue(t *testing.T) {
	got := HostRegexValue(LogicalPattern{Prefix: "ab12cd34", PlatformDomain: "olares.com"})
	want := `^ab12cd34\.[a-z0-9]([-a-z0-9]*[a-z0-9])?\.olares\.com$`
	if got != want {
		t.Errorf("HostRegexValue = %q, want %q", got, want)
	}
	appGot := HostRegexValue(LogicalPattern{Prefix: "e31111940", PlatformDomain: "olares.cn"})
	appWant := `^e31111940\.[a-z0-9]([-a-z0-9]*[a-z0-9])?\.olares\.cn$`
	if appGot != appWant {
		t.Errorf("application HostRegexValue = %q, want %q", appGot, appWant)
	}
}

func TestHTTPRouteHostnamesAndHeaderMatches(t *testing.T) {
	patterns := []string{"ab12cd34.*.olares.com", "ab12cd34.shared.olares.com", "exact.example.com", "ab12cd34.*.olares.com"}
	hosts := HTTPRouteHostnames(patterns)
	if len(hosts) != 3 {
		t.Fatalf("hosts = %v, want 3 deduped", hosts)
	}
	if hosts[0] != "*.olares.com" || hosts[1] != "ab12cd34.shared.olares.com" || hosts[2] != "exact.example.com" {
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
