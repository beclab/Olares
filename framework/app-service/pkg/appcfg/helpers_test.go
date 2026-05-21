// Package appcfg helpers test: per-viewer Shared URL and logical hostPattern.
//
// Covers SharedEntranceHostPrefix / GenSharedEntranceURLForUser /
// LogicalHostPattern. These three are the only public surfaces v2 adds to
// appcfg and are consumed by:
//   - pkg/gateway   (SRR writer)
//   - controllers   (reconcile loop)
//   - cmd l4-bfl-proxy  (informally — viewers obtain the same URL via
//                        SharedEntranceHostPrefix re-implementation; we keep
//                        the canonical computation here as the test oracle).
package appcfg

import (
	"strings"
	"testing"
)

func TestSharedEntranceHostPrefix_Stable(t *testing.T) {
	got := SharedEntranceHostPrefix("a5be2268", "ollamav2")
	if len(got) != 8 {
		t.Fatalf("len(hash8) = %d, want 8", len(got))
	}
	for _, r := range got {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			t.Fatalf("hash8 %q has non-lowercase-hex rune %q", got, r)
		}
	}
	// md5("a5be2268:shared:ollamav2") prefix is a stable contract; the
	// digest is also computed by l4-bfl-proxy and e2e scripts (Python).
	again := SharedEntranceHostPrefix("a5be2268", "ollamav2")
	if got != again {
		t.Fatalf("hash8 not stable: %q vs %q", got, again)
	}
	if SharedEntranceHostPrefix("A5BE2268", " OllamaV2 ") != got {
		t.Fatalf("hash8 must be case/whitespace insensitive: %q vs %q",
			SharedEntranceHostPrefix("A5BE2268", " OllamaV2 "), got)
	}
}

func TestSharedEntranceHostPrefix_DiffersByEntrance(t *testing.T) {
	a := SharedEntranceHostPrefix("a5be2268", "ollamav2")
	b := SharedEntranceHostPrefix("a5be2268", "ollamav3")
	if a == b {
		t.Fatalf("entrance change must change hash8: got %q for both", a)
	}
}

func TestGenSharedEntranceURLForUser(t *testing.T) {
	cases := []struct {
		name     string
		appid    string
		entrance string
		viewer   string
		domain   string
		want     string
	}{
		{
			name: "happy path",
			appid: "a5be2268", entrance: "ollamav2",
			viewer: "alice", domain: "olares.com",
			want: "https://" + SharedEntranceHostPrefix("a5be2268", "ollamav2") + ".alice.olares.com",
		},
		{
			name: "uppercase normalized",
			appid: "A5BE2268", entrance: "OLLAMAv2",
			viewer: "Alice", domain: "OLARES.COM",
			want: "https://" + SharedEntranceHostPrefix("a5be2268", "ollamav2") + ".alice.olares.com",
		},
		{
			name: "trailing dot in domain stripped",
			appid: "a5be2268", entrance: "ollamav2",
			viewer: "alice", domain: "olares.com.",
			want: "https://" + SharedEntranceHostPrefix("a5be2268", "ollamav2") + ".alice.olares.com",
		},
		{name: "empty viewer", appid: "x", entrance: "y", viewer: "", domain: "olares.com", want: ""},
		{name: "empty domain", appid: "x", entrance: "y", viewer: "v", domain: "", want: ""},
		{name: "empty appid", appid: "", entrance: "y", viewer: "v", domain: "olares.com", want: ""},
		{name: "empty entrance", appid: "x", entrance: "", viewer: "v", domain: "olares.com", want: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := GenSharedEntranceURLForUser(tc.appid, tc.entrance, tc.viewer, tc.domain)
			if got != tc.want {
				t.Fatalf("GenSharedEntranceURLForUser: got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestGenSharedEntranceURLForUser_DiffersByViewer(t *testing.T) {
	a := GenSharedEntranceURLForUser("a5be2268", "ollamav2", "alice", "olares.com")
	b := GenSharedEntranceURLForUser("a5be2268", "ollamav2", "alice", "olares.com")
	if a == b {
		t.Fatalf("viewer change must change URL (got %q == %q)", a, b)
	}
	if !strings.Contains(a, ".alice.") || !strings.Contains(b, ".alice.") {
		t.Fatalf("URL must contain viewer label: %q / %q", a, b)
	}
}

func TestLogicalHostPattern(t *testing.T) {
	got := LogicalHostPattern("a5be2268", "ollamav2", "olares.com")
	want := SharedEntranceHostPrefix("a5be2268", "ollamav2") + ".*.olares.com"
	if got != want {
		t.Fatalf("LogicalHostPattern: got %q, want %q", got, want)
	}
	if g := LogicalHostPattern("A5BE2268", "OllamaV2", "OLARES.COM."); g != want {
		t.Fatalf("LogicalHostPattern not normalized: got %q, want %q", g, want)
	}
	if LogicalHostPattern("", "x", "y") != "" || LogicalHostPattern("x", "", "y") != "" || LogicalHostPattern("x", "y", "") != "" {
		t.Fatal("LogicalHostPattern: empty inputs must return empty string")
	}
}

// Cross-helper invariant: URL host part for viewer "x" is exactly
// "<LogicalHostPattern> with * replaced by x".
func TestURL_LogicalPattern_Consistency(t *testing.T) {
	pat := LogicalHostPattern("a5be2268", "ollamav2", "olares.com")
	url := GenSharedEntranceURLForUser("a5be2268", "ollamav2", "alice", "olares.com")
	wantHost := strings.Replace(pat, "*", "alice", 1)
	if url != "https://"+wantHost {
		t.Fatalf("URL %q is inconsistent with logical pattern %q", url, pat)
	}
}
