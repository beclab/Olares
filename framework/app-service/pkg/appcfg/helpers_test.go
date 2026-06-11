package appcfg

import (
	"errors"
	"strings"
	"testing"
)

func TestSharedEntranceID_SingleAndMulti(t *testing.T) {
	got, err := SharedEntranceID("a5be2268", 0, 1)
	if err != nil {
		t.Fatalf("single entrance: %v", err)
	}
	if got != "a5be2268" {
		t.Fatalf("single entrance id = %q, want %q", got, "a5be2268")
	}

	got0, err := SharedEntranceID("A5BE2268", 0, 2)
	if err != nil {
		t.Fatalf("multi[0]: %v", err)
	}
	got1, err := SharedEntranceID("a5be2268", 1, 2)
	if err != nil {
		t.Fatalf("multi[1]: %v", err)
	}
	if got0 != "a5be22680" || got1 != "a5be22681" {
		t.Fatalf("multi entrance ids = %q/%q, want %q/%q", got0, got1, "a5be22680", "a5be22681")
	}
}

func TestSharedEntranceID_IndexOutOfRange(t *testing.T) {
	cases := []int{-1, 2, 3}
	for _, idx := range cases {
		_, err := SharedEntranceID("a5be2268", idx, 2)
		var eidErr *EIDError
		if !errors.As(err, &eidErr) {
			t.Fatalf("index=%d: expected EIDError, got %v", idx, err)
		}
		if eidErr.Code != "EID_INDEX_OUT_OF_RANGE" {
			t.Fatalf("index=%d: code=%q, want EID_INDEX_OUT_OF_RANGE", idx, eidErr.Code)
		}
	}
}

func TestSharedEntranceID_CountOutOfRange(t *testing.T) {
	cases := []int{0, 11, -1}
	for _, count := range cases {
		_, err := SharedEntranceID("a5be2268", 0, count)
		var eidErr *EIDError
		if !errors.As(err, &eidErr) {
			t.Fatalf("count=%d: expected EIDError, got %v", count, err)
		}
		if eidErr.Code != "EID_TOO_MANY_ENTRANCES" {
			t.Fatalf("count=%d: code=%q, want EID_TOO_MANY_ENTRANCES", count, eidErr.Code)
		}
	}
}

func TestSharedEntranceID_EmptyAppID(t *testing.T) {
	for _, appid := range []string{"", "   "} {
		_, err := SharedEntranceID(appid, 0, 1)
		var eidErr *EIDError
		if !errors.As(err, &eidErr) {
			t.Fatalf("appid=%q: expected EIDError, got %v", appid, err)
		}
		if eidErr.Code != "EID_EMPTY_APPID" {
			t.Fatalf("appid=%q: code=%q, want EID_EMPTY_APPID", appid, eidErr.Code)
		}
	}
}

func TestGenSharedEntranceURLForUser(t *testing.T) {
	cases := []struct {
		name    string
		appid   string
		idx     int
		count   int
		viewer  string
		domain  string
		want    string
		errCode string
	}{
		{
			name:  "happy path",
			appid: "a5be2268", idx: 0, count: 1,
			viewer: "alice", domain: "olares.com",
			want: "https://a5be2268.alice.olares.com",
		},
		{
			name:  "uppercase normalized",
			appid: "A5BE2268", idx: 1, count: 2,
			viewer: "Alice", domain: "OLARES.COM",
			want: "https://a5be22681.alice.olares.com",
		},
		{
			name:  "trailing dot in domain stripped",
			appid: "a5be2268", idx: 0, count: 1,
			viewer: "alice", domain: "olares.com.",
			want: "https://a5be2268.alice.olares.com",
		},
		{name: "empty viewer", appid: "x", idx: 0, count: 1, viewer: "", domain: "olares.com", errCode: "EID_INCOMPLETE_URL_INPUT"},
		{name: "empty domain", appid: "x", idx: 0, count: 1, viewer: "v", domain: "", errCode: "EID_INCOMPLETE_URL_INPUT"},
		{name: "empty appid", appid: "", idx: 0, count: 1, viewer: "v", domain: "olares.com", errCode: "EID_EMPTY_APPID"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := GenSharedEntranceURLForUser(tc.appid, tc.idx, tc.count, tc.viewer, tc.domain)
			if tc.errCode == "" {
				if err != nil {
					t.Fatalf("GenSharedEntranceURLForUser error: %v", err)
				}
				if got != tc.want {
					t.Fatalf("GenSharedEntranceURLForUser: got %q, want %q", got, tc.want)
				}
				return
			}
			var eidErr *EIDError
			if !errors.As(err, &eidErr) {
				t.Fatalf("expected EIDError, got %v", err)
			}
			if eidErr.Code != tc.errCode {
				t.Fatalf("error code=%q, want %q", eidErr.Code, tc.errCode)
			}
		})
	}
}

func TestGenSharedEntranceURLForUser_DiffersByViewer(t *testing.T) {
	a, err := GenSharedEntranceURLForUser("a5be2268", 0, 1, "alice", "olares.com")
	if err != nil {
		t.Fatalf("alice: %v", err)
	}
	b, err := GenSharedEntranceURLForUser("a5be2268", 0, 1, "bob", "olares.com")
	if err != nil {
		t.Fatalf("bob: %v", err)
	}
	if a == b {
		t.Fatalf("viewer change must change URL (got %q == %q)", a, b)
	}
	if !strings.Contains(a, ".alice.") || !strings.Contains(b, ".bob.") {
		t.Fatalf("URL must contain viewer label: %q / %q", a, b)
	}
}

func TestLogicalHostPattern(t *testing.T) {
	got, err := LogicalHostPattern("a5be2268", 0, 1, "olares.com")
	if err != nil {
		t.Fatalf("LogicalHostPattern error: %v", err)
	}
	want := "a5be2268.*.olares.com"
	if got != want {
		t.Fatalf("LogicalHostPattern: got %q, want %q", got, want)
	}
	g, err := LogicalHostPattern("A5BE2268", 0, 1, "OLARES.COM.")
	if err != nil {
		t.Fatalf("LogicalHostPattern normalize error: %v", err)
	}
	if g != want {
		t.Fatalf("LogicalHostPattern not normalized: got %q, want %q", g, want)
	}
	if _, err := LogicalHostPattern("", 0, 1, "y"); err == nil {
		t.Fatal("empty appid must return error")
	}
	if _, err := LogicalHostPattern("x", 0, 1, ""); err == nil {
		t.Fatal("empty domain must return error")
	}
}

// Cross-helper invariant: URL host part for viewer "x" is exactly
// "<LogicalHostPattern> with * replaced by x".
func TestURL_LogicalPattern_Consistency(t *testing.T) {
	pat, err := LogicalHostPattern("a5be2268", 0, 1, "olares.com")
	if err != nil {
		t.Fatalf("LogicalHostPattern: %v", err)
	}
	url, err := GenSharedEntranceURLForUser("a5be2268", 0, 1, "alice", "olares.com")
	if err != nil {
		t.Fatalf("GenSharedEntranceURLForUser: %v", err)
	}
	wantHost := strings.Replace(pat, "*", "alice", 1)
	if url != "https://"+wantHost {
		t.Fatalf("URL %q is inconsistent with logical pattern %q", url, pat)
	}
}
