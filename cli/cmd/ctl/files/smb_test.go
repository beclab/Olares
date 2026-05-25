package files

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beclab/Olares/cli/internal/files/smbmount"
)

// TestNewSMBCommand_Tree confirms the shape of the `files smb`
// subtree: parent + three direct children, plus the per-history
// list/add/rm grandchildren. Locking the tree down at the cobra
// level catches a regression where a future refactor accidentally
// drops a verb.
func TestNewSMBCommand_Tree(t *testing.T) {
	cmd := NewSMBCommand(nil)
	if cmd.Use != "smb" {
		t.Errorf("Use = %q, want \"smb\"", cmd.Use)
	}
	wantTopLevel := map[string]bool{
		"mount":   false,
		"unmount": false,
		"history": false,
	}
	for _, sub := range cmd.Commands() {
		// `Use` may carry trailing flag annotations (e.g.
		// "mount <smb-url> [flags]"); trim to the verb.
		verb := strings.SplitN(sub.Use, " ", 2)[0]
		if _, ok := wantTopLevel[verb]; ok {
			wantTopLevel[verb] = true
			continue
		}
		t.Errorf("unexpected top-level subcommand %q", verb)
	}
	for verb, seen := range wantTopLevel {
		if !seen {
			t.Errorf("missing top-level subcommand %q", verb)
		}
	}

	// history subtree.
	var history *struct{ found bool }
	for _, sub := range cmd.Commands() {
		verb := strings.SplitN(sub.Use, " ", 2)[0]
		if verb != "history" {
			continue
		}
		history = &struct{ found bool }{found: true}
		wantHistory := map[string]bool{"list": false, "add": false, "rm": false}
		for _, gc := range sub.Commands() {
			gv := strings.SplitN(gc.Use, " ", 2)[0]
			if _, ok := wantHistory[gv]; ok {
				wantHistory[gv] = true
				continue
			}
			t.Errorf("unexpected history subcommand %q", gv)
		}
		for v, s := range wantHistory {
			if !s {
				t.Errorf("missing history subcommand %q", v)
			}
		}
	}
	if history == nil || !history.found {
		t.Fatal("history subtree missing")
	}
}

// TestNewSMBMountCommand_Flags pins the cobra flag plumbing for
// `mount`: --user / --password / --password-stdin / --node / --json.
// A regression that drops --password-stdin (the script-friendly
// secret-passing channel) is the kind of thing this catches.
func TestNewSMBMountCommand_Flags(t *testing.T) {
	cmd := newSMBMountCommand(nil)
	if f := cmd.Flags().Lookup("user"); f == nil || f.Shorthand != "u" {
		t.Errorf("--user/-u flag missing or wrong shorthand: %+v", f)
	}
	if f := cmd.Flags().Lookup("password"); f == nil || f.Shorthand != "p" {
		t.Errorf("--password/-p flag missing or wrong shorthand: %+v", f)
	}
	if f := cmd.Flags().Lookup("password-stdin"); f == nil || f.Value.Type() != "bool" {
		t.Errorf("--password-stdin missing or wrong type: %+v", f)
	}
	if f := cmd.Flags().Lookup("node"); f == nil {
		t.Error("--node flag missing")
	}
	if f := cmd.Flags().Lookup("json"); f == nil || f.Value.Type() != "bool" {
		t.Errorf("--json flag missing or wrong type: %+v", f)
	}
	// --no-history: opt-out from the per-node history autofill the
	// mount command picks up by default. Regression guard: dropping
	// this flag would silently re-enable the old behavior where a
	// stale saved password could keep getting reused.
	if f := cmd.Flags().Lookup("no-history"); f == nil || f.Value.Type() != "bool" {
		t.Errorf("--no-history flag missing or wrong type: %+v", f)
	}
	if !strings.Contains(cmd.Long, "//host.local") {
		t.Error("Long help should include a sample SMB URL")
	}
	if !strings.Contains(cmd.Long, "Saved favorite") {
		t.Error("Long help should describe the saved-favorite autofill behavior")
	}
}

