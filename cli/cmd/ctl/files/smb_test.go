package files

import (
	"strings"
	"testing"
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
	if !strings.Contains(cmd.Long, "//host.local") {
		t.Error("Long help should include a sample SMB URL")
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

// TestResolveSMBPassword covers the three-mode resolver:
//   - explicit --password wins
//   - --password-stdin reads one line, trims trailing CR/LF
//   - missing both with a non-TTY stdin → error (the interactive
//     branch needs a TTY and is impractical to unit-test cleanly)
func TestResolveSMBPassword(t *testing.T) {
	t.Run("explicit password wins", func(t *testing.T) {
		got, err := resolveSMBPassword(&smbMountOptions{password: "x"}, strings.NewReader(""), nopWriter{})
		if err != nil {
			t.Fatal(err)
		}
		if got != "x" {
			t.Errorf("got %q", got)
		}
	})
	t.Run("password-stdin happy path", func(t *testing.T) {
		got, err := resolveSMBPassword(&smbMountOptions{passwordStdin: true}, strings.NewReader("s3cret\n"), nopWriter{})
		if err != nil {
			t.Fatal(err)
		}
		if got != "s3cret" {
			t.Errorf("got %q", got)
		}
	})
	t.Run("password-stdin without trailing newline", func(t *testing.T) {
		got, err := resolveSMBPassword(&smbMountOptions{passwordStdin: true}, strings.NewReader("s3cret"), nopWriter{})
		if err != nil {
			t.Fatal(err)
		}
		if got != "s3cret" {
			t.Errorf("got %q", got)
		}
	})
	t.Run("password-stdin empty input", func(t *testing.T) {
		_, err := resolveSMBPassword(&smbMountOptions{passwordStdin: true}, strings.NewReader(""), nopWriter{})
		if err == nil {
			t.Fatal("expected error for empty stdin password")
		}
		if !strings.Contains(err.Error(), "empty") {
			t.Errorf("error = %v", err)
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
