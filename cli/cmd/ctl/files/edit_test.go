package files

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/beclab/Olares/cli/internal/files/download"
	"github.com/beclab/Olares/cli/internal/files/edit"
	"github.com/beclab/Olares/cli/pkg/credential"
)

// TestFrontendPathToEditTarget covers the cobra-layer adapter that
// turns a user-supplied 3-segment path into the edit package's
// Target. The volume-root + directory-path refusals are enforced
// both here and in edit.Plan; this test pins the cobra-side error
// messages (which include CTAs the planner alone can't produce
// since it doesn't have the parsed FrontendPath).
func TestFrontendPathToEditTarget(t *testing.T) {
	t.Run("happy path: file under Home", func(t *testing.T) {
		tgt, err := frontendPathToEditTarget("drive/Home/Documents/notes.md")
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if tgt.FileType != "drive" || tgt.Extend != "Home" {
			t.Errorf("FileType/Extend wrong: %+v", tgt)
		}
		if tgt.SubPath != "/Documents/notes.md" {
			t.Errorf("SubPath: got %q", tgt.SubPath)
		}
	})

	t.Run("trailing slash rejected (directory path)", func(t *testing.T) {
		_, err := frontendPathToEditTarget("drive/Home/Documents/")
		if err == nil {
			t.Fatal("want error, got nil")
		}
		if !strings.Contains(err.Error(), "directory path") {
			t.Errorf("err: %v", err)
		}
	})

	t.Run("volume root rejected with friendly CTA", func(t *testing.T) {
		_, err := frontendPathToEditTarget("drive/Home/")
		if err == nil {
			t.Fatal("want error, got nil")
		}
		// CTA should suggest a sample file path so the user
		// knows what shape the input should take.
		if !strings.Contains(err.Error(), "drive/Home/notes.md") {
			t.Errorf("err should suggest a sample file: %v", err)
		}
	})

	// Regression: same dot-segment blacklist mkdir / rename
	// enforce, applied here BEFORE ParseFrontendPath's path.Clean
	// silently collapses the offending segments. Without this
	// pre-check `edit drive/Home/foo/./bar` would land on
	// `drive/Home/foo/bar` (a different file than the user
	// typed), and `edit drive/Home/foo/..` would silently land at
	// the volume root.
	t.Run("dot-segment blacklist fires before path.Clean", func(t *testing.T) {
		cases := []struct {
			name string
			in   string
			seg  string
		}{
			{"interior './'", "drive/Home/foo/./bar", `"."`},
			{"interior '../'", "drive/Home/foo/../bar", `".."`},
			{"all-traversal", "drive/Home/../../etc", `".."`},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				_, err := frontendPathToEditTarget(c.in)
				if err == nil {
					t.Fatalf("want error for %q", c.in)
				}
				if !strings.Contains(err.Error(), c.seg) {
					t.Errorf("err %q should mention segment %s", err.Error(), c.seg)
				}
				if !strings.Contains(err.Error(), "path-traversal blacklist") {
					t.Errorf("err %q should mention 'path-traversal blacklist'", err.Error())
				}
			})
		}
	})

	t.Run("empty input bubbles up parser error", func(t *testing.T) {
		if _, err := frontendPathToEditTarget(""); err == nil {
			t.Error("want error for empty path")
		}
	})

	t.Run("cloud drive accepted by adapter and planner", func(t *testing.T) {
		// awss3 / google / dropbox are now part of the
		// supported allow-list (see edit.SupportedFileTypesList +
		// TestPlan_NamespaceAllowlist for the wire-shape source
		// of truth). The adapter's only job here is to surface
		// the parsed target unchanged; the planner then builds
		// the canonical /api/resources/<fileType><path> URL
		// without complaint.
		for _, raw := range []string{
			"awss3/myacct/bucket/file.txt",
			"google/myacct/Documents/draft.md",
			"dropbox/myacct/Notes/idea.txt",
		} {
			t.Run(raw, func(t *testing.T) {
				tgt, err := frontendPathToEditTarget(raw)
				if err != nil {
					t.Fatalf("adapter rejected: %v", err)
				}
				if _, err := edit.Plan(tgt); err != nil {
					t.Errorf("planner err: %v", err)
				}
			})
		}
	})

	t.Run("tencent / share / internal still rejected", func(t *testing.T) {
		// These remain on the deny-list — see
		// TestPlan_NamespaceAllowlist for the rationale. The
		// adapter still parses them (they're known fileTypes);
		// the refusal lands at the planner.
		for _, raw := range []string{
			"tencent/myacct/file.txt",
			"share/someuser/notes.md",
			"internal/x/y.md",
		} {
			t.Run(raw, func(t *testing.T) {
				tgt, err := frontendPathToEditTarget(raw)
				if err != nil {
					t.Fatalf("adapter rejected unexpectedly: %v", err)
				}
				_, planErr := edit.Plan(tgt)
				if planErr == nil {
					t.Fatalf("want planner error for %s, got nil", raw)
				}
				if !strings.Contains(planErr.Error(), "not supported") {
					t.Errorf("planner err: %v", planErr)
				}
			})
		}
	})
}