// TestNewSMBHistoryAddCommand_Flags pins the per-verb flag plumbing
// for `history add` — the same secret-passing trio as `mount` plus
// --node.
func TestNewSMBHistoryAddCommand_Flags(t *testing.T) {
	cmd := newSMBHistoryAddCommand(nil)
	for _, name := range []string{"user", "password", "password-stdin", "node"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("--%s flag missing", name)
		}
	}
}

// TestRunSMBMount_RejectsBadInputs covers every CLI-side guard that
// fires before any wire call: malformed URL, conflicting password
// flags. The error messages MUST mention the relevant flag(s) so
// the user can act without re-reading the help.
func TestRunSMBMount_RejectsBadInputs(t *testing.T) {
	cases := []struct {
		name    string
		url     string
		opts    smbMountOptions
		errSubs []string
	}{
		{
			name:    "missing // prefix",
			url:     "host/share",
			opts:    smbMountOptions{user: "u", password: "p"},
			errSubs: []string{"//", "host/share"},
		},
		{
			name:    "password and password-stdin together",
			url:     "//h/s",
			opts:    smbMountOptions{user: "u", password: "p", passwordStdin: true},
			errSubs: []string{"--password", "--password-stdin", "mutually exclusive"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := runSMBMount(nil, nil, nopWriter{}, strings.NewReader(""), c.url, &c.opts)
			if err == nil {
				t.Fatal("expected error")
			}
			for _, s := range c.errSubs {
				if !strings.Contains(err.Error(), s) {
					t.Errorf("error %q missing substring %q", err.Error(), s)
				}
			}
		})
	}
}

// TestResolveSMBPassword covers the four-mode resolver:
//   - explicit --password wins
//   - --password-stdin reads one line, trims trailing CR/LF
//   - historyPwHint is used when -p / --password-stdin are absent
//   - missing all of the above with a non-TTY stdin → error (the
//     interactive branch needs a TTY and is impractical to unit-
//     test cleanly).
//
// The history-hint branch is the autofill payoff: with no flags
// passed, a non-empty hint short-circuits the interactive prompt so
// users with a saved favorite don't have to re-type the password
// they already stored via `files smb history add`.
func TestResolveSMBPassword(t *testing.T) {
	t.Run("explicit password wins over history hint", func(t *testing.T) {
		// -p must beat the history hint even when both are
		// non-empty — explicit flags always take precedence.
		got, err := resolveSMBPassword(&smbMountOptions{password: "x"}, strings.NewReader(""), nopWriter{}, "history-pw")
		if err != nil {
			t.Fatal(err)
		}
		if got != "x" {
			t.Errorf("got %q", got)
		}
	})
	t.Run("password-stdin wins over history hint", func(t *testing.T) {
		got, err := resolveSMBPassword(&smbMountOptions{passwordStdin: true}, strings.NewReader("s3cret\n"), nopWriter{}, "history-pw")
		if err != nil {
			t.Fatal(err)
		}
		if got != "s3cret" {
			t.Errorf("got %q", got)
		}
	})
	t.Run("password-stdin without trailing newline", func(t *testing.T) {
		got, err := resolveSMBPassword(&smbMountOptions{passwordStdin: true}, strings.NewReader("s3cret"), nopWriter{}, "")
		if err != nil {
			t.Fatal(err)
		}
		if got != "s3cret" {
			t.Errorf("got %q", got)
		}
	})
	t.Run("password-stdin empty input", func(t *testing.T) {
		_, err := resolveSMBPassword(&smbMountOptions{passwordStdin: true}, strings.NewReader(""), nopWriter{}, "")
		if err == nil {
			t.Fatal("expected error for empty stdin password")
		}
		if !strings.Contains(err.Error(), "empty") {
			t.Errorf("error = %v", err)
		}
	})
	t.Run("history hint used when no flag supplied", func(t *testing.T) {
		// No -p, no --password-stdin → the hint must short-circuit
		// the TTY-check / prompt branch. This is the whole point
		// of the autofill feature; if this test fails, mounting
		// against a saved favorite has regressed to "prompt me
		// every time" behavior.
		got, err := resolveSMBPassword(&smbMountOptions{}, strings.NewReader(""), nopWriter{}, "saved-pw")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "saved-pw" {
			t.Errorf("got %q, want history hint", got)
		}
	})
}

