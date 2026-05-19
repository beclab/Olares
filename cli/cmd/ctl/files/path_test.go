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
			// Regression: previously HasTrailingSlash() reported false here
			// while String() rendered "drive/Home/" — an inconsistency that
			// misled callers branching on directory-vs-file. The bare
			// `<fileType>/<extend>` form has no valid non-directory reading.
			name:         "drive Home root without trailing slash",
			input:        "drive/Home",
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

// TestFrontendPathIsExternalNodeRoot pins the volume-listing-layer
// detection used by mkdir / cp / mv / upload to fast-fail writes
// against `external/<node>/`. Reads (`ls`, `cat`, `rm`, `rename`,
// `share`) work at this layer, so the predicate is intentionally
// narrow: only `external/<node>/` (with or without the trailing
// slash, both render SubPath="/") returns true. Any subpath beyond
// `<node>` — even one we'll likely never validate against a real
// volume name like `<volume>` — is outside the predicate's scope
// and is left to the server.
func TestFrontendPathIsExternalNodeRoot(t *testing.T) {
	cases := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "external node root with trailing slash",
			path: "external/node-1/",
			want: true,
		},
		{
			// Bare `<fileType>/<extend>` synthesizes SubPath="/"
			// (see ParseFrontendPath docstring), so it must report
			// the same as the trailing-slash form.
			name: "external node root without trailing slash",
			path: "external/node-1",
			want: true,
		},
		{
			name: "external one segment past node (volume root)",
			path: "external/node-1/hdd1/",
			want: false,
		},
		{
			name: "external nested subdir",
			path: "external/node-1/hdd1/Movies/",
			want: false,
		},
		{
			// cache/<node>/ is a real per-node directory (the
			// volume listing layer is external-only); the
			// predicate must NOT trip on it.
			name: "cache node root is not external",
			path: "cache/node-1/",
			want: false,
		},
		{
			name: "drive Home root is not external",
			path: "drive/Home/",
			want: false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fp, err := ParseFrontendPath(c.path)
			if err != nil {
				t.Fatalf("ParseFrontendPath(%q): %v", c.path, err)
			}
			if got := fp.IsExternalNodeRoot(); got != c.want {
				t.Errorf("IsExternalNodeRoot() = %v, want %v", got, c.want)
			}
		})
	}
}

// TestFrontendPathIsCacheNodeRoot pins the share-side fast-fail
// predicate that mirrors the LarePass web app's `/Files/Cache/`
// node-picker UX: the LarePass app renders that view via
// formatAppDataNode (apps/.../api/files/v2/cache/data.ts L33-47) and
// synthesizes children from the Olares cluster's node list instead
// of hitting /api/resources/cache/, so a "share this row" operation
// at that level points at a node selector, not a real dataset. The
// CLI's share-create path uses this predicate to refuse such shares
// up front — see `frontendPathToShareTarget` in share.go.
//
// Scope is intentionally narrow: only `cache/<node>/` (with or
// without trailing slash, both render SubPath="/") trips the
// predicate. Anything past the node — even a single segment like
// `cache/<node>/<app>` — falls through, because once the user picks
// a node the wire goes back to the regular /api/resources/cache/
// listing and shares are meaningful. Other namespaces (drive,
// external, sync, ...) are out of scope by construction.
func TestFrontendPathIsCacheNodeRoot(t *testing.T) {
	cases := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "cache node root with trailing slash",
			path: "cache/node-1/",
			want: true,
		},
		{
			// Bare <fileType>/<extend> synthesizes SubPath="/", so
			// the predicate must give the same answer.
			name: "cache node root without trailing slash",
			path: "cache/node-1",
			want: true,
		},
		{
			name: "cache one segment past node",
			path: "cache/node-1/app1/",
			want: false,
		},
		{
			name: "cache nested subpath",
			path: "cache/node-1/app1/data/cache.bin",
			want: false,
		},
		{
			// external/<node>/ has its own predicate
			// (IsExternalNodeRoot); this one must NOT claim it.
			name: "external node root is not cache",
			path: "external/node-1/",
			want: false,
		},
		{
			name: "drive Home root is not cache",
			path: "drive/Home/",
			want: false,
		},
		{
			name: "sync repo root is not cache",
			path: "sync/abc-repo/",
			want: false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fp, err := ParseFrontendPath(c.path)
			if err != nil {
				t.Fatalf("ParseFrontendPath(%q): %v", c.path, err)
			}
			if got := fp.IsCacheNodeRoot(); got != c.want {
				t.Errorf("IsCacheNodeRoot() = %v, want %v", got, c.want)
			}
		})
	}
}

