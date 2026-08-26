package download

import (
	"context"
	"errors"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestTitleToFileName(t *testing.T) {
	cases := []struct {
		name  string
		title string
		want  string
	}{
		{name: "plain title kept verbatim", title: "Black Myth: Zhong Kui Trailer", want: "Black Myth: Zhong Kui Trailer"},
		{name: "surrounding space trimmed", title: "  clip  ", want: "clip"},
		// A slash in a title made the daemon mkdir the head of the
		// title and nest the file inside it.
		{name: "path separators become underscores", title: `Foo / Bar \ Baz`, want: "Foo _ Bar _ Baz"},
		{name: "nul dropped", title: "a\x00b", want: "ab"},
		{name: "empty stays empty", title: "   ", want: ""},
		{name: "traversal segment refused", title: "..", want: ""},
		{name: "single dot refused", title: ".", want: ""},
		{name: "dots inside a name are fine", title: "Episode 1..mp4", want: "Episode 1..mp4"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := titleToFileName(tc.title); got != tc.want {
				t.Fatalf("titleToFileName(%q) = %q, want %q", tc.title, got, tc.want)
			}
		})
	}
}

// TestTitleToFileNameFitsDaemonBudget pins the truncation contract: an
// over-long name makes the daemon's open() fail ENAMETOOLONG, and a CJK
// title hits the byte cap long before the character one.
func TestTitleToFileNameFitsDaemonBudget(t *testing.T) {
	cases := []struct {
		name  string
		title string
	}{
		{name: "ascii", title: strings.Repeat("a", 300)},
		{name: "cjk", title: strings.Repeat("中", 300)},
		{name: "emoji", title: strings.Repeat("😀", 300)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := titleToFileName(tc.title)
			if utf8.RuneCountInString(got) > fileNameMaxChars {
				t.Fatalf("%d chars exceeds the %d-char budget", utf8.RuneCountInString(got), fileNameMaxChars)
			}
			if len(got) > fileNameMaxBytes {
				t.Fatalf("%d bytes exceeds the %d-byte budget", len(got), fileNameMaxBytes)
			}
			if got == "" {
				t.Fatal("truncation must keep a usable stem")
			}
		})
	}
}

// TestTitleToFileNameKeepsExtension pins that truncation eats the stem,
// not the suffix — a clipped ".mp4" would leave the file unopenable by
// extension and defeat the whole point of naming it.
func TestTitleToFileNameKeepsExtension(t *testing.T) {
	got := titleToFileName(strings.Repeat("中", 200) + ".mp4")
	if !strings.HasSuffix(got, ".mp4") {
		t.Fatalf("truncated name %q lost its extension", got)
	}
	if len(got) > fileNameMaxBytes {
		t.Fatalf("%d bytes exceeds the %d-byte budget", len(got), fileNameMaxBytes)
	}
}

func TestDisplayName(t *testing.T) {
	cases := []struct {
		name string
		task DownloadTask
		want string
	}{
		{
			name: "file_name wins",
			task: DownloadTask{FileName: "clip.mp4", URL: "https://host/x"},
			want: "clip.mp4",
		},
		{
			name: "folder marker stripped",
			task: DownloadTask{FileName: "owner/repo/"},
			want: "owner/repo",
		},
		{
			name: "torrent metadata name once aria2 parsed it",
			task: DownloadTask{
				URL:   "magnet:?xt=urn:btih:abc123def4567890",
				Extra: map[string]interface{}{"torrent_meta": map[string]interface{}{"name": "ubuntu.iso"}},
			},
			want: "ubuntu.iso",
		},
		{
			name: "magnet display name",
			task: DownloadTask{URL: "magnet:?xt=urn:btih:abc&dn=Ubuntu%2024.04"},
			want: "Ubuntu 24.04",
		},
		{
			name: "magnet info-hash when there is no dn",
			task: DownloadTask{URL: "magnet:?xt=urn:btih:0123456789abcdef0123"},
			want: "magnet:0123456789ab…",
		},
		{
			name: "magnet with neither",
			task: DownloadTask{URL: "magnet:?tr=udp://tracker"},
			want: "magnet:…",
		},
		{
			name: "url basename when it looks like a file",
			task: DownloadTask{URL: "https://host/dir/movie.mkv?sig=abc"},
			want: "movie.mkv",
		},
		{
			name: "percent-encoded basename decoded",
			task: DownloadTask{URL: "https://host/my%20movie.mkv"},
			want: "my movie.mkv",
		},
		// The regression this whole change is about: /watch is a
		// routing verb, and showing the URL read as a filename the
		// server never wrote.
		{
			name: "youtube watch url is not a name",
			task: DownloadTask{URL: "https://www.youtube.com/watch?v=oi2QgPH61JM"},
			want: "-",
		},
		{
			name: "bilibili video path is not a name",
			task: DownloadTask{URL: "https://www.bilibili.com/video/BV1xx411c7mD"},
			want: "-",
		},
		{
			name: "no url at all",
			task: DownloadTask{},
			want: "-",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := displayName(tc.task); got != tc.want {
				t.Fatalf("displayName(%+v) = %q, want %q", tc.task, got, tc.want)
			}
		})
	}
}