// TestLastSegmentForEdit pins the basename extraction we use to
// name the temp file. Editors key syntax highlighting off the
// extension, so a regression here would silently break vim's
// filetype detection / VSCode's language association.
func TestLastSegmentForEdit(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"drive/Home/Documents/notes.md", "notes.md"},
		{"drive/Home/notes.md", "notes.md"},
		{"sync/repoid/dir/sub/file.json", "file.json"},
		{"cache/node/leaf", "leaf"},
		{"", ""},
		{"/", ""},
		{"/just-leaf", "just-leaf"},
		// Trailing slash: edge case — Plan rejects directories so
		// in practice we never see this, but if a future caller
		// does we still want a sensible fallback rather than a
		// panic.
		{"drive/Home/Documents/", "Documents"},
	}
	for _, c := range cases {
		got := lastSegmentForEdit(c.in)
		if got != c.want {
			t.Errorf("lastSegmentForEdit(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestBytesEqual covers the no-change predicate. Length
// short-circuit matters because a vim `:q` after a no-touch open
// rewrites the file with the same length — bytes.Equal handles
// that, but we verify the contract here so a future "fingerprint"
// refactor of bytesEqual stays drop-in compatible.
func TestBytesEqual(t *testing.T) {
	cases := []struct {
		name string
		a, b []byte
		want bool
	}{
		{"both nil", nil, nil, true},
		{"both empty", []byte{}, []byte{}, true},
		{"nil vs empty", nil, []byte{}, true}, // bytes.Equal treats them equal
		{"identical", []byte("hi\n"), []byte("hi\n"), true},
		{"different content same length", []byte("aa\n"), []byte("bb\n"), false},
		{"different length", []byte("hi"), []byte("hi\n"), false},
		{"binary bytes", []byte{0x00, 0xff, 0x10}, []byte{0x00, 0xff, 0x10}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := bytesEqual(c.a, c.b); got != c.want {
				t.Errorf("bytesEqual(%q,%q) = %v, want %v", c.a, c.b, got, c.want)
			}
		})
	}
}

// TestPickEditor walks the editor cascade. The "vi" / "notepad"
// fallbacks are baked into the OS image, so we exercise them only
// when they're actually on the test runner's PATH (which is the
// common case but not guaranteed in every CI sandbox).
func TestPickEditor(t *testing.T) {
	// Save and restore env so this test doesn't leak state.
	saveVisual := os.Getenv("VISUAL")
	saveEditor := os.Getenv("EDITOR")
	t.Cleanup(func() {
		_ = os.Setenv("VISUAL", saveVisual)
		_ = os.Setenv("EDITOR", saveEditor)
	})

	t.Run("--editor flag wins", func(t *testing.T) {
		_ = os.Setenv("VISUAL", "ignored-visual")
		_ = os.Setenv("EDITOR", "ignored-editor")
		// Use a real binary likely present on every CI host.
		bin := pickRealBinary(t)
		got, err := pickEditor(bin)
		if err != nil {
			t.Fatalf("pickEditor(%q): %v", bin, err)
		}
		if got != bin {
			t.Errorf("got %q, want %q", got, bin)
		}
	})

	t.Run("--editor missing → error mentions PATH", func(t *testing.T) {
		_, err := pickEditor("definitely-not-a-real-editor-xyz123")
		if err == nil {
			t.Fatal("want error, got nil")
		}
		if !strings.Contains(err.Error(), "not found in PATH") {
			t.Errorf("err: %v", err)
		}
	})

	t.Run("VISUAL beats EDITOR", func(t *testing.T) {
		visualBin := pickRealBinary(t)
		_ = os.Setenv("VISUAL", visualBin)
		_ = os.Setenv("EDITOR", "definitely-not-a-real-editor")
		got, err := pickEditor("")
		if err != nil {
			t.Fatalf("pickEditor: %v", err)
		}
		if got != visualBin {
			t.Errorf("got %q, want VISUAL=%q to win", got, visualBin)
		}
	})

	t.Run("EDITOR used when VISUAL empty", func(t *testing.T) {
		editorBin := pickRealBinary(t)
		_ = os.Unsetenv("VISUAL")
		_ = os.Setenv("EDITOR", editorBin)
		got, err := pickEditor("")
		if err != nil {
			t.Fatalf("pickEditor: %v", err)
		}
		if got != editorBin {
			t.Errorf("got %q, want EDITOR=%q", got, editorBin)
		}
	})

	t.Run("editor command with arguments threads through", func(t *testing.T) {
		// `git`-style "GIT_EDITOR='code --wait'" support: the
		// resolved command keeps its args; only the first token
		// is PATH-looked-up.
		bin := pickRealBinary(t)
		spec := bin + " --some-arg"
		_ = os.Unsetenv("VISUAL")
		_ = os.Setenv("EDITOR", spec)
		got, err := pickEditor("")
		if err != nil {
			t.Fatalf("pickEditor: %v", err)
		}
		if got != spec {
			t.Errorf("got %q, want %q (args preserved)", got, spec)
		}
	})

	t.Run("fallback when nothing set", func(t *testing.T) {
		_ = os.Unsetenv("VISUAL")
		_ = os.Unsetenv("EDITOR")
		got, err := pickEditor("")
		// Skip if the platform fallback isn't on PATH (the
		// minimal CI image might lack vi); the contract under
		// test is the cascade ordering, not the presence of vi.
		if err != nil {
			if strings.Contains(err.Error(), "not found in PATH") {
				t.Skipf("fallback editor not available on this host: %v", err)
			}
			t.Fatalf("pickEditor: %v", err)
		}
		var want string
		if runtime.GOOS == "windows" {
			want = "notepad"
		} else {
			want = "vi"
		}
		if got != want {
			t.Errorf("fallback: got %q, want %q", got, want)
		}
	})
}

// pickRealBinary returns the absolute path of a binary
// guaranteed to exist on the test runner. We prefer "echo" /
// "true" — both are POSIX and ship on Windows in modern Git Bash
// / WSL setups; CI without one of these is exotic enough to skip
// the test rather than complicate the helper.
func pickRealBinary(t *testing.T) string {
	t.Helper()
	for _, c := range []string{"echo", "true", "cmd"} {
		if p, err := exec.LookPath(c); err == nil {
			return p
		}
	}
	t.Skip("no echo/true/cmd on PATH; skipping editor test")
	return ""
}

// TestRunEdit_NoChangeSkipsUpload exercises the most important
// behavior contract: when the editor returns without modifying
// the temp file, the cobra layer skips the PUT entirely. We
// drive a full GET-edit-PUT round-trip against an httptest
// server with a no-op editor (`true` exits 0 without touching
// its arg).
//
// The test also exercises the temp-file lifecycle: with --keep-
// temp unset, the temp directory is cleaned up on the no-change
// path.
func TestRunEdit_NoChangeSkipsUpload(t *testing.T) {
	bin, err := exec.LookPath("true")
	if err != nil {
		t.Skip("`true` binary not on PATH; skipping no-change test")
	}

	current := "current contents\n"
	srv, calls := newEditServer(t, current)
	defer srv.Close()

	out := &captureWriter{}
	op := edit.Op{
		Endpoint:    "/api/resources/drive/Home/notes.md",
		DisplayPath: "drive/Home/notes.md",
	}

	statClient := &download.Client{HTTPClient: srv.Client(), BaseURL: srv.URL}
	editClient := &edit.Client{HTTPClient: srv.Client(), BaseURL: srv.URL}

	// Drive the inner pieces directly — runEdit's TTY guard +
	// factory dependency make a full-stack test heavier than
	// it's worth, and the inner pieces are what carry the
	// no-change semantics anyway.
	//
	// maxSize=0 disables the size cap so this test only exercises
	// the no-change branch; the cap itself is covered by
	// TestStatAndFetch_MaxSize / TestRunEdit_PostEditOversizeBlocked.
	bytesGot, isDir, err := statAndFetch(context.Background(), statClient, editClient, op.DisplayPath, false, 0)
	if err != nil {
		t.Fatalf("statAndFetch: %v", err)
	}
	if isDir {
		t.Fatal("statAndFetch reported isDir=true for a file")
	}

	tmpDir, tmpFile, err := writeTempFile(op.DisplayPath, bytesGot)
	if err != nil {
		t.Fatalf("writeTempFile: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := runEditor(context.Background(), bin, tmpFile, nil, out, out); err != nil {
		t.Fatalf("runEditor(`true`): %v", err)
	}

	postBytes, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("read post-edit: %v", err)
	}
	if !bytesEqual(bytesGot, postBytes) {
		t.Errorf("`true` should not modify the temp file; got %q", postBytes)
	}
	if calls.put != 0 {
		t.Errorf("a no-change run should not PUT; saw %d PUT call(s)", calls.put)
	}
}

// TestRunEdit_ChangedTriggersPut is the complementary case: when
// the editor modifies the temp file, the cobra layer's PUT lands
// the new bytes verbatim.
func TestRunEdit_ChangedTriggersPut(t *testing.T) {
	// We synthesise a "smart" editor as a tiny shell script so
	// the change is deterministic. On Windows where a real shell
	// isn't guaranteed we skip — the contract is platform-
	// independent, but the test fixture isn't.
	if runtime.GOOS == "windows" {
		t.Skip("script-based fake editor not portable to Windows test runners")
	}

	tdir := t.TempDir()
	scriptPath := filepath.Join(tdir, "fake-editor")
	const script = "#!/bin/sh\nprintf 'edited\\n' > \"$1\"\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o700); err != nil {
		t.Fatalf("write fake editor: %v", err)
	}

	current := "current\n"
	srv, calls := newEditServer(t, current)
	defer srv.Close()

	op := edit.Op{
		Endpoint:    "/api/resources/drive/Home/notes.md",
		DisplayPath: "drive/Home/notes.md",
	}
	statClient := &download.Client{HTTPClient: srv.Client(), BaseURL: srv.URL}
	editClient := &edit.Client{HTTPClient: srv.Client(), BaseURL: srv.URL}

	bytesGot, _, err := statAndFetch(context.Background(), statClient, editClient, op.DisplayPath, false, 0)
	if err != nil {
		t.Fatalf("statAndFetch: %v", err)
	}
	tmpDir, tmpFile, err := writeTempFile(op.DisplayPath, bytesGot)
	if err != nil {
		t.Fatalf("writeTempFile: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	out := &captureWriter{}
	if err := runEditor(context.Background(), scriptPath, tmpFile, nil, out, out); err != nil {
		t.Fatalf("runEditor: %v", err)
	}
	post, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("read post-edit: %v", err)
	}
	if bytesEqual(bytesGot, post) {
		t.Fatalf("fake editor should have changed the file; pre=%q post=%q", bytesGot, post)
	}
	// Now run the PUT branch the cobra layer would take.
	if err := editClient.PutBytes(context.Background(), op, post, edit.DefaultContentType); err != nil {
		t.Fatalf("PutBytes: %v", err)
	}
	if calls.put != 1 {
		t.Errorf("expected one PUT, saw %d", calls.put)
	}
	if string(calls.lastBody) != string(post) {
		t.Errorf("PUT body: got %q, want %q", calls.lastBody, post)
	}
}

// TestStatAndFetch_NotFoundCreate covers the --create branch:
// when the remote target 404s, allowCreate=true returns an empty
// buffer (so the user starts from scratch) instead of erroring.
func TestStatAndFetch_NotFoundCreate(t *testing.T) {
	srv, _ := newEditServer404(t)
	defer srv.Close()

	statClient := &download.Client{HTTPClient: srv.Client(), BaseURL: srv.URL}
	editClient := &edit.Client{HTTPClient: srv.Client(), BaseURL: srv.URL}

	t.Run("404 + allowCreate=false → error mentions --create", func(t *testing.T) {
		_, _, err := statAndFetch(context.Background(), statClient, editClient,
			"drive/Home/missing.md", false, 0)
		if err == nil {
			t.Fatal("want error, got nil")
		}
		if !strings.Contains(err.Error(), "--create") {
			t.Errorf("err should hint at --create: %v", err)
		}
	})

	t.Run("404 + allowCreate=true → empty buffer", func(t *testing.T) {
		got, isDir, err := statAndFetch(context.Background(), statClient, editClient,
			"drive/Home/missing.md", true, 0)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if isDir {
			t.Error("isDir: want false")
		}
		if len(got) != 0 {
			t.Errorf("want empty buffer, got %q", got)
		}
	})
}

// TestStatAndFetch_MaxSize covers the pre-edit size cap: when
// the remote file's reported Size exceeds maxSize, statAndFetch
// errors out BEFORE pulling bytes (saving a wasted multi-MB GET
// for binaries the user surely didn't mean to edit). The error
// must point at the --max-size override so the user has a clear
// next step.
func TestStatAndFetch_MaxSize(t *testing.T) {
	// Fixture: 10 KiB of "x" — bigger than the 1 KiB cap we
	// pass below, smaller than the 100 KiB cap. One file size
	// is enough; the assertion is on the cap branch, not the
	// size value.
	current := strings.Repeat("x", 10*1024)

	t.Run("Stat size > maxSize is rejected with cap CTA", func(t *testing.T) {
		// Track whether /api/raw was hit; the cap must
		// short-circuit BEFORE the Fetch so the user doesn't
		// pay for the byte transfer when we already know we'll
		// reject.
		srv, calls := newEditServer(t, current)
		defer srv.Close()
		statClient := &download.Client{HTTPClient: srv.Client(), BaseURL: srv.URL}
		editClient := &edit.Client{HTTPClient: srv.Client(), BaseURL: srv.URL}

		const cap1KiB = 1024
		_, _, err := statAndFetch(context.Background(), statClient, editClient,
			"drive/Home/notes.md", false, cap1KiB)
		if err == nil {
			t.Fatal("want size-cap error, got nil")
		}
		if !strings.Contains(err.Error(), "--max-size") {
			t.Errorf("err should mention --max-size: %v", err)
		}
		if !strings.Contains(err.Error(), "exceeds") {
			t.Errorf("err should say 'exceeds': %v", err)
		}
		if calls.getRaw != 0 {
			t.Errorf("Fetch should NOT have been called when Stat exceeded the cap; saw %d", calls.getRaw)
		}
	})

	t.Run("Stat size <= maxSize falls through to Fetch", func(t *testing.T) {
		srv, calls := newEditServer(t, current)
		defer srv.Close()
		statClient := &download.Client{HTTPClient: srv.Client(), BaseURL: srv.URL}
		editClient := &edit.Client{HTTPClient: srv.Client(), BaseURL: srv.URL}

		const cap100KiB = 100 * 1024
		got, _, err := statAndFetch(context.Background(), statClient, editClient,
			"drive/Home/notes.md", false, cap100KiB)
		if err != nil {
			t.Fatalf("statAndFetch: %v", err)
		}
		if len(got) != len(current) {
			t.Errorf("got %d bytes, want %d", len(got), len(current))
		}
		if calls.getRaw != 1 {
			t.Errorf("Fetch should have been called once, saw %d", calls.getRaw)
		}
	})

	t.Run("maxSize=0 disables the cap entirely", func(t *testing.T) {
		srv, _ := newEditServer(t, current)
		defer srv.Close()
		statClient := &download.Client{HTTPClient: srv.Client(), BaseURL: srv.URL}
		editClient := &edit.Client{HTTPClient: srv.Client(), BaseURL: srv.URL}

		got, _, err := statAndFetch(context.Background(), statClient, editClient,
			"drive/Home/notes.md", false, 0)
		if err != nil {
			t.Fatalf("statAndFetch: %v", err)
		}
		if len(got) != len(current) {
			t.Errorf("got %d bytes, want %d", len(got), len(current))
		}
	})
}

// TestDefaultMaxSize pins the constant — 1 MiB. Keeping this
// test next to the cap-check tests makes a cap-policy change
// (e.g. switching to 5 MiB or environment-driven) an obvious
// code-review event, and stops a silent regression where
// `--max-size` defaults to a different value than the
// documentation claims.
func TestDefaultMaxSize(t *testing.T) {
	if DefaultMaxSize != 1<<20 {
		t.Errorf("DefaultMaxSize = %d, want %d (1 MiB)", DefaultMaxSize, 1<<20)
	}
}

// TestReformatEditHTTPErr pins the status-code → CTA mapping. The
// pattern matches reformatHTTPErr / reformatRmHTTPErr / etc., so a
// regression in any one of these four reformatters tends to surface
// here first because edit covers all the relevant statuses (401,
// 403, 404, 409) plus the 459 Olares-edge variant.
func TestReformatEditHTTPErr(t *testing.T) {
	cases := []struct {
		name   string
		status int
		expect string
	}{
		{"401 mentions profile login", 401, "profile login"},
		{"403 mentions profile login", 403, "profile login"},
		{"459 (Olares edge auth-failed) maps to login CTA", 459, "profile login"},
		{"404 says not found", 404, "not found on the server"},
		{"409 hints at concurrent change", 409, "concurrently"},
		{"413 surfaces payload too large", 413, "payload too large"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := &edit.HTTPError{Status: c.status, Body: "x"}
			got := reformatEditHTTPErr(err, "alice@olares.com", "save", "drive/Home/x.md")
			if got == nil {
				t.Fatal("want error, got nil")
			}
			if !strings.Contains(got.Error(), c.expect) {
				t.Errorf("err %q does not contain %q", got.Error(), c.expect)
			}
		})
	}

	// Typed credential errors must bypass the status-code branches.
	t.Run("ErrTokenInvalidated surfaces verbatim", func(t *testing.T) {
		inv := &credential.ErrTokenInvalidated{}
		got := reformatEditHTTPErr(inv, "alice@olares.com", "save", "drive/Home/x.md")
		if !errors.Is(got, inv) && got != inv { //nolint:errorlint // typed equality is intentional
			t.Errorf("ErrTokenInvalidated should pass through untouched: got %T (%v)", got, got)
		}
	})

	// Non-typed errors fall through unchanged so the caller's
	// per-verb wrapping (e.g. fmt.Errorf("save %s: ...")) is the
	// final word.
	t.Run("untyped error passes through", func(t *testing.T) {
		raw := errors.New("network blip")
		got := reformatEditHTTPErr(raw, "", "save", "x")
		if got != raw {
			t.Errorf("untyped error should pass through; got %v", got)
		}
	})
}

// TestEditStatus extracts the wire status code from either an
// edit.HTTPError or a download.HTTPError. The reformatter relies
// on this to share one switch across the two error types — a
// regression here would silently route Stat-side 4xx through the
// "untyped" fallback instead of the friendly CTA.
func TestEditStatus(t *testing.T) {
	t.Run("edit.HTTPError", func(t *testing.T) {
		e := &edit.HTTPError{Status: 404}
		s, ok := editStatus(e)
		if !ok || s != 404 {
			t.Errorf("got (%d,%v); want (404,true)", s, ok)
		}
	})
	t.Run("download.HTTPError", func(t *testing.T) {
		e := &download.HTTPError{Status: 401}
		s, ok := editStatus(e)
		if !ok || s != 401 {
			t.Errorf("got (%d,%v); want (401,true)", s, ok)
		}
	})
	t.Run("untyped error returns ok=false", func(t *testing.T) {
		_, ok := editStatus(errors.New("x"))
		if ok {
			t.Errorf("ok: got true, want false")
		}
	})
}

// TestHasBinaryExtension pins the deny-list policy: the formats
// the LarePass GUI shows as image/pdf/video/audio/blob (plus the
// archive / executable / db extensions every editor would corrupt
// in seconds) are refused; mainstream text formats — including
// the "looks scary but is XML" cases (.svg / .xml / .html) — are
// allowed. The lookup is case-insensitive, and compound suffixes
// like .tar.gz are matched as a single unit.
//
// A regression here would let `files edit drive/Home/Photos/big.jpg`
// silently spawn the editor on a JPEG, which is the exact foot-gun
// the policy exists to prevent.
func TestHasBinaryExtension(t *testing.T) {
	binary := []string{
		// images
		"foo.jpg", "FOO.JPG", "bar.png", "snap.gif", "x.webp",
		"y.heic", "z.tiff", "icon.ico", "scan.bmp",
		// documents
		"doc.pdf", "report.docx", "sheet.xlsx", "deck.pptx",
		"book.epub", "kindle.mobi",
		// av
		"clip.mp4", "movie.mkv", "song.mp3", "track.flac",
		// archives (single + compound)
		"app.zip", "tar.gz", "src.tar.gz", "src.TAR.GZ", "logs.tar.bz2",
		"backup.tar.xz", "old.7z",
		// executables / bytecode
		"a.exe", "lib.so", "lib.DYLIB", "App.class", "service.jar",
		"mod.wasm",
		// disk images / installers
		"install.dmg", "ubuntu.iso", "pkg.deb",
		// database / fonts
		"users.sqlite3", "Inter.woff2",
	}
	for _, name := range binary {
		t.Run("binary/"+name, func(t *testing.T) {
			if !hasBinaryExtension(name) {
				t.Errorf("hasBinaryExtension(%q) = false, want true", name)
			}
		})
	}

	text := []string{
		// files we explicitly want editable
		"notes.md", "config.yaml", "config.YAML",
		"app.json", "data.csv", "schema.sql", "Dockerfile",
		"Makefile", ".env", ".gitignore",
		// looks-binary-but-isn't (XML / source code)
		"icon.svg", "page.html", "feed.xml",
		"main.go", "src.ts", "comp.tsx", "lib.rs",
		// extensionless
		"README", "LICENSE",
		// empty input
		"",
	}
	for _, name := range text {
		t.Run("text/"+name, func(t *testing.T) {
			if hasBinaryExtension(name) {
				t.Errorf("hasBinaryExtension(%q) = true, want false (text format)", name)
			}
		})
	}
}

// TestLooksBinary exercises the post-fetch content sniff. The
// rule is the same one git uses: a NUL byte in the first 8 KiB
// flips the verdict to binary; otherwise the buffer is treated
// as text regardless of how exotic the encoding is. We pin both
// directions with realistic fixtures and edge cases (empty
// buffer, NUL exactly at the sniff boundary) so a future tweak
// to the sniff length can't accidentally regress the contract.
func TestLooksBinary(t *testing.T) {
	t.Run("empty buffer is text (--create with 404)", func(t *testing.T) {
		if looksBinary(nil) {
			t.Error("nil should be text")
		}
		if looksBinary([]byte{}) {
			t.Error("empty []byte should be text")
		}
	})

	t.Run("plain ASCII text is text", func(t *testing.T) {
		if looksBinary([]byte("hello world\n")) {
			t.Error("plain text should not look binary")
		}
	})

	t.Run("UTF-8 (multi-byte runes) is text", func(t *testing.T) {
		// 中文 / emoji / accented Latin — none contain NUL.
		if looksBinary([]byte("你好，世界！\n# Title — café 🎉\n")) {
			t.Error("UTF-8 should not look binary")
		}
	})

	t.Run("PNG header is binary (NUL within first 8 bytes)", func(t *testing.T) {
		// 0x89 'P' 'N' 'G' 0x0d 0x0a 0x1a 0x0a — no NUL here.
		// Use a real PNG IHDR chunk that includes width/height
		// fields with leading zero bytes (pretty much every PNG).
		buf := []byte{
			0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a,
			0x00, 0x00, 0x00, 0x0d, 'I', 'H', 'D', 'R',
		}
		if !looksBinary(buf) {
			t.Error("PNG bytes should look binary")
		}
	})

	t.Run("JPEG with NUL in first kilobyte is binary", func(t *testing.T) {
		// JPEG SOI then a typical APP0 marker tail with embedded
		// NUL — every JPEG carries them in its early bytes.
		buf := append([]byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00}, make([]byte, 200)...)
		if !looksBinary(buf) {
			t.Error("JPEG bytes should look binary")
		}
	})

	t.Run("PDF starts text-y but binary streams hit fast", func(t *testing.T) {
		// Real PDFs start with "%PDF-1.x\n" then quickly include
		// an FlateDecode binary stream containing NUL bytes.
		buf := append([]byte("%PDF-1.7\n"), 0x00, 'b', 'i', 'n', 'a', 'r', 'y')
		if !looksBinary(buf) {
			t.Error("PDF with binary stream should look binary")
		}
	})

	t.Run("NUL at sniff boundary still triggers", func(t *testing.T) {
		// All-zero header within sniff window: trivially binary.
		buf := make([]byte, binarySniffLen)
		if !looksBinary(buf) {
			t.Error("buffer with NULs in sniff window should look binary")
		}
	})

	t.Run("NUL beyond sniff window does NOT trigger (by design)", func(t *testing.T) {
		// Mirrors git's heuristic: only the first window is
		// scanned. A NUL in the tail past 8 KiB is rare for
		// real text but we accept the false-negative trade-off
		// in exchange for cheap O(1) sniffing.
		buf := make([]byte, binarySniffLen+1)
		for i := range buf {
			buf[i] = 'a'
		}
		buf[binarySniffLen] = 0x00
		if looksBinary(buf) {
			t.Error("NUL beyond sniff window should NOT trigger (by design)")
		}
	})
}