// TestRunSMBUnmount_RejectsBadName guards the cobra-layer pre-flight
// for `unmount`: empty name, slashes in the name, etc. — these
// would otherwise produce a 404 with no useful diagnostic.
func TestRunSMBUnmount_RejectsBadName(t *testing.T) {
	cases := []struct {
		name    string
		entry   string
		errSubs []string
	}{
		{"empty", "", []string{"empty"}},
		{"contains slash", "external/main/x", []string{"must not contain", "/"}},
		{"contains backslash", "smb\\share", []string{"must not contain", "\\"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := runSMBUnmount(nil, nil, nopWriter{}, c.entry, &smbUnmountOptions{node: "main"})
			if err == nil {
				t.Fatal("expected error")
			}
			for _, s := range c.errSubs {
				if !strings.Contains(err.Error(), s) {
					t.Errorf("error %q missing substring %q", err.Error(), s)
				}
			}
		})
	}
}

// TestRunSMBHistoryAdd_FlagValidation pins the cobra-layer flag
// validation that fires before any wire call:
//   - --password and --password-stdin together → mutual-exclusion error
//   - --password without --user → "needs both halves" error
//   - URL not starting with `//` → format error
func TestRunSMBHistoryAdd_FlagValidation(t *testing.T) {
	cases := []struct {
		name    string
		url     string
		opts    smbHistoryAddOptions
		errSubs []string
	}{
		{
			name:    "missing // prefix",
			url:     "host/share",
			opts:    smbHistoryAddOptions{},
			errSubs: []string{"//"},
		},
		{
			name:    "password + password-stdin",
			url:     "//h/s",
			opts:    smbHistoryAddOptions{user: "u", password: "p", passwordStdin: true},
			errSubs: []string{"mutually exclusive"},
		},
		{
			name:    "password without user",
			url:     "//h/s",
			opts:    smbHistoryAddOptions{password: "p"},
			errSubs: []string{"--user", "both halves"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := runSMBHistoryAdd(nil, nil, nopWriter{}, strings.NewReader(""), c.url, &c.opts)
			if err == nil {
				t.Fatal("expected error")
			}
			for _, s := range c.errSubs {
				if !strings.Contains(err.Error(), s) {
					t.Errorf("error %q missing substring %q", err.Error(), s)
				}
			}
		})
	}
}

// TestRunSMBHistoryRm_RejectsBadURL covers the cobra-layer pre-flight
// for `history rm`: malformed URLs and the all-empty edge case.
func TestRunSMBHistoryRm_RejectsBadURL(t *testing.T) {
	t.Run("missing // prefix", func(t *testing.T) {
		err := runSMBHistoryRm(nil, nil, nopWriter{}, []string{"host/share"}, &smbHistoryRmOptions{})
		if err == nil || !strings.Contains(err.Error(), "//") {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("all-empty input", func(t *testing.T) {
		err := runSMBHistoryRm(nil, nil, nopWriter{}, []string{"   ", ""}, &smbHistoryRmOptions{})
		if err == nil || !strings.Contains(err.Error(), "no SMB urls") {
			t.Errorf("err = %v", err)
		}
	})
}

// TestDisplayUser pins the small UX helper that turns an empty
// string into the LarePass "(anonymous)" label.
func TestDisplayUser(t *testing.T) {
	if got := displayUser(""); got != "(anonymous)" {
		t.Errorf("displayUser(\"\") = %q", got)
	}
	if got := displayUser("alice"); got != "alice" {
		t.Errorf("displayUser(\"alice\") = %q", got)
	}
}

// nopWriter is a minimal io.Writer for tests that don't care about
// human-readable output. We avoid pulling in io.Discard explicitly
// so the test file's import block stays lean.
type nopWriter struct{}

func (nopWriter) Write(p []byte) (int, error) { return len(p), nil }

// newHistoryTestClient stands up an httptest.Server that serves a
// canned /api/smb_history/<node>/ response and returns an
// smbmount.Client pointed at it. Centralised so each applySMBHistory
// test can dial in its own fixture without re-writing the muxer.
//
// `entriesBody` is the raw JSON body the server emits — usually the
// bare-array shape `[{url,...}, ...]` that HistoryList expects.
// `status` lets a test simulate a 401/500/etc. failure mode (soft-
// fail behavior is one of the things this helper exercises).
func newHistoryTestClient(t *testing.T, status int, entriesBody string) *smbmount.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if status != 0 && status != http.StatusOK {
			http.Error(w, entriesBody, status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, entriesBody)
	}))
	t.Cleanup(srv.Close)
	return &smbmount.Client{HTTPClient: srv.Client(), BaseURL: srv.URL}
}

