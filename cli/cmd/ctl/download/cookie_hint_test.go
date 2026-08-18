package download

import (
	"strings"
	"testing"
)

const wantImportPrefix = "olares-cli settings integration cookie import --domain "

func TestCookieHostFromURLNormalizes(t *testing.T) {
	cases := []struct {
		raw  string
		want string
	}{
		{"https://WWW.YouTube.com/watch?v=abc", "youtube.com"},
		{"https://m.youtube.com/watch?v=abc", "youtube.com"},
		{"https://music.youtube.com/watch?v=abc", "youtube.com"},
		{"https://www.bilibili.com/video/BV1", "bilibili.com"},
		{"magnet:?xt=urn:btih:abc", ""},
		{"", ""},
		{"not a url", ""},
	}
	for _, tc := range cases {
		if got := cookieHostFromURL(tc.raw); got != tc.want {
			t.Fatalf("cookieHostFromURL(%q) = %q, want %q", tc.raw, got, tc.want)
		}
	}
}

// Cookies live in another command family, so a hint that only says
// "upload a cookie" leaves an agent with nowhere to go. Every cookie
// path has to spell out the whole runnable command.
func TestCookieRequiredHintNamesTheDomain(t *testing.T) {
	got := cookieRequiredHint("https://www.youtube.com/watch?v=abc")
	if !strings.Contains(got, wantImportPrefix+"youtube.com --file cookies.txt") {
		t.Fatalf("hint = %q", got)
	}
}

func TestCookieRequiredHintFallsBackWithoutAHost(t *testing.T) {
	for _, raw := range []string{"", "magnet:?xt=urn:btih:abc", "not a url"} {
		got := cookieRequiredHint(raw)
		if !strings.Contains(got, wantImportPrefix+"<domain>") {
			t.Fatalf("hint for %q = %q", raw, got)
		}
	}
}

func TestInspectCookieHintCoversCodesAndCategories(t *testing.T) {
	cases := []struct {
		name string
		data InspectData
		want bool
	}{
		{"501 cookie required", InspectData{ErrorCode: 501}, true},
		{"private resource", InspectData{ErrorCode: 507}, true},
		{"authorization failed", InspectData{ErrorCategory: "authorization_failed"}, true},
		{"bot detected", InspectData{ErrorCategory: "BOT_DETECTED"}, true},
		{"deleted is not fixable by cookies", InspectData{ErrorCode: 508, ErrorCategory: "deleted"}, false},
		{"healthy probe", InspectData{Provider: "yt-dlp"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := inspectCookieHint(tc.data, "https://youtube.com/watch?v=abc")
			if (got != "") != tc.want {
				t.Fatalf("hint = %q, want present=%v", got, tc.want)
			}
		})
	}
}

func TestTaskCookieHintUsesTheTaskURL(t *testing.T) {
	got := taskCookieHint(DownloadTask{
		URL:         "https://www.bilibili.com/video/BV1",
		ErrCategory: "authorization_failed",
	})
	if !strings.Contains(got, wantImportPrefix+"bilibili.com") {
		t.Fatalf("hint = %q", got)
	}
	if other := taskCookieHint(DownloadTask{ErrCategory: "network_error"}); other != "" {
		t.Fatalf("network failure must not blame cookies: %q", other)
	}
}

func TestTaskErrorRecoveryHintsCookieImportOnCreate(t *testing.T) {
	req := NewDownloadReq{URL: "https://www.youtube.com/watch?v=abc"}

	got := taskErrorRecovery("POST", "/api/download", 501, "cookies required", "", req)
	if !strings.Contains(got, wantImportPrefix+"youtube.com") {
		t.Fatalf("501 status recovery = %q", got)
	}
	got = taskErrorRecovery("POST", "/api/download", 501, "cookies required", "", &req)
	if !strings.Contains(got, wantImportPrefix+"youtube.com") {
		t.Fatalf("501 status recovery with pointer body = %q", got)
	}
}