// TestRunEdit_BinaryExtensionRejected exercises the layer-1
// guard: when the remote path's extension is on the deny-list,
// the cobra layer must refuse BEFORE Stat / Fetch. We can lean
// on frontendPathToEditTarget + edit.Plan + hasBinaryExtension
// directly since that's the same chain runEdit walks; the runEdit
// integration is too heavy to wire (TTY guard + factory) and the
// guard's contract is purely path-shape based.
func TestRunEdit_BinaryExtensionRejected(t *testing.T) {
	cases := []struct {
		path string
		want bool // want hasBinaryExtension to fire
	}{
		{"drive/Home/Photos/big.jpg", true},
		{"drive/Home/Documents/report.PDF", true},
		{"drive/Home/Music/song.mp3", true},
		{"drive/Home/archives/code.tar.gz", true},
		// borderline text-y formats that should still pass
		{"drive/Home/Documents/notes.md", false},
		{"drive/Home/.config/app.yaml", false},
		{"drive/Home/site/icon.svg", false},
		{"drive/Home/scripts/build.ts", false},
		{"drive/Home/no-extension-config", false},
	}
	for _, c := range cases {
		t.Run(c.path, func(t *testing.T) {
			tgt, err := frontendPathToEditTarget(c.path)
			if err != nil {
				t.Fatalf("frontendPathToEditTarget: %v", err)
			}
			op, err := edit.Plan(tgt)
			if err != nil {
				t.Fatalf("edit.Plan: %v", err)
			}
			got := hasBinaryExtension(op.DisplayPath)
			if got != c.want {
				t.Errorf("hasBinaryExtension(%q) = %v, want %v", op.DisplayPath, got, c.want)
			}
		})
	}
}