// TestReadSMBUserPrompt covers the byte-by-byte username reader.
// Three cases pin the contract:
//
//  1. Happy path — typed username followed by Enter, returned
//     verbatim without the trailing newline.
//  2. CR/LF tolerance — a Windows-style "\r\n" terminator must not
//     leak a literal `\r` into the parsed username (which would
//     then go onto the wire and almost certainly fail SMB auth).
//  3. Empty submission (just Enter) — taken at face value as the
//     "anonymous" gesture, returns an empty string with no error.
//
// We deliberately avoid testing prompt-write failures here: the
// fmt.Fprint to `out` only errors on a writer that the cobra layer
// never feeds us in practice (production `out` is the command's
// OutOrStdout, tests pass nopWriter / bytes.Buffer). A failure mode
// that can't happen in production isn't worth a brittle test fixture.
func TestReadSMBUserPrompt(t *testing.T) {
	t.Run("happy path returns the typed username", func(t *testing.T) {
		var buf bytes.Buffer
		got, err := readSMBUserPrompt(strings.NewReader("alice\n"), &buf)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "alice" {
			t.Errorf("got %q, want %q", got, "alice")
		}
		// The prompt must include the "(empty for anonymous)"
		// hint so the user knows pressing Enter is a legit
		// choice, not an error.
		if !strings.Contains(buf.String(), "empty for anonymous") {
			t.Errorf("prompt = %q, missing 'empty for anonymous' hint", buf.String())
		}
	})
	t.Run("CRLF terminator is stripped", func(t *testing.T) {
		got, err := readSMBUserPrompt(strings.NewReader("alice\r\n"), nopWriter{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// A leaked `\r` would corrupt the wire username for any
		// pasted-from-Windows input; the byte reader must drop
		// it before the trailing `\n` is consumed.
		if got != "alice" {
			t.Errorf("got %q, want %q (CR not stripped?)", got, "alice")
		}
	})
	t.Run("empty submission returns empty string for anonymous", func(t *testing.T) {
		got, err := readSMBUserPrompt(strings.NewReader("\n"), nopWriter{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "" {
			t.Errorf("got %q, want empty (anonymous)", got)
		}
	})
}

// TestPromptSMBUserIfNeeded_EarlyReturns covers every short-circuit
// path the helper must respect — these branches all run BEFORE the
// TTY check, so we can exercise them with a plain strings.Reader
// and assert the helper does NOT consume any input (i.e. it did
// short-circuit). The TTY-yes path itself is impractical to unit-
// test cleanly for the same reason TestResolveSMBPassword skips its
// interactive branch (term.IsTerminal needs a real TTY fd); that
// path is covered indirectly by TestReadSMBUserPrompt above.
//
// Each early-return case represents a "user / script has already
// committed to a credential plan, don't pop a prompt mid-flow":
//
//   - explicit -u: the username is set, prompt would be redundant.
//   - explicit -p: we're heading for a non-interactive password
//     path; the symmetric "ask for username too" gesture doesn't
//     make sense.
//   - --password-stdin: same as -p — script context, no prompt.
//   - historyPwHint set: applySMBHistoryDefaults already filled
//     o.user (the hint contract requires username match), so
//     prompting would race against a value that's already correct.
func TestPromptSMBUserIfNeeded_EarlyReturns(t *testing.T) {
	cases := []struct {
		name string
		opts smbMountOptions
		hint string
	}{
		{name: "explicit -u short-circuits", opts: smbMountOptions{user: "alice"}},
		{name: "explicit -p short-circuits", opts: smbMountOptions{password: "p"}},
		{name: "--password-stdin short-circuits", opts: smbMountOptions{passwordStdin: true}},
		{name: "history password hint short-circuits", opts: smbMountOptions{}, hint: "saved-pw"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Use a reader that would PANIC if read — guarantees
			// the early-return branches don't even touch stdin.
			panicReader := readerFunc(func(_ []byte) (int, error) {
				t.Fatal("readSMBUserPrompt must not be reached when an early-return branch applies")
				return 0, nil
			})
			before := c.opts.user
			err := promptSMBUserIfNeeded(&c.opts, c.hint, panicReader, nopWriter{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c.opts.user != before {
				t.Errorf("o.user mutated from %q to %q on a short-circuit path", before, c.opts.user)
			}
		})
	}
}

// readerFunc lets a test build an io.Reader from a closure — used
// above to assert no read happens on early-return paths.
type readerFunc func([]byte) (int, error)

func (f readerFunc) Read(p []byte) (int, error) { return f(p) }

// TestApplySMBHistoryDefaults covers the autofill bridge between
// `smb history` (saved favorites) and `smb mount` (interactive
// prompt). The bridge has three jobs:
//
//  1. When the URL matches an entry, fill in the missing -u flag
//     and return a password hint so resolveSMBPassword can skip the
//     prompt.
//  2. When the user passed an explicit -u that disagrees with the
//     favorite's username, do nothing — lending one account's
//     password to a different account is never safe.
//  3. Be a soft failure: a 5xx / 401 / network blip on
//     /api/smb_history/ must not block the mount; we log a warning
//     and let the user fall through to the prompt / explicit flags.
//
// All three are pinned below; together they describe the contract
// the cobra layer presents to the user. A regression here re-
// introduces the original UX bug (saved favorites ignored, prompt
// every time).
func TestApplySMBHistoryDefaults(t *testing.T) {
	const node = "olares-worker"
	const url = "//192.168.31.82/1231"
	// Fixture matching the user's reported repro: a favorite with
	// username + password saved, mounted via a flag-less `mount`.
	body := `[{"url":"//192.168.31.82/1231","username":"test111","password":"secret-pw"},` +
		`{"url":"//192.168.31.82/232","username":"","password":""}]`

	t.Run("URL match fills missing user + returns password hint", func(t *testing.T) {
		c := newHistoryTestClient(t, http.StatusOK, body)
		o := &smbMountOptions{} // no -u, no -p — the canonical case
		var buf bytes.Buffer
		hint := applySMBHistoryDefaults(context.Background(), c, &buf, node, url, o)
		if hint != "secret-pw" {
			t.Errorf("hint = %q, want %q", hint, "secret-pw")
		}
		if o.user != "test111" {
			t.Errorf("o.user = %q, want %q", o.user, "test111")
		}
		if !strings.Contains(buf.String(), "using saved credentials") {
			t.Errorf("expected 'using saved credentials' log line, got %q", buf.String())
		}
	})

	t.Run("--no-history skips the lookup entirely", func(t *testing.T) {
		c := newHistoryTestClient(t, http.StatusOK, body)
		o := &smbMountOptions{noHistory: true}
		var buf bytes.Buffer
		hint := applySMBHistoryDefaults(context.Background(), c, &buf, node, url, o)
		if hint != "" {
			t.Errorf("hint = %q, want empty (--no-history)", hint)
		}
		if o.user != "" {
			t.Errorf("o.user = %q, want unchanged (--no-history)", o.user)
		}
		if buf.Len() != 0 {
			t.Errorf("--no-history must not emit any log line, got %q", buf.String())
		}
	})

	t.Run("explicit -p suppresses the hint but still fills user", func(t *testing.T) {
		// User chose to pass their own password but no -u; we
		// should still adopt the favorite's username (otherwise
		// the mount goes out as anonymous with the user's
		// password, which is nonsensical SMB). The password hint
		// must be empty since the explicit -p wins.
		c := newHistoryTestClient(t, http.StatusOK, body)
		o := &smbMountOptions{password: "explicit-pw"}
		hint := applySMBHistoryDefaults(context.Background(), c, nopWriter{}, node, url, o)
		if hint != "" {
			t.Errorf("hint = %q, want empty (explicit -p)", hint)
		}
		if o.user != "test111" {
			t.Errorf("o.user = %q, want adopted from favorite", o.user)
		}
	})

	t.Run("explicit --password-stdin suppresses the hint", func(t *testing.T) {
		c := newHistoryTestClient(t, http.StatusOK, body)
		o := &smbMountOptions{passwordStdin: true}
		hint := applySMBHistoryDefaults(context.Background(), c, nopWriter{}, node, url, o)
		if hint != "" {
			t.Errorf("hint = %q, want empty (--password-stdin)", hint)
		}
	})

	t.Run("explicit -u that disagrees blocks cross-account lending", func(t *testing.T) {
		// User passed -u bob but the favorite holds test111's
		// password. We must NOT lend test111's password to bob;
		// no hint, no user override (bob stays bob), and a
		// one-line note explains why.
		c := newHistoryTestClient(t, http.StatusOK, body)
		o := &smbMountOptions{user: "bob"}
		var buf bytes.Buffer
		hint := applySMBHistoryDefaults(context.Background(), c, &buf, node, url, o)
		if hint != "" {
			t.Errorf("hint = %q, want empty (cross-account)", hint)
		}
		if o.user != "bob" {
			t.Errorf("o.user = %q, want %q (must not overwrite explicit -u)", o.user, "bob")
		}
		if !strings.Contains(buf.String(), "history has saved credentials for user") {
			t.Errorf("expected cross-account note, got %q", buf.String())
		}
	})

	t.Run("no URL match returns empty + no log", func(t *testing.T) {
		c := newHistoryTestClient(t, http.StatusOK, `[{"url":"//other.host/share","username":"x","password":"y"}]`)
		o := &smbMountOptions{}
		var buf bytes.Buffer
		hint := applySMBHistoryDefaults(context.Background(), c, &buf, node, url, o)
		if hint != "" {
			t.Errorf("hint = %q, want empty (no match)", hint)
		}
		if o.user != "" {
			t.Errorf("o.user = %q, want unchanged", o.user)
		}
		if buf.Len() != 0 {
			t.Errorf("no-match path must stay quiet, got %q", buf.String())
		}
	})

	t.Run("URL match but favorite has no password fills user only", func(t *testing.T) {
		// `//host/anon` favorite has username but empty password —
		// e.g. the user prepared the entry via `smb history add
		// //host/anon -u guest` without saving a secret. We
		// should adopt the username so the user doesn't have to
		// retype it, but the password hint stays empty and the
		// caller falls through to the prompt.
		c := newHistoryTestClient(t, http.StatusOK, `[{"url":"//host/anon","username":"guest","password":""}]`)
		o := &smbMountOptions{}
		hint := applySMBHistoryDefaults(context.Background(), c, nopWriter{}, node, "//host/anon", o)
		if hint != "" {
			t.Errorf("hint = %q, want empty (no saved password)", hint)
		}
		if o.user != "guest" {
			t.Errorf("o.user = %q, want %q", o.user, "guest")
		}
	})

	t.Run("history endpoint error is a soft failure", func(t *testing.T) {
		// 500 from the history endpoint must not block the mount —
		// the user's existing credential flow (flags / prompt)
		// has to keep working. We expect an empty hint, an
		// unchanged user, and a single-line warning so the user
		// knows autofill was attempted.
		c := newHistoryTestClient(t, http.StatusInternalServerError, "boom")
		o := &smbMountOptions{}
		var buf bytes.Buffer
		hint := applySMBHistoryDefaults(context.Background(), c, &buf, node, url, o)
		if hint != "" {
			t.Errorf("hint = %q, want empty (history endpoint failed)", hint)
		}
		if o.user != "" {
			t.Errorf("o.user = %q, want unchanged on soft-fail", o.user)
		}
		if !strings.Contains(buf.String(), "SMB history unavailable") {
			t.Errorf("expected soft-fail note, got %q", buf.String())
		}
	})
}