func TestIsHuggingFaceURL(t *testing.T) {
	for _, raw := range []string{
		"https://huggingface.co/org/repo",
		"https://HuggingFace.co/org/repo/tree/main",
		"https://hf.co/org/repo",
		"hf://datasets/squad",
		"huggingface:org/repo",
	} {
		if !isHuggingFaceURL(raw) {
			t.Fatalf("%q should route to HuggingFace", raw)
		}
	}
	for _, raw := range []string{
		"",
		"https://example.com/huggingface.co/org/repo",
		"https://www.youtube.com/watch?v=abc",
		"magnet:?xt=urn:btih:abc",
	} {
		if isHuggingFaceURL(raw) {
			t.Fatalf("%q should not route to HuggingFace", raw)
		}
	}
}

func TestRenameLockedProvider(t *testing.T) {
	cases := []struct {
		name        string
		rawURL      string
		torrentFile string
		extra       map[string]string
		want        string
	}{
		{name: "plain url is renameable", rawURL: "https://host/x.mp4"},
		{name: "youtube is renameable", rawURL: "https://www.youtube.com/watch?v=abc"},
		{name: "torrent upload", rawURL: "", torrentFile: "./x.torrent", want: "torrent upload"},
		{name: "magnet", rawURL: "magnet:?xt=urn:btih:abc", want: "magnet link"},
		{name: "huggingface url", rawURL: "https://huggingface.co/org/repo", want: "HuggingFace repo"},
		{
			name:   "hf dest in extra",
			rawURL: "https://host/x",
			extra:  map[string]string{"_hf_dest": "cache"},
			want:   "HuggingFace repo",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := renameLockedProvider(tc.rawURL, tc.torrentFile, tc.extra); got != tc.want {
				t.Fatalf("renameLockedProvider = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestInspectedFileName(t *testing.T) {
	t.Run("title becomes the file name", func(t *testing.T) {
		d := &fakeDoer{resp: []byte(`{"code":200,"data":{"provider":"yt-dlp","title":"Foo / Bar"}}`)}
		got := inspectedFileName(context.Background(), &preparedClient{doer: d}, nameProbe{url: "https://www.youtube.com/watch?v=abc"})
		if got != "Foo _ Bar" {
			t.Fatalf("file name = %q, want %q", got, "Foo _ Bar")
		}
		if !strings.HasPrefix(d.lastPath, "/api/url/inspect?") || !strings.Contains(d.lastPath, "url=") {
			t.Fatalf("unexpected probe path %q", d.lastPath)
		}
	})

	// A probe that fails (a 505 for want of a cookie, a channel URL that
	// outruns the server deadline) must not stop a create that would
	// otherwise succeed: the provider writes the real name back anyway.
	t.Run("probe failure omits the field", func(t *testing.T) {
		d := &fakeDoer{err: errors.New("daemon returned code 505 (network)")}
		if got := inspectedFileName(context.Background(), &preparedClient{doer: d}, nameProbe{url: "https://host/x"}); got != "" {
			t.Fatalf("file name = %q, want empty", got)
		}
	})

	t.Run("no title omits the field", func(t *testing.T) {
		d := &fakeDoer{resp: []byte(`{"code":200,"data":{"provider":"aria2"}}`)}
		if got := inspectedFileName(context.Background(), &preparedClient{doer: d}, nameProbe{url: "https://host/x.zip"}); got != "" {
			t.Fatalf("file name = %q, want empty", got)
		}
	})

	t.Run("no url means no probe", func(t *testing.T) {
		d := &fakeDoer{resp: []byte(`{"code":200,"data":{"title":"x"}}`)}
		if got := inspectedFileName(context.Background(), &preparedClient{doer: d}, nameProbe{url: "  "}); got != "" {
			t.Fatalf("file name = %q, want empty", got)
		}
		if d.lastPath != "" {
			t.Fatalf("a blank URL must not be probed, got %q", d.lastPath)
		}
	})
}