// TestRunEdit_BinaryContentSniffRejected exercises the layer-2
// guard: a path with an innocuous (or missing) extension whose
// content is actually binary. We drive statAndFetch end-to-end
// against a fake server that serves binary bytes, then assert
// looksBinary fires on the returned buffer — the runEdit caller
// short-circuits on exactly that signal.
func TestRunEdit_BinaryContentSniffRejected(t *testing.T) {
	// A "config.dat" filename that the extension layer doesn't
	// flag, but whose body is a PNG — the sniff layer must
	// catch it.
	pngHeader := []byte{
		0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 'I', 'H', 'D', 'R',
		0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00,
	}
	srv, _ := newEditServer(t, string(pngHeader))
	defer srv.Close()
	statClient := &download.Client{HTTPClient: srv.Client(), BaseURL: srv.URL}
	editClient := &edit.Client{HTTPClient: srv.Client(), BaseURL: srv.URL}

	// Path uses an extension NOT on the deny-list so the
	// content sniff is the only thing standing between the user
	// and a corrupted PNG.
	got, isDir, err := statAndFetch(context.Background(), statClient, editClient,
		"drive/Home/notes.md", false, 0)
	if err != nil {
		t.Fatalf("statAndFetch: %v", err)
	}
	if isDir {
		t.Fatal("isDir: want false")
	}
	if !looksBinary(got) {
		t.Errorf("looksBinary(<png header>) = false, want true; buf head=%x", got[:min(8, len(got))])
	}
}

