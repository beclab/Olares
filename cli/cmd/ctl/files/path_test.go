package files

import (
	"strings"
	"testing"
)

func TestParseFrontendPath(t *testing.T) {
	cases := []struct {
		name           string
		input          string
		wantFileType   string
		wantExtend     string
		wantSubPath    string
		wantTrailing   bool
		wantString     string
		wantErrSubstr  string
	}{
		{
			name:         "drive Home root with trailing slash",
			input:        "drive/Home/",
			wantFileType: "drive",
			wantExtend:   "Home",
			wantSubPath:  "/",
			wantTrailing: true,
			wantString:   "drive/Home/",
		},
		{
			name:         "drive Home subdir",
			input:        "drive/Home/Documents",
			wantFileType: "drive",
			wantExtend:   "Home",
			wantSubPath:  "/Documents",
			wantString:   "drive/Home/Documents",
		},
		{
			name:         "drive Home subdir with trailing slash preserved",
			input:        "drive/Home/Documents/",
			wantFileType: "drive",
			wantExtend:   "Home",
			wantSubPath:  "/Documents/",
			wantTrailing: true,
			wantString:   "drive/Home/Documents/",
		},
		{
			name:         "drive Data root",
			input:        "drive/Data/",
			wantFileType: "drive",
			wantExtend:   "Data",
			wantSubPath:  "/",
			wantTrailing: true,
			wantString:   "drive/Data/",
		},
		{
			name:         "sync repo",
			input:        "sync/abc-123-repo/sub/dir",
			wantFileType: "sync",
			wantExtend:   "abc-123-repo",
			wantSubPath:  "/sub/dir",
			wantString:   "sync/abc-123-repo/sub/dir",
		},
		{
			name:         "awss3 nested",
			input:        "awss3/myaccount/bucket/key.txt",
			wantFileType: "awss3",
			wantExtend:   "myaccount",
			wantSubPath:  "/bucket/key.txt",
			wantString:   "awss3/myaccount/bucket/key.txt",
		},
		{
			name:         "leading slash tolerated",
			input:        "/drive/Home/",
			wantFileType: "drive",
			wantExtend:   "Home",
			wantSubPath:  "/",
			wantTrailing: true,
			wantString:   "drive/Home/",
		},
		{
			name:         "double slashes collapsed",
			input:        "drive/Home//Documents///nested",
			wantFileType: "drive",
			wantExtend:   "Home",
			wantSubPath:  "/Documents/nested",
			wantString:   "drive/Home/Documents/nested",
		},
		{
			name:          "empty",
			input:         "",
			wantErrSubstr: "is empty",
		},
		{
			name:          "only slashes",
			input:         "///",
			wantErrSubstr: "empty after trimming",
		},
		{
			name:          "single segment",
			input:         "drive",
			wantErrSubstr: "must have <fileType>/<extend>",
		},
		{
			name:          "single segment with trailing slash",
			input:         "drive/",
			wantErrSubstr: "must have <fileType>/<extend>",
		},
		{
			name:          "unknown fileType",
			input:         "foo/bar/",
			wantErrSubstr: "unknown fileType",
		},
		{
			name:          "drive bad extend",
			input:         "drive/Other/",
			wantErrSubstr: "drive extend must be Home or Data",
		},
		{
			name:          "uppercase fileType rejected",
			input:         "Drive/Home/",
			wantErrSubstr: "unknown fileType",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ParseFrontendPath(c.input)
			if c.wantErrSubstr != "" {
				if err == nil {
					t.Fatalf("ParseFrontendPath(%q): want error containing %q, got nil (parsed=%+v)", c.input, c.wantErrSubstr, got)
				}
				if !strings.Contains(err.Error(), c.wantErrSubstr) {
					t.Fatalf("ParseFrontendPath(%q): want error containing %q, got %q", c.input, c.wantErrSubstr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseFrontendPath(%q): unexpected error: %v", c.input, err)
			}
			if got.FileType != c.wantFileType {
				t.Errorf("FileType = %q, want %q", got.FileType, c.wantFileType)
			}
			if got.Extend != c.wantExtend {
				t.Errorf("Extend = %q, want %q", got.Extend, c.wantExtend)
			}
			if got.SubPath != c.wantSubPath {
				t.Errorf("SubPath = %q, want %q", got.SubPath, c.wantSubPath)
			}
			if got.HasTrailingSlash() != c.wantTrailing {
				t.Errorf("HasTrailingSlash() = %v, want %v", got.HasTrailingSlash(), c.wantTrailing)
			}
			if s := got.String(); s != c.wantString {
				t.Errorf("String() = %q, want %q", s, c.wantString)
			}
		})
	}
}

func TestFrontendPathURLPath(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no special chars",
			input: "drive/Home/Documents",
			want:  "drive/Home/Documents",
		},
		{
			name:  "trailing slash preserved",
			input: "drive/Home/Documents/",
			want:  "drive/Home/Documents/",
		},
		{
			name:  "extend root",
			input: "drive/Home/",
			want:  "drive/Home/",
		},
		{
			name:  "filename with space",
			input: "drive/Home/My Documents/notes.md",
			want:  "drive/Home/My%20Documents/notes.md",
		},
		{
			name:  "filename with hash and question mark",
			input: "drive/Home/a#b?c.txt",
			want:  "drive/Home/a%23b%3Fc.txt",
		},
		{
			name:  "filename with plus and percent",
			input: "drive/Home/100%/x+y.txt",
			want:  "drive/Home/100%25/x%2By.txt",
		},
		{
			name:  "parens and space like duplicate filename",
			input: "drive/Home/Documents/report (1).txt",
			want:  "drive/Home/Documents/report%20(1).txt",
		},
		{
			name:  "non-ASCII filename",
			input: "drive/Home/笔记/分享.md",
			want:  "drive/Home/%E7%AC%94%E8%AE%B0/%E5%88%86%E4%BA%AB.md",
		},
		{
			name:  "slashes still act as separators",
			input: "drive/Home/a/b/c",
			want:  "drive/Home/a/b/c",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fp, err := ParseFrontendPath(c.input)
			if err != nil {
				t.Fatalf("ParseFrontendPath(%q): unexpected error: %v", c.input, err)
			}
			if got := fp.URLPath(); got != c.want {
				t.Errorf("URLPath() = %q, want %q", got, c.want)
			}
		})
	}
}
