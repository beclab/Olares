package me

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings me password ...`
//
// Backed by POST /api/users/<username>/password on user-service. The SPA
// calls this from the Settings -> Account -> Password change-password
// dialog, with the body shape (`stores/settings/user.ts:121`):
//
//   { username, current_password, password }
//
// Both passwords are pre-salted on the SPA via passwordAddSort →
// saltedMD5, which only kicks in for OS version >= 1.12.0-0. We mirror
// that exact behavior in passwordhash.go so the CLI works against both
// fresh installs and pre-1.12 instances.
//
// Note: BFL also exposes a more direct PUT /bfl/iam/v1alpha1/users/:user/
// password. We pick the user-service flow on purpose because it's the
// SPA-facing one (so server-side guards / auditing match what the SPA
// users see), and the wire shape is the one the SPA tests exercise.
//
// Role: any authenticated user can change their *own* password (this is
// the Person page, not Accounts), so no PreflightRole gate. Changing
// somebody else's password lives in Phase 3 (admin-scoped users CRUD)
// and will need a separate verb / role check.

func NewPasswordCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "password",
		Short: "self-service password management (Settings -> Person -> Password)",
		Long: `Self-service password management for the currently signed-in user.

Subcommands:
  set                          change the current user's password
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newPasswordSetCommand(f))
	return cmd
}

func newPasswordSetCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		stdinPasswords bool
	)
	cmd := &cobra.Command{
		Use:   "set",
		Short: "change the current user's password",
		Long: `Change the password of the currently signed-in Olares user.

By default the command interactively prompts for the current password
and the new password (twice, the second time as a confirmation). For
automation, pass --passwords-stdin and pipe two newline-separated
passwords on stdin: the current one first, the new one second.
Confirmation is skipped in that mode.

Example (interactive):
  olares-cli settings me password set

Example (automation):
  printf '%s\n%s\n' "$CURRENT" "$NEW" |
    olares-cli settings me password set --passwords-stdin

The new password is hashed locally with the same MD5+salt scheme the
SPA uses (when the target OS version is >= 1.12.0-0); your raw
password never leaves the machine.

After a successful change, your existing CLI access token may still be
valid for a while (Authelia caches sessions), but if subsequent CLI
calls return 401, re-login with:

  olares-cli profile login --olares-id <olares-id>
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runPasswordSet(c.Context(), f, stdinPasswords)
		},
	}
	cmd.Flags().BoolVar(&stdinPasswords, "passwords-stdin", false, "read current and new passwords as two newline-separated lines from stdin (skips interactive confirmation)")
	return cmd
}

func runPasswordSet(ctx context.Context, f *cmdutil.Factory, stdinPasswords bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}

	username := localPart(pc.profile.OlaresID)
	if username == "" {
		return fmt.Errorf("could not derive username from olares-id %q (expected <user>@<domain>)", pc.profile.OlaresID)
	}

	// Probe /api/olares-info up front so the salt-vs-no-salt decision is
	// made *before* we ask for the password. If the probe fails we keep
	// going with an empty version string, which falls through to the
	// "no salt" branch — the upstream BFL will reject mismatched hashes
	// equally cleanly.
	var info olaresInfoResp
	_ = doGetEnvelope(ctx, pc.doer, "/api/olares-info", &info)

	current, next, err := readPasswords(stdinPasswords, os.Stdin, os.Stderr)
	if err != nil {
		return err
	}

	body := map[string]string{
		"username":         username,
		"current_password": saltedPassword(current, info.OsVersion),
		"password":         saltedPassword(next, info.OsVersion),
	}
	path := "/api/users/" + url.PathEscape(username) + "/password"

	if err := doMutateEnvelope(ctx, pc.doer, "POST", path, body, nil); err != nil {
		return err
	}
	fmt.Println("Password updated. If subsequent CLI commands return 401, run `olares-cli profile login` to refresh your session.")
	return nil
}

// localPart extracts "alice" from "alice@olares.com". Returns "" when
// the input is empty or has no "@" separator (callers should treat that
// as a soft error and surface a clear message).
func localPart(olaresID string) string {
	olaresID = strings.TrimSpace(olaresID)
	if olaresID == "" {
		return ""
	}
	if i := strings.IndexByte(olaresID, '@'); i > 0 {
		return olaresID[:i]
	}
	return olaresID
}

// readPasswords centralizes the two input modes. Interactive mode goes
// through golang.org/x/term so the password isn't echoed; stdin mode is
// non-interactive and reads two newline-separated lines.
func readPasswords(stdinMode bool, in io.Reader, prompt io.Writer) (string, string, error) {
	if stdinMode {
		return readPasswordsFromStdin(in)
	}
	return readPasswordsInteractive(prompt)
}

func readPasswordsFromStdin(in io.Reader) (string, string, error) {
	rd := bufio.NewReader(in)
	current, err := rd.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", "", fmt.Errorf("read current password from stdin: %w", err)
	}
	current = strings.TrimRight(current, "\r\n")
	next, err := rd.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", "", fmt.Errorf("read new password from stdin: %w", err)
	}
	next = strings.TrimRight(next, "\r\n")
	if current == "" {
		return "", "", fmt.Errorf("--passwords-stdin: current password is empty")
	}
	if next == "" {
		return "", "", fmt.Errorf("--passwords-stdin: new password is empty")
	}
	return current, next, nil
}

func readPasswordsInteractive(prompt io.Writer) (string, string, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return "", "", fmt.Errorf("stdin is not a terminal — pipe passwords with --passwords-stdin instead")
	}

	current, err := promptHidden(prompt, "Current password: ")
	if err != nil {
		return "", "", err
	}
	if current == "" {
		return "", "", fmt.Errorf("current password is empty")
	}

	next, err := promptHidden(prompt, "New password: ")
	if err != nil {
		return "", "", err
	}
	if next == "" {
		return "", "", fmt.Errorf("new password is empty")
	}

	confirm, err := promptHidden(prompt, "Confirm new password: ")
	if err != nil {
		return "", "", err
	}
	if confirm != next {
		return "", "", fmt.Errorf("password confirmation does not match")
	}
	return current, next, nil
}

func promptHidden(prompt io.Writer, label string) (string, error) {
	if _, err := fmt.Fprint(prompt, label); err != nil {
		return "", err
	}
	pwBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	// Always print a newline so subsequent prompts (or the success line)
	// don't run on the same row as the hidden input.
	fmt.Fprintln(prompt)
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}
	return string(pwBytes), nil
}