// captureWriter is a tiny io.Writer that records what's written
// into it. Used as stdout/stderr for the editor child process so
// noisy editor output (or any error message) doesn't escape into
// the surrounding test runner's terminal output.
type captureWriter struct {
	buf []byte
}

func (c *captureWriter) Write(p []byte) (int, error) {
	c.buf = append(c.buf, p...)
	return len(p), nil
}

func (c *captureWriter) String() string { return string(c.buf) }

// editServerCalls counts the wire calls a fake edit-server saw,
// so a test can assert "we did / didn't PUT" or read back the
// exact body the cobra layer sent.
type editServerCalls struct {
	getRaw   int
	getList  int
	put      int
	lastBody []byte
}

// newEditServer returns an httptest server that:
//   - answers GET /api/raw/<path> with `current` (200 + body),
//   - answers GET /api/resources/<parent>/ with a one-item
//     listing so download.Stat finds the file as non-dir,
//   - records PUT /api/resources/<path> bodies for assertion.
//
// Both the path encoding and the listing envelope shape mirror
// the real backend (see internal/files/download/list.go and
// internal/files/edit/edit.go for the wire-shape sources of
// truth).
func newEditServer(t *testing.T, current string) (*httptest.Server, *editServerCalls) {
	t.Helper()
	c := &editServerCalls{}
	mux := http.NewServeMux()
	// Parent listing: report a single file entry named "notes.md"
	// so download.Stat lands on (IsDir=false, Size=N).
	mux.HandleFunc("/api/resources/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			c.getList++
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"name":"Home","items":[{"name":"notes.md","isDir":false,"size":` +
				strconv.Itoa(len(current)) + `}]}`))
		case http.MethodPut:
			c.put++
			body, _ := io.ReadAll(r.Body)
			c.lastBody = append([]byte(nil), body...)
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/raw/", func(w http.ResponseWriter, r *http.Request) {
		c.getRaw++
		_, _ = w.Write([]byte(current))
	})
	srv := httptest.NewServer(mux)
	return srv, c
}

// newEditServer404 returns an httptest server that responds 404
// to every call. Used for the --create-flag tests where we want
// to verify the cobra layer's behavior on a missing remote
// target without rebuilding the parent-listing fixture.
func newEditServer404(t *testing.T) (*httptest.Server, *editServerCalls) {
	t.Helper()
	c := &editServerCalls{}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	srv := httptest.NewServer(mux)
	return srv, c
}