// TestFrontendPathIsProtectedDriveHomeChild pins the
// LarePass-aligned policy that the system-managed first-level
// children directly under drive/Home/ refuse rename / delete / move
// (the same shape the web app's `disableMenuItem` array enforces in
// apps/packages/app/src/stores/operation.ts when the user is sitting
// at /Files/Home/).
//
// The predicate is intentionally narrow: only EXACT first-level
// matches under drive/Home/ count, and only against the
// case-sensitive names in ProtectedDriveHomeChildren. Anything else
// — the volume root itself, deeper paths, drive/Data/<same-name>,
// other namespaces — falls through, letting the user keep using
// `mv` / `rm` / `rename` against arbitrary user content unaffected.
func TestFrontendPathIsProtectedDriveHomeChild(t *testing.T) {
	cases := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "drive Home Pictures protected",
			path: "drive/Home/Pictures",
			want: true,
		},
		{
			// Trailing slash is the directory marker used by ls /
			// rm / mv etc. — must not change the policy outcome.
			name: "drive Home Pictures slash protected",
			path: "drive/Home/Pictures/",
			want: true,
		},
		{
			name: "drive Home Documents protected",
			path: "drive/Home/Documents",
			want: true,
		},
		{
			name: "drive Home Movies protected",
			path: "drive/Home/Movies",
			want: true,
		},
		{
			name: "drive Home Music protected",
			path: "drive/Home/Music/",
			want: true,
		},
		{
			name: "drive Home Downloads protected",
			path: "drive/Home/Downloads",
			want: true,
		},
		{
			name: "drive Home Code protected",
			path: "drive/Home/Code",
			want: true,
		},
		{
			name: "drive Home Cache protected",
			path: "drive/Home/Cache",
			want: true,
		},
		{
			name: "drive Home Data protected",
			path: "drive/Home/Data",
			want: true,
		},
		{
			// Defensive entry mirrored from the web app's array —
			// guards against the historical Home/Home/ nested shape.
			name: "drive Home Home protected",
			path: "drive/Home/Home/",
			want: true,
		},
		{
			name: "drive Home Ollama protected",
			path: "drive/Home/Ollama",
			want: true,
		},
		{
			name: "drive Home Huggingface protected (one-word casing)",
			path: "drive/Home/Huggingface",
			want: true,
		},
		{
			// Deeper subpath under a protected name: user content,
			// not the protected entry itself — must NOT trip.
			name: "drive Home Pictures with deeper subpath unaffected",
			path: "drive/Home/Pictures/Trip2024/",
			want: false,
		},
		{
			name: "drive Home Documents nested file unaffected",
			path: "drive/Home/Documents/notes.md",
			want: false,
		},
		{
			// Volume root itself — already rejected by rename / rm
			// / cp's own root-refusal, but the predicate must not
			// claim ownership of that error message.
			name: "drive Home root not protected",
			path: "drive/Home/",
			want: false,
		},
		{
			// Case sensitivity: GUI compares the enum string
			// values exactly, so lowercase variants must NOT
			// match. (They also wouldn't exist as real dirs since
			// these names are system-bootstrapped with fixed
			// casing.)
			name: "drive Home pictures lowercase not protected",
			path: "drive/Home/pictures",
			want: false,
		},
		{
			name: "drive Home HuggingFace mixed case not protected",
			path: "drive/Home/HuggingFace",
			want: false,
		},
		{
			// Same name under a different drive extend: Data
			// already has its own root and isn't a Home child, so
			// the policy does not apply.
			name: "drive Data Pictures not protected",
			path: "drive/Data/Pictures",
			want: false,
		},
		{
			// Other fileTypes are out of scope by construction.
			name: "sync repo Pictures not protected",
			path: "sync/abc-repo/Pictures",
			want: false,
		},
		{
			name: "external node Pictures not protected",
			path: "external/node-1/hdd1/Pictures",
			want: false,
		},
		{
			// Unrelated drive/Home child (user-created): must NOT
			// match — the predicate is for SYSTEM-managed names
			// only, not "any first-level entry".
			name: "drive Home user folder unaffected",
			path: "drive/Home/MyProjects/",
			want: false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fp, err := ParseFrontendPath(c.path)
			if err != nil {
				t.Fatalf("ParseFrontendPath(%q): %v", c.path, err)
			}
			if got := fp.IsProtectedDriveHomeChild(); got != c.want {
				t.Errorf("IsProtectedDriveHomeChild() = %v, want %v", got, c.want)
			}
		})
	}
}

// TestProtectedDriveHomeChildrenList pins the alphabetical, stable
// rendering used in error messages — both that the list contains the
// expected names and that the joiner is exactly ", " so callers can
// rely on the format for downstream parsing or i18n.
func TestProtectedDriveHomeChildrenList(t *testing.T) {
	got := ProtectedDriveHomeChildrenList()
	want := "Cache, Code, Data, Documents, Downloads, Home, Huggingface, Movies, Music, Ollama, Pictures"
	if !strings.Contains(got, "Pictures") || !strings.Contains(got, "Huggingface") {
		t.Fatalf("ProtectedDriveHomeChildrenList() missing core entries: %q", got)
	}
	if got != want {
		t.Errorf("ProtectedDriveHomeChildrenList() = %q, want %q", got, want)
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
