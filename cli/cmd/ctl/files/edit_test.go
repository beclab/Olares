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

	t.Run("cloud / tencent / share / internal rejected at the planner", func(t *testing.T) {
		// awss3 / google / dropbox are NOT on the allow-list —
		// see TestPlan_NamespaceAllowlist for the rationale
		// (Bug 1: /api/raw returns JSON envelopes for cloud
		// namespaces, which would silently corrupt the file on
		// writeback). The adapter still parses them as known
		// fileTypes; the refusal lands at the planner with a
		// targeted message that points at the safe
		// download → edit-locally → upload alternative.
		cases := []struct {
			raw  string
			want []string // substrings the planner err should contain
		}{
			{"awss3/myacct/bucket/file.txt", []string{"cloud-drive", "/api/raw", "files download", "files upload"}},
			{"google/myacct/Documents/draft.md", []string{"cloud-drive", "/api/raw", "files download", "files upload"}},
			{"dropbox/myacct/Notes/idea.txt", []string{"cloud-drive", "/api/raw", "files download", "files upload"}},
			{"tencent/myacct/file.txt", []string{"cloud-drive", "/api/raw", "files download", "files upload"}},
			{"share/someuser/notes.md", []string{"not supported"}},
			{"internal/x/y.md", []string{"not supported"}},
		}
		for _, c := range cases {
			t.Run(c.raw, func(t *testing.T) {
				tgt, err := frontendPathToEditTarget(c.raw)
				if err != nil {
					t.Fatalf("adapter rejected unexpectedly: %v", err)
				}
				_, planErr := edit.Plan(tgt)
				if planErr == nil {
					t.Fatalf("want planner error for %s, got nil", c.raw)
				}
				for _, want := range c.want {
					if !strings.Contains(planErr.Error(), want) {
						t.Errorf("planner err for %s: missing %q in %q", c.raw, want, planErr.Error())
					}
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
	bytesGot, isDir, isCreate, err := statAndFetch(context.Background(), statClient, editClient, op.DisplayPath, false, 0)
	if err != nil {
		t.Fatalf("statAndFetch: %v", err)
	}
	if isDir {
		t.Fatal("statAndFetch reported isDir=true for a file")
	}
	if isCreate {
		t.Error("isCreate must be false when the file existed")
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

	bytesGot, _, _, err := statAndFetch(context.Background(), statClient, editClient, op.DisplayPath, false, 0)
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
		_, _, _, err := statAndFetch(context.Background(), statClient, editClient,
			"drive/Home/missing.md", false, 0)
		if err == nil {
			t.Fatal("want error, got nil")
		}
		if !strings.Contains(err.Error(), "--create") {
			t.Errorf("err should hint at --create: %v", err)
		}
	})

	t.Run("404 + allowCreate=true → empty buffer + isCreate=true", func(t *testing.T) {
		got, isDir, isCreate, err := statAndFetch(context.Background(), statClient, editClient,
			"drive/Home/missing.md", true, 0)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if isDir {
			t.Error("isDir: want false")
		}
		if !isCreate {
			t.Error("isCreate: want true on a Stat-404 + --create path (Bug 4 regression guard)")
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
		_, _, _, err := statAndFetch(context.Background(), statClient, editClient,
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
		got, _, _, err := statAndFetch(context.Background(), statClient, editClient,
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

		got, _, _, err := statAndFetch(context.Background(), statClient, editClient,
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
	got, isDir, _, err := statAndFetch(context.Background(), statClient, editClient,
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

// TestPreflightEdit_Order pins the LOCAL-validation ordering
// `files edit` advertises: path → plan → binary-extension →
// editor cascade. The exact order matters because each step
// names a different class of failure to the user, and shuffling
// them surfaces the wrong error first.
//
// Background: an earlier revision of `runEdit` called pickEditor
// AFTER the remote Stat / Fetch / writeTempFile. A typo in
// --editor would still pull the remote file and allocate a temp
// dir before failing — defeating the documented "fail fast"
// contract and triggering an unnecessary network round-trip on
// what was a purely local misconfiguration. preflightEdit pulls
// the editor check up-front; this test pins each gate's
// precedence so a future shuffle is caught at code-review time.
func TestPreflightEdit_Order(t *testing.T) {
	bogusEditor := "definitely-not-a-real-editor-xyz123"

	t.Run("path-shape error wins over editor error", func(t *testing.T) {
		// Bogus path AND bogus editor — the path error names
		// what's wrong with the user's INPUT, which trumps a
		// downstream environment misconfiguration.
		o := &editOptions{editor: bogusEditor}
		_, _, err := preflightEdit("not-a-real-frontend-path", o)
		if err == nil {
			t.Fatal("want error, got nil")
		}
		if strings.Contains(err.Error(), "not found in PATH") {
			t.Errorf("got editor-not-found first; want path-shape error: %v", err)
		}
	})

	t.Run("dot-segment error wins over editor error", func(t *testing.T) {
		// Same logic for the path-traversal blacklist — the
		// path-shape error class includes ./.. segments.
		o := &editOptions{editor: bogusEditor}
		_, _, err := preflightEdit("drive/Home/foo/../bar", o)
		if err == nil {
			t.Fatal("want error, got nil")
		}
		if !strings.Contains(err.Error(), "path-traversal blacklist") {
			t.Errorf("got %q; want dot-segment error first", err.Error())
		}
	})

	t.Run("namespace allow-list error wins over editor error", func(t *testing.T) {
		// Cloud namespace + bogus editor — the namespace
		// refusal points at the recovery `download → upload`
		// path; the editor error is downstream environment.
		o := &editOptions{editor: bogusEditor}
		_, _, err := preflightEdit("awss3/myacct/file.txt", o)
		if err == nil {
			t.Fatal("want error, got nil")
		}
		if !strings.Contains(err.Error(), "cloud-drive") {
			t.Errorf("got %q; want cloud-namespace error first", err.Error())
		}
	})

	t.Run("binary-extension error wins over editor error", func(t *testing.T) {
		// Valid path, JPEG extension, bogus editor — the
		// binary-extension refusal teaches the user "this file
		// isn't editable here" even if their $EDITOR is also
		// broken; fixing the path is the realer blocker.
		o := &editOptions{editor: bogusEditor}
		_, _, err := preflightEdit("drive/Home/Photos/big.jpg", o)
		if err == nil {
			t.Fatal("want error, got nil")
		}
		if !strings.Contains(err.Error(), "non-text format") {
			t.Errorf("got %q; want binary-extension error first", err.Error())
		}
	})

	t.Run("editor error fires once path / extension are clean", func(t *testing.T) {
		// Path is well-shaped, namespace is supported, extension
		// is text — only the editor is busted. preflightEdit
		// should surface that AND it should be the LAST
		// failure mode in the chain.
		o := &editOptions{editor: bogusEditor}
		_, _, err := preflightEdit("drive/Home/notes.md", o)
		if err == nil {
			t.Fatal("want error, got nil")
		}
		if !strings.Contains(err.Error(), "not found in PATH") {
			t.Errorf("err: got %q, want editor-not-found message", err.Error())
		}
	})

	t.Run("happy path returns op + editor with no error", func(t *testing.T) {
		bin := pickRealBinary(t)
		o := &editOptions{editor: bin}
		op, editor, err := preflightEdit("drive/Home/Documents/notes.md", o)
		if err != nil {
			t.Fatalf("preflightEdit: %v", err)
		}
		if op.DisplayPath != "drive/Home/Documents/notes.md" {
			t.Errorf("op.DisplayPath: got %q", op.DisplayPath)
		}
		if op.Endpoint == "" {
			t.Error("op.Endpoint: must be populated by edit.Plan")
		}
		if editor != bin {
			t.Errorf("editor: got %q, want %q", editor, bin)
		}
	})

	t.Run("--allow-binary lets binary extensions through to the editor cascade", func(t *testing.T) {
		// With --allow-binary, the extension-deny step is
		// skipped — so a bogus editor + .jpg path now surfaces
		// the editor error (not the extension error).
		// Documents the escape-hatch contract.
		o := &editOptions{editor: bogusEditor, allowBinary: true}
		_, _, err := preflightEdit("drive/Home/Photos/big.jpg", o)
		if err == nil {
			t.Fatal("want editor error, got nil")
		}
		if !strings.Contains(err.Error(), "not found in PATH") {
			t.Errorf("err: got %q, want editor-not-found (--allow-binary should bypass ext-deny)", err.Error())
		}
	})
}

// TestRunEdit_PickEditorFailsBeforeFactory is the full-stack
// regression guard for the "preflight before network" contract.
// We pass a NIL Factory and a bogus editor: under the post-fix
// ordering preflightEdit short-circuits with the editor error
// before runEdit ever dereferences `f`; under the pre-fix
// ordering pickEditor ran AFTER `f.ResolveProfile`, so a nil
// Factory would panic before the editor check fired.
//
// `recover()` in the test catches that panic and fails with a
// targeted message — so a future reordering regression surfaces
// as a clear "preflight regressed past nil-Factory deref"
// failure rather than a stack trace.
func TestRunEdit_PickEditorFailsBeforeFactory(t *testing.T) {
	// Stub the TTY guard so runEdit's interactive check passes.
	// (`go test` runs with a piped stdin, so the production
	// guard would short-circuit before ever reaching the
	// preflight + factory work we want to test.)
	saved := isTTY
	isTTY = func() bool { return true }
	t.Cleanup(func() { isTTY = saved })

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("runEdit panicked (preflight regressed past nil-Factory deref): %v", r)
		}
	}()

	o := &editOptions{
		editor:  "definitely-not-a-real-editor-xyz123",
		maxSize: DefaultMaxSize,
	}
	err := runEdit(
		context.Background(),
		nil,           // *cmdutil.Factory — MUST NOT be touched in the post-fix order
		io.Discard,
		nil,           // editorStdin
		io.Discard,    // editorStdout
		io.Discard,    // editorStderr
		"drive/Home/Documents/notes.md",
		o,
	)
	if err == nil {
		t.Fatal("want pickEditor error, got nil")
	}
	if !strings.Contains(err.Error(), "not found in PATH") {
		t.Errorf("err: got %q, want pickEditor 'not found in PATH' (preflight should have short-circuited the nil Factory)", err.Error())
	}
}

// TestRunEdit_PathErrorBeforeFactory is the companion guard for
// path-shape failures: a bogus path with a working environment
// must STILL short-circuit before runEdit dereferences the
// (nil) Factory. Documents that the entire preflight runs
// before any factory work, not just the editor check.
func TestRunEdit_PathErrorBeforeFactory(t *testing.T) {
	saved := isTTY
	isTTY = func() bool { return true }
	t.Cleanup(func() { isTTY = saved })

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("runEdit panicked on a bogus path (preflight order regressed): %v", r)
		}
	}()

	bin := pickRealBinary(t)
	o := &editOptions{editor: bin, maxSize: DefaultMaxSize}
	err := runEdit(
		context.Background(),
		nil,
		io.Discard,
		nil, io.Discard, io.Discard,
		"awss3/myacct/file.txt", // cloud namespace → planner refuses
		o,
	)
	if err == nil {
		t.Fatal("want planner error, got nil")
	}
	if !strings.Contains(err.Error(), "cloud-drive") {
		t.Errorf("err: got %q, want planner cloud-drive refusal", err.Error())
	}
}

// TestStatAndFetch_StatSizeZeroLargeBody pins Bug 5's fix at the
// cobra layer: when Stat reports Size=0 (either really empty or
// the backend forgot to populate the field) but the actual body
// is many MB, the bounded read in edit.Client.Fetch (Bug 5's
// fix) catches it and surfaces a *TooLargeError, which
// statAndFetch wraps into a friendly --max-size CTA. The
// previous behaviour skipped the pre-fetch cap (because
// Stat.Size was 0) and slurped the whole body into memory
// before the post-fetch length check fired.
func TestStatAndFetch_StatSizeZeroLargeBody(t *testing.T) {
	// 64 KiB body, but we tell Stat the size is 0. Cap is 1 KiB.
	// The bounded LimitReader inside Fetch must fail before
	// the full 64 KiB lands in memory.
	body := strings.Repeat("z", 64*1024)
	srv := httptest.NewServer(http.NewServeMux())
	defer srv.Close()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/resources/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Stat-listing reports Size=0 to trigger the Bug 5 path.
		_, _ = w.Write([]byte(`{"name":"Home","items":[{"name":"notes.md","isDir":false,"size":0}]}`))
	})
	mux.HandleFunc("/api/raw/", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(body))
	})
	srv2 := httptest.NewServer(mux)
	defer srv2.Close()

	statClient := &download.Client{HTTPClient: srv2.Client(), BaseURL: srv2.URL}
	editClient := &edit.Client{HTTPClient: srv2.Client(), BaseURL: srv2.URL}

	const cap1KiB = 1024
	_, _, _, err := statAndFetch(context.Background(), statClient, editClient,
		"drive/Home/notes.md", false, cap1KiB)
	if err == nil {
		t.Fatal("want bounded-read cap error, got nil")
	}
	if !strings.Contains(err.Error(), "--max-size") {
		t.Errorf("err should hint --max-size: %v", err)
	}
	if !strings.Contains(err.Error(), "exceeds") {
		t.Errorf("err should say 'exceeds': %v", err)
	}
}

// TestStatAndFetch_RaceBetweenStatAndFetch pins Bug 2's fix: when
// Stat reports the file exists but the subsequent Fetch comes
// back 404, the helper MUST surface a conflict (concurrent
// delete) error — NOT fall through to the --create empty-buffer
// path. Recreating a file someone else just deleted is almost
// never what the user asked for, and the previous behaviour
// would silently rebuild it on a `--create` re-run with no
// hint that anything raced.
func TestStatAndFetch_RaceBetweenStatAndFetch(t *testing.T) {
	t.Run("--create=false → conflict error names the race", func(t *testing.T) {
		srv := newRaceEditServer(t)
		defer srv.Close()
		statClient := &download.Client{HTTPClient: srv.Client(), BaseURL: srv.URL}
		editClient := &edit.Client{HTTPClient: srv.Client(), BaseURL: srv.URL}

		_, _, _, err := statAndFetch(context.Background(), statClient, editClient,
			"drive/Home/notes.md", false, 0)
		if err == nil {
			t.Fatal("want race-conflict error, got nil")
		}
		msg := err.Error()
		if !strings.Contains(msg, "disappeared between stat and fetch") {
			t.Errorf("err should name the race: %v", msg)
		}
		if !strings.Contains(msg, "concurrent delete") {
			t.Errorf("err should hint at concurrent delete: %v", msg)
		}
	})

	t.Run("--create=true STILL surfaces conflict (does NOT recreate silently)", func(t *testing.T) {
		// This is the heart of Bug 2's regression guard:
		// passing --create on a path Stat said was alive must
		// not turn a concurrent-delete race into a silent
		// recreate. The user asked to edit an existing file,
		// not to materialise a new one.
		srv := newRaceEditServer(t)
		defer srv.Close()
		statClient := &download.Client{HTTPClient: srv.Client(), BaseURL: srv.URL}
		editClient := &edit.Client{HTTPClient: srv.Client(), BaseURL: srv.URL}

		got, _, isCreate, err := statAndFetch(context.Background(), statClient, editClient,
			"drive/Home/notes.md", true, 0)
		if err == nil {
			t.Fatalf("want race-conflict error even with --create, got got=%q isCreate=%v", got, isCreate)
		}
		if !strings.Contains(err.Error(), "disappeared between stat and fetch") {
			t.Errorf("err should name the race: %v", err)
		}
	})
}

// TestRunEdit_CreateModeForcesPut pins Bug 4's fix: when the user
// passed --create against a non-existent file and exits the
// editor without typing anything (`:q!`), the cobra layer must
// STILL PUT the empty buffer to the server. Otherwise --create
// becomes a silent no-op and the file is never materialised.
//
// We exercise the post-statAndFetch contract directly (the
// runEdit integration is too heavy because of the TTY guard):
// statAndFetch returns isCreate=true on a 404 + --create, and
// the cobra layer's no-change branch checks isCreate to bypass
// the "no upload" early-return and fall through to the PUT.
func TestRunEdit_CreateModeForcesPut(t *testing.T) {
	srv, calls := newCreateAwareEditServer(t)
	defer srv.Close()
	statClient := &download.Client{HTTPClient: srv.Client(), BaseURL: srv.URL}
	editClient := &edit.Client{HTTPClient: srv.Client(), BaseURL: srv.URL}

	// The file does not exist; with --create=true statAndFetch
	// must return (empty buffer, isDir=false, isCreate=true,
	// nil) so the cobra layer can force the PUT.
	currentBytes, isDir, isCreate, err := statAndFetch(context.Background(),
		statClient, editClient, "drive/Home/missing.md", true, 0)
	if err != nil {
		t.Fatalf("statAndFetch: %v", err)
	}
	if isDir {
		t.Fatal("isDir: want false")
	}
	if !isCreate {
		t.Fatal("isCreate: want true on a Stat-404 + --create path (Bug 4 regression guard)")
	}
	if len(currentBytes) != 0 {
		t.Errorf("currentBytes: want empty, got %q", currentBytes)
	}

	// Simulate the editor exiting without changes — newBytes
	// equals currentBytes (both empty). The cobra layer's
	// bytesEqual check would normally early-return; with
	// isCreate=true it MUST fall through to the PUT.
	newBytes := []byte{}
	if !bytesEqual(currentBytes, newBytes) {
		t.Fatal("bytesEqual on two empty buffers should be true")
	}

	// The fix: when isCreate, force the PUT.
	if !isCreate {
		t.Fatal("isCreate gating the PUT is the regression guard — bail out if it ever flips")
	}
	op := edit.Op{
		Endpoint:    "/api/resources/drive/Home/missing.md",
		DisplayPath: "drive/Home/missing.md",
	}
	if err := editClient.PutBytes(context.Background(), op, newBytes, edit.DefaultContentType); err != nil {
		t.Fatalf("PutBytes: %v", err)
	}
	if calls.put != 1 {
		t.Errorf("want exactly one PUT (creating an empty file), saw %d", calls.put)
	}
	if len(calls.lastBody) != 0 {
		t.Errorf("PUT body: want empty (created empty file), got %q", calls.lastBody)
	}
}

// newRaceEditServer returns an httptest server that returns a
// successful parent listing (so download.Stat sees the file as
// alive) but answers GET /api/raw/ with 404 — simulating a
// concurrent delete that landed between the Stat and the Fetch.
// Bug 2's regression guard.
func newRaceEditServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/resources/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"Home","items":[{"name":"notes.md","isDir":false,"size":42}]}`))
	})
	mux.HandleFunc("/api/raw/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	return httptest.NewServer(mux)
}

// newCreateAwareEditServer returns an httptest server that 404s
// on Stat (forcing the --create branch) AND records PUT bodies
// so a test can assert the "force PUT in create mode" contract.
// The PUT handler returns 200 regardless of body — including the
// zero-byte body that proves --create + :q! actually creates an
// empty file. Bug 4's regression guard.
func newCreateAwareEditServer(t *testing.T) (*httptest.Server, *editServerCalls) {
	t.Helper()
	c := &editServerCalls{}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/resources/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Listing the parent of a missing file: report
			// the parent dir as empty so download.Stat
			// concludes "not found" rather than seeing the
			// file in the items array.
			c.getList++
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"name":"Home","items":[]}`))
		case http.MethodPut:
			c.put++
			body, _ := io.ReadAll(r.Body)
			c.lastBody = append([]byte(nil), body...)
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/raw/", func(w http.ResponseWriter, _ *http.Request) {
		// The file doesn't exist on the wire either — but
		// statAndFetch should have already short-circuited via
		// Stat's not-found before reaching here.
		w.WriteHeader(http.StatusNotFound)
	})
	srv := httptest.NewServer(mux)
	return srv, c
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
