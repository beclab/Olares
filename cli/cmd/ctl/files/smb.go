package files

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/beclab/Olares/cli/internal/files/smbmount"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/credential"
)

// NewSMBCommand returns the `olares-cli files smb` parent — CLI
// counterpart of the LarePass web app's "Connect to Server" flow
// (apps/packages/app/src/components/files/smb/ConnectServerStep[1-3].vue).
//
// The flow mounts an external SMB share into the per-user
// files-backend's `external/<node>/...` namespace so it shows up in
// regular `files ls` / `files cp` / etc.; from there every other
// `files` verb works against `external/<node>/<entry>/...` the same
// way it does for any other namespace.
//
// Three sub-surfaces:
//
//   - `mount`  / `unmount`   wire calls to `/api/mount` and
//     `/api/unmount` — the actual mount lifecycle.
//   - `history list / add / rm`   the per-node "Favorite Servers"
//     book the GUI plumbs through `/api/smb_history/<node>/`. Used
//     to remember `//host/share` URLs (and optionally the saved
//     credentials) across mounts.
//
// `--node` resolution mirrors `files cp` / `files upload`: explicit
// flag wins; otherwise the first entry from `/api/nodes/`. The CLI
// only fetches `/api/nodes/` when the flag is absent (cheap probe,
// only happens once per invocation).
func NewSMBCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "smb",
		Short: "mount / unmount external SMB shares and manage the per-node history",
		Long: `Mount external SMB shares into the per-user files-backend's
` + "`external/<node>/...`" + ` namespace, and manage the per-node "Favorite
Servers" history the GUI plumbs through ` + "`/api/smb_history/`" + `.

This is the CLI counterpart of the LarePass app's "Connect to Server"
modal (apps/.../components/files/smb/ConnectServerStep[1-3].vue). Once
an SMB target is mounted, every other ` + "`files`" + ` verb works against
the resulting ` + "`external/<node>/<entry>/...`" + ` path the same way it does
for any other namespace.

Wire shape:

    POST   /api/mount/[<node>/]?external_type=smb
        body: {smbPath, user, password}
        reply: {code, message, data}
            code 200 → mounted; visible at external/<node>/<entry>/
            code 300 → smbPath was a host-only address; data is the
                       list of discovered shares — re-run mount with
                       one of them.
    POST   /api/unmount/external/<node>/<name>/?external_type=smb
        body: {}
    GET    /api/smb_history/<node>/                 → array of entries
    PUT    /api/smb_history/<node>/                 body: array (upsert)
    DELETE /api/smb_history/<node>/                 body: array of {url}

Sub-commands:

    mount   <smb-url>      [-u <user>] [-p <password> | --password-stdin] [--node N]
    unmount <name>         [--node N]
    history list           [--node N] [--json]
    history add <smb-url>  [-u <user>] [-p <password> | --password-stdin] [--node N]
    history rm  <smb-url>... [--node N]

Examples:

    # Mount a specific share with credentials.
    olares-cli files smb mount //host.local/Public -u alice -p s3cret

    # Server-side share discovery: type the host alone, get a list,
    # re-run with the chosen share path.
    olares-cli files smb mount //host.local
    # → server returned 3 shares: //host.local/Public, //host.local/Movies, ...
    # → re-run mount with one of them.

    # Stash a favorite for later (no credentials, prompts at mount time).
    olares-cli files smb history add //host.local/Public

    # List the favorites for the current node.
    olares-cli files smb history list

    # Remove a favorite by URL.
    olares-cli files smb history rm //host.local/Public

    # Inspect the mounted entries (every external mount is just a child
    # of external/<node>/).
    olares-cli files ls external/<node>/

    # Unmount when done.
    olares-cli files smb unmount <entry-name>

Security:

    --password is the most convenient form for ad-hoc invocations,
    but it ends up in shell history. For scripts, prefer
    --password-stdin and pipe the password (` + "`printf '%s' \"$PW\" | ...`" + `).
    For interactive sessions, omit both flags and the CLI prompts
    for the password without echoing it to the terminal.
`,
	}
	cmd.AddCommand(
		newSMBMountCommand(f),
		newSMBUnmountCommand(f),
		newSMBHistoryCommand(f),
	)
	for _, sub := range cmd.Commands() {
		sub.SilenceUsage = true
	}
	return cmd
}

// smbMountOptions captures flag state for `files smb mount`.
type smbMountOptions struct {
	user          string
	password      string
	passwordStdin bool
	node          string
	jsonOut       bool
	// noHistory, when true, suppresses the per-node SMB history
	// lookup that would otherwise autofill missing -u / -p values
	// from a previously-saved favorite. Set via --no-history. Useful
	// when the saved credentials are known stale (e.g. the user just
	// rotated the password and wants the new value to take effect
	// without first running `smb history rm` + `smb history add`).
	noHistory bool
}

// newSMBMountCommand: `files smb mount <smb-url>`.
//
// Sends a single POST /api/mount/<node>/?external_type=smb. On the
// happy path (code 200) the new entry shows up under
// `external/<node>/`; the helper line we print afterward points the
// user at the canonical `files ls external/<node>/` discovery path.
//
// On code 300 (server returned a list of discovered shares because
// `<smb-url>` was a host-only address), we print the list and exit
// non-zero — same UX as a "pick one and re-run" prompt, scriptable.
func newSMBMountCommand(f *cmdutil.Factory) *cobra.Command {
	o := &smbMountOptions{}
	cmd := &cobra.Command{
		Use:   "mount <smb-url> [flags]",
		Short: "mount an external SMB share into external/<node>/",
		Long: `Mount an external SMB share into ` + "`external/<node>/...`" + `.

` + "`<smb-url>`" + ` is the SMB share address starting with two slashes:

    //host.local/Public           → a specific share
    //host.local                  → host-only; server replies with the
                                    list of discovered shares (the
                                    CLI prints them and exits non-zero
                                    so a script can re-target).

Credentials (resolved in priority order):

    1. Explicit flags    -u + (-p OR --password-stdin) wins for both fields.
    2. Saved favorite    If ` + "`<smb-url>`" + ` matches an entry from
                         ` + "`files smb history list`" + ` on the target node, its
                         saved username / password autofill any flags the
                         user did NOT pass. Mirrors LarePass's "Connect to
                         Server" autofill behavior; suppress with --no-history.
                         The saved password is ONLY used when the effective
                         username matches the favorite's username — we do
                         not lend one account's password to a different
                         account.
    3. Interactive       Last-resort prompt. When -u is also missing
                         (and history didn't fill it in) the CLI
                         first asks "SMB username (empty for
                         anonymous):" — echoed, since the username
                         isn't sensitive — then "SMB password:"
                         without echo. TTY is required; non-TTY
                         errors out at the password step so scripts
                         fail loud instead of hanging.

Flags:

    -u / --user            SMB username (required for non-anonymous shares
                           when no matching favorite exists)
    -p / --password        SMB password (echoed in shell history!)
    --password-stdin       read password from the first stdin line (preferred for scripts)
    --no-history           skip the per-node history autofill; force
                           explicit flags / interactive prompt only.

After a successful mount the entry appears under
` + "`external/<node>/<entry>/`" + ` — confirm with ` + "`olares-cli files ls external/<node>/`" + `.

Examples:

    # Explicit credentials (highest precedence).
    olares-cli files smb mount //host.local/Public -u alice -p s3cret

    # Reuse saved favorite — if //host.local/Public is in the per-node
    # history with a username + password, no flags needed.
    olares-cli files smb history add //host.local/Public -u alice -p s3cret
    olares-cli files smb mount //host.local/Public
    # → · using saved credentials from SMB history (user=alice)
    # → ✓ mounted ...

    # CI-friendly: pipe the password.
    printf '%s' "$SMB_PASSWORD" | \
        olares-cli files smb mount //host.local/Public -u alice --password-stdin

    # Interactive (TTY only), opting out of history autofill.
    olares-cli files smb mount //host.local/Public -u alice --no-history
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSMBMount(cmd.Context(), f, cmd.OutOrStdout(), os.Stdin, args[0], o)
		},
	}
	cmd.Flags().StringVarP(&o.user, "user", "u", "", "SMB username")
	cmd.Flags().StringVarP(&o.password, "password", "p", "", "SMB password (also: --password-stdin to read from stdin)")
	cmd.Flags().BoolVar(&o.passwordStdin, "password-stdin", false, "read SMB password from the first stdin line (mutually exclusive with --password)")
	cmd.Flags().StringVar(&o.node, "node", "", "target node (defaults to the first /api/nodes/ entry)")
	cmd.Flags().BoolVar(&o.jsonOut, "json", false, "print code-300 share list as JSON (one path per line in default mode)")
	cmd.Flags().BoolVar(&o.noHistory, "no-history", false, "skip the per-node SMB history autofill (force explicit flags / interactive prompt)")
	return cmd
}

func runSMBMount(
	ctx context.Context,
	f *cmdutil.Factory,
	out io.Writer,
	in io.Reader,
	smbURL string,
	o *smbMountOptions,
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	smbURL = strings.TrimSpace(smbURL)
	if !strings.HasPrefix(smbURL, "//") {
		return fmt.Errorf("smb url %q must start with `//` (e.g. //host.local/Public)", smbURL)
	}
	if o.password != "" && o.passwordStdin {
		return errors.New("--password and --password-stdin are mutually exclusive")
	}

	// Factory + HTTP client + node MUST be resolved before we can
	// consult the per-node SMB history. We reorder these steps
	// ahead of password resolution so the history-autofill branch
	// can fill in missing credentials BEFORE we'd otherwise drop
	// into the interactive prompt (which is the whole point of the
	// feature — mirroring LarePass's "Connect to Server" autofill).
	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return err
	}
	httpClient, err := f.HTTPClient(ctx)
	if err != nil {
		return err
	}
	client := &smbmount.Client{
		HTTPClient: httpClient,
		BaseURL:    rp.FilesURL,
	}

	node, err := resolveSMBNode(ctx, client, o.node)
	if err != nil {
		return reformatSMBHTTPErr(err, rp.OlaresID, fmt.Sprintf("mount %s", smbURL))
	}

	// History autofill — best-effort. A network blip / 401 on the
	// history endpoint must not block the mount, so we swallow the
	// error here (with a one-line warning) and fall through to the
	// flag- / prompt-driven credential path. The downstream Mount
	// call still propagates a real auth rejection if the user-
	// provided creds turn out to be wrong.
	historyPwHint := applySMBHistoryDefaults(ctx, client, out, node, smbURL, o)

	// User prompt for the username MUST come before the password
	// resolution, otherwise an interactive `mount //host/share` with
	// no flags AND no history match silently defaults to
	// (anonymous), then only prompts for the password — which is
	// exactly the trap behind `code 500: Incorrect username or
	// password` for shares that actually require a real account.
	// This call no-ops in every non-interactive / already-resolved
	// path so script flows are unaffected.
	if err := promptSMBUserIfNeeded(o, historyPwHint, in, out); err != nil {
		return err
	}

	password, err := resolveSMBPassword(o, in, out, historyPwHint)
	if err != nil {
		return err
	}

	displayNode := node
	if displayNode == "" {
		displayNode = "(no node)"
	}
	fmt.Fprintf(out, "mount: %s @ %s (user=%s)\n", smbURL, displayNode, displayUser(o.user))

	res, err := client.Mount(ctx, node, smbmount.MountOptions{
		SMBPath:  smbURL,
		User:     o.user,
		Password: password,
	})
	if err != nil {
		return reformatSMBHTTPErr(err, rp.OlaresID, fmt.Sprintf("mount %s", smbURL))
	}

	switch res.Code {
	case 200:
		fmt.Fprintf(out, "  ✓ mounted; the share is now visible at external/%s/<entry>/\n", node)
		fmt.Fprintf(out, "    confirm with: olares-cli files ls external/%s/\n", node)
		return nil
	case 300:
		// "Pick a share and re-run" — same UX the GUI surfaces in
		// ConnectServerPath.vue, scriptable here. Exit non-zero so a
		// shell `if !` branch can detect the multi-share case.
		if o.jsonOut {
			payload := map[string]any{
				"code":  300,
				"paths": res.Paths,
			}
			b, _ := json.MarshalIndent(payload, "", "  ")
			fmt.Fprintln(out, string(b))
		} else {
			fmt.Fprintf(out, "server returned %d candidate share path(s) — pick one and re-run mount:\n", len(res.Paths))
			for _, p := range res.Paths {
				fmt.Fprintf(out, "  %s\n", p)
			}
			if len(res.Paths) == 0 {
				fmt.Fprintln(out, "  (server returned no paths)")
			}
		}
		return fmt.Errorf("mount returned a multi-share list (code 300); re-run with one of the paths above")
	}
	return fmt.Errorf("mount returned unexpected code %d (message=%q)", res.Code, res.Message)
}

// applySMBHistoryDefaults consults the per-node SMB favorites for an
// entry that matches `smbURL` and uses its saved credentials to
// fill any flags the user did NOT pass on the command line.
//
// Return value is the saved-password HINT — non-empty only when we
// have a usable password to offer (history has one AND the effective
// username matches the favorite's username). `resolveSMBPassword`
// consumes this as a lower-precedence fallback than -p /
// --password-stdin but higher than the interactive prompt, so the
// "I already saved this in the favorites" case is one keystroke
// instead of a re-typed password.
//
// As a side effect this function may MUTATE `o.user` when the user
// did not pass -u — favorites are keyed by URL so adopting the
// saved username is the only way to make the saved password
// actually usable. The mount progress line later prints from
// `o.user`, so the user sees which account is being attempted.
//
// Soft-fail policy: every failure mode below is logged with a single
// `· note:` line and the function falls through (no error returned).
// The downstream Mount call is the authoritative auth gate; a flaky
// /api/smb_history/ endpoint must not block a perfectly valid
// explicit-credentials mount.
//
//   - --no-history was passed   → skip entirely, no log.
//   - history endpoint errors   → warning log, no password hint.
//   - URL doesn't match anything → no log, no password hint
//     (this is the routine "first time mounting this URL" case).
//   - URL matches but the user passed an explicit -u that disagrees
//     with the favorite's username → adopt no defaults, no password
//     hint (lending creds across accounts is never safe).
//   - URL matches but the favorite has no password → fill user only
//     if needed; still falls through to the prompt / explicit -p.
func applySMBHistoryDefaults(
	ctx context.Context,
	client *smbmount.Client,
	out io.Writer,
	node, smbURL string,
	o *smbMountOptions,
) string {
	if o.noHistory {
		return ""
	}
	entries, err := client.HistoryList(ctx, node)
	if err != nil {
		// One-line, non-blocking warning. Don't reformat through
		// reformatSMBHTTPErr — the user still gets a working CLI
		// even if history is unreachable; only surface enough info
		// to diagnose later.
		fmt.Fprintf(out, "  · note: SMB history unavailable (%v); proceeding without autofill\n", err)
		return ""
	}
	var match *smbmount.HistoryEntry
	for i := range entries {
		if entries[i].URL == smbURL {
			match = &entries[i]
			break
		}
	}
	if match == nil {
		return ""
	}

	// Effective username: explicit -u wins; otherwise we adopt the
	// favorite's. If the two disagree, the user is explicitly
	// asking for a different account — don't quietly slip in the
	// saved password (which belongs to a different identity).
	if o.user != "" && match.Username != "" && o.user != match.Username {
		fmt.Fprintf(out, "  · note: SMB history has saved credentials for user %q but -u %q was passed; using flags as-is\n",
			match.Username, o.user)
		return ""
	}
	if o.user == "" && match.Username != "" {
		o.user = match.Username
	}

	// Password hint only when:
	//   - the user did NOT pass -p / --password-stdin (those take
	//     priority and we don't want to fight them);
	//   - the favorite actually has a password to lend;
	//   - the effective username equals the favorite's username
	//     (so we're not lending creds across accounts — same
	//     guard as the disagreement branch above, but for the
	//     symmetric "favorite has user, user passed nothing" case
	//     which the o.user == "" assignment just handled).
	if o.password != "" || o.passwordStdin {
		return ""
	}
	if match.Password == "" {
		return ""
	}
	if match.Username != o.user {
		return ""
	}
	fmt.Fprintf(out, "  · using saved credentials from SMB history (user=%s)\n", displayUser(o.user))
	return match.Password
}

// readSMBUserPrompt reads a single line from `in` as the SMB username
// (echoed — unlike password input, the username is not sensitive).
//
// We deliberately read byte-by-byte rather than via bufio.NewReader
// so that when `in` IS os.Stdin in production, no read-ahead buffer
// accidentally swallows bytes that the subsequent term.ReadPassword
// call (on the same fd) needs to see. In normal canonical-mode TTY
// input this is a theoretical concern — the terminal only delivers
// bytes on Enter, one logical line at a time — but a pasted multi-
// line clipboard payload could trip it; one byte at a time keeps the
// contract simple and the fd boundary clean for ReadPassword.
//
// CR (`\r`) is stripped from the tail so a Windows-style pasted line
// doesn't smuggle a literal `\r` into the wire username.
func readSMBUserPrompt(in io.Reader, out io.Writer) (string, error) {
	if _, err := fmt.Fprint(out, "SMB username (empty for anonymous): "); err != nil {
		return "", err
	}
	var buf []byte
	one := make([]byte, 1)
	for {
		n, err := in.Read(one)
		if n > 0 {
			if one[0] == '\n' {
				break
			}
			buf = append(buf, one[0])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("read SMB username: %w", err)
		}
	}
	return strings.TrimRight(string(buf), "\r"), nil
}

// promptSMBUserIfNeeded fills `o.user` from an interactive TTY
// prompt when every cheaper credential source has failed. The
// helper is a no-op in every other path so script flows are
// untouched. Conditions for prompting (all must hold):
//
//   - `o.user == ""` — the user did NOT pass -u on the command line.
//   - `o.password == ""` and `!o.passwordStdin` — there's no flag-
//     supplied password either; we'd otherwise be heading for the
//     interactive password prompt, so ALSO prompting for the
//     username keeps the two halves of an SMB credential together
//     in one interactive gesture.
//   - `historyPwHint == ""` — applySMBHistoryDefaults didn't manage
//     to lend us a saved password. (When it did, it also mutated
//     o.user with the favorite's username, so this branch wouldn't
//     have triggered anyway.)
//   - stdin is a real TTY — non-TTY contexts cannot honor an
//     interactive prompt; we preserve the historical "silently
//     anonymous" behavior in that case rather than block on input
//     that will never arrive.
//
// Without this helper the original bug surfaces: `mount //host/share`
// with no flags AND no matching favorite silently mounts as
// (anonymous), then prompts only for a password — every non-
// anonymous share rejects with `code 500: Incorrect username or
// password` and the user has no obvious place to inject the missing
// username. With this helper they get a clear two-step
// "SMB username: ... SMB password: ..." sequence instead.
//
// Empty submission (Enter without typing) is taken at face value as
// the explicit "this share allows anonymous access" gesture — same
// wire shape as the pre-helper default, but now an audible user
// choice rather than an invisible one.
func promptSMBUserIfNeeded(o *smbMountOptions, historyPwHint string, in io.Reader, out io.Writer) error {
	if o.user != "" {
		return nil
	}
	if o.password != "" || o.passwordStdin || historyPwHint != "" {
		return nil
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return nil
	}
	user, err := readSMBUserPrompt(in, out)
	if err != nil {
		return err
	}
	o.user = user
	return nil
}

// resolveSMBPassword consolidates the input modes for the mount
// command. Priority order (matches the docstring on the cobra Long
// description so users can map flag combinations 1:1 to behavior):
//
//  1. --password         explicit literal, used as-is
//  2. --password-stdin   first line of stdin, CR/LF stripped
//  3. historyPwHint      pre-resolved by applySMBHistoryDefaults;
//                        non-empty only when the URL matched a
//                        favorite AND the effective username matches
//                        the favorite's username (no cross-account
//                        lending). Passed in rather than re-fetched
//                        here so this helper stays free of an
//                        smbmount.Client dependency and easy to
//                        unit-test in isolation.
//  4. interactive        TTY-only prompt without echo via
//                        golang.org/x/term — same UX as the
//                        original implementation.
//
// Note on empty passwords: anonymous SMB shares accept an empty
// password, so when --password is the literal "" we DO pass it
// through. The interactive / stdin branches reject empty input
// because those modes are explicit user gestures and an accidental
// empty submit is almost always a mistake.
func resolveSMBPassword(o *smbMountOptions, in io.Reader, prompt io.Writer, historyPwHint string) (string, error) {
	if o.password != "" {
		return o.password, nil
	}
	if o.passwordStdin {
		rd := bufio.NewReader(in)
		line, err := rd.ReadString('\n')
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("read password from stdin: %w", err)
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			return "", errors.New("--password-stdin: password is empty")
		}
		return line, nil
	}
	if historyPwHint != "" {
		return historyPwHint, nil
	}
	// Interactive — prompt without echo. Anonymous SMB shares are a
	// thing too: an empty password is valid (LarePass also accepts
	// it). We pass the empty string through unchanged so the server
	// is the authoritative gate.
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return "", errors.New("stdin is not a terminal — pipe a password with --password-stdin or pass --password explicitly")
	}
	if _, err := fmt.Fprint(prompt, "SMB password: "); err != nil {
		return "", err
	}
	pwBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(prompt)
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}
	return string(pwBytes), nil
}

// displayUser folds an empty user into the LarePass app's "anonymous"
// label so the progress line reads naturally.
func displayUser(u string) string {
	if u == "" {
		return "(anonymous)"
	}
	return u
}

// resolveSMBNode mirrors files cp's `--node` cascade: explicit flag
// wins; otherwise the first entry from /api/nodes/. The wire layer
// errors on an empty list — that's appropriate here because every
// SMB verb needs a concrete node in its URL (mount uses it for the
// path segment, unmount/history bake it into a fixed-shape URL).
//
// LarePass drops the node segment entirely when its `nodes` store
// is empty, but its `nodes` store is hydrated elsewhere; on a CLI
// the only way to discover nodes is /api/nodes/, so an empty result
// from there is genuinely "no path to mount on" and we surface it
// rather than send a /api/mount/?... that the server would 404.
func resolveSMBNode(ctx context.Context, client *smbmount.Client, flagNode string) (string, error) {
	if flagNode != "" {
		return flagNode, nil
	}
	nodes, err := client.FetchNodes(ctx)
	if err != nil {
		return "", err
	}
	return nodes[0].Name, nil
}

// smbUnmountOptions captures flag state for `files smb unmount`.
type smbUnmountOptions struct {
	node string
}

// newSMBUnmountCommand: `files smb unmount <name>`.
//
// `<name>` is the entry name as it appears in `external/<node>/`
// (e.g. `smb-host-share`). The wire URL is
// `/api/unmount/external/<node>/<name>/?external_type=smb`.
func newSMBUnmountCommand(f *cmdutil.Factory) *cobra.Command {
	o := &smbUnmountOptions{}
	cmd := &cobra.Command{
		Use:   "unmount <name> [--node <node>]",
		Short: "unmount a previously-mounted SMB entry from external/<node>/",
		Long: `Unmount an SMB entry from ` + "`external/<node>/`" + `.

` + "`<name>`" + ` is the entry name visible in ` + "`files ls external/<node>/`" + ` —
typically something like ` + "`smb-host-share`" + `. Use ` + "`files ls external/<node>/`" + `
first to discover the exact name.

Wire shape:

    POST /api/unmount/external/<node>/<name>/?external_type=smb
    body: {}

Examples:

    olares-cli files ls external/main/
    olares-cli files smb unmount smb-host-share --node main

After unmount the entry disappears from ` + "`external/<node>/`" + ` immediately;
re-run ` + "`files ls`" + ` to confirm.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSMBUnmount(cmd.Context(), f, cmd.OutOrStdout(), args[0], o)
		},
	}
	cmd.Flags().StringVar(&o.node, "node", "", "node hosting the mount (defaults to the first /api/nodes/ entry)")
	return cmd
}

func runSMBUnmount(
	ctx context.Context,
	f *cmdutil.Factory,
	out io.Writer,
	name string,
	o *smbUnmountOptions,
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("entry name is empty")
	}
	if strings.ContainsAny(name, "/\\") {
		return fmt.Errorf("entry name %q must not contain '/' or '\\\\'; pass the bare entry name as it appears under external/<node>/", name)
	}

	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return err
	}
	httpClient, err := f.HTTPClient(ctx)
	if err != nil {
		return err
	}
	client := &smbmount.Client{
		HTTPClient: httpClient,
		BaseURL:    rp.FilesURL,
	}

	node, err := resolveSMBNode(ctx, client, o.node)
	if err != nil {
		return reformatSMBHTTPErr(err, rp.OlaresID, fmt.Sprintf("unmount %s", name))
	}
	if node == "" {
		return errors.New("could not resolve a node for unmount; pass --node <name> explicitly")
	}

	display := fmt.Sprintf("external/%s/%s", node, name)
	fmt.Fprintf(out, "unmount: %s\n", display)
	if err := client.Unmount(ctx, "external", node, name, "smb"); err != nil {
		return reformatSMBHTTPErr(err, rp.OlaresID, fmt.Sprintf("unmount %s", display))
	}
	fmt.Fprintf(out, "  ✓ unmounted %s\n", display)
	return nil
}

// newSMBHistoryCommand groups the per-node SMB favorites verbs.
// Pure cobra plumbing — every concrete verb is in its own factory
// function below.
func newSMBHistoryCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history",
		Short: "manage the per-node SMB favorites (LarePass \"Favorite Servers\")",
		Long: `Manage the per-node SMB favorites — the same "Favorite Servers" list the
LarePass web app keeps in its "Connect to Server" dialog
(apps/.../components/files/smb/ConnectServerStep1.vue).

Wire shape (every verb hits the same URL with a different method):

    GET    /api/smb_history/<node>/                 → array of entries
    PUT    /api/smb_history/<node>/                 body: array (upsert)
    DELETE /api/smb_history/<node>/                 body: array of {url}

Each entry carries a ` + "`url`" + ` (e.g. ` + "`//host.local/Public`" + `) and optional
saved credentials. Use ` + "`history add`" + ` with -u / -p / --password-stdin to
include credentials so a future ` + "`mount`" + ` can pull them straight from the
favorites list. Use ` + "`history list --json`" + ` if you want raw access to the
saved-credential fields (` + "`username`" + ` / ` + "`password`" + ` / ` + "`timestamp`" + `).

` + "`<node>`" + ` defaults to the first ` + "`/api/nodes/`" + ` entry, same as ` + "`files smb mount`" + `.

Examples:

    olares-cli files smb history list
    olares-cli files smb history add //host.local/Public -u alice -p s3cret
    olares-cli files smb history rm  //host.local/Public
`,
	}
	cmd.AddCommand(
		newSMBHistoryListCommand(f),
		newSMBHistoryAddCommand(f),
		newSMBHistoryRmCommand(f),
	)
	for _, sub := range cmd.Commands() {
		sub.SilenceUsage = true
	}
	return cmd
}

// smbHistoryListOptions captures flag state for `history list`.
type smbHistoryListOptions struct {
	node    string
	jsonOut bool
}

func newSMBHistoryListCommand(f *cmdutil.Factory) *cobra.Command {
	o := &smbHistoryListOptions{}
	cmd := &cobra.Command{
		Use:   "list [--node <node>] [--json]",
		Short: "list the per-node SMB favorites",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSMBHistoryList(cmd.Context(), f, cmd.OutOrStdout(), o)
		},
	}
	cmd.Flags().StringVar(&o.node, "node", "", "node whose history to read (defaults to the first /api/nodes/ entry)")
	cmd.Flags().BoolVar(&o.jsonOut, "json", false, "print each entry as JSON (one per line)")
	return cmd
}

func runSMBHistoryList(
	ctx context.Context,
	f *cmdutil.Factory,
	out io.Writer,
	o *smbHistoryListOptions,
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return err
	}
	httpClient, err := f.HTTPClient(ctx)
	if err != nil {
		return err
	}
	client := &smbmount.Client{
		HTTPClient: httpClient,
		BaseURL:    rp.FilesURL,
	}
	node, err := resolveSMBNode(ctx, client, o.node)
	if err != nil {
		return reformatSMBHTTPErr(err, rp.OlaresID, "smb history list")
	}
	if node == "" {
		return errors.New("could not resolve a node for SMB history; pass --node <name> explicitly")
	}
	entries, err := client.HistoryList(ctx, node)
	if err != nil {
		return reformatSMBHTTPErr(err, rp.OlaresID, "smb history list")
	}
	if o.jsonOut {
		enc := json.NewEncoder(out)
		for _, e := range entries {
			if err := enc.Encode(e); err != nil {
				return err
			}
		}
		return nil
	}
	if len(entries) == 0 {
		fmt.Fprintf(out, "(no SMB history entries for node %q)\n", node)
		return nil
	}
	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "URL\tUSERNAME\tHAS-PASSWORD")
	for _, e := range entries {
		hasPwd := "no"
		if e.Password != "" {
			hasPwd = "yes"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\n", e.URL, displayUser(e.Username), hasPwd)
	}
	return tw.Flush()
}

// smbHistoryAddOptions captures flag state for `history add`.
type smbHistoryAddOptions struct {
	user          string
	password      string
	passwordStdin bool
	node          string
}

func newSMBHistoryAddCommand(f *cmdutil.Factory) *cobra.Command {
	o := &smbHistoryAddOptions{}
	cmd := &cobra.Command{
		Use:   "add <smb-url> [-u <user>] [-p <password> | --password-stdin] [--node <node>]",
		Short: "add or update an SMB favorite (per-node, with optional saved credentials)",
		Long: `Add or update an entry in the per-node SMB favorites.

Credential handling:

    no -u / -p     URL-only entry (mount will prompt for credentials).
    -u + -p        URL + saved credentials (mount can reuse them
                   without prompting).
    --password-stdin  same as -p but reads the secret from the first
                   stdin line (preferred for scripts).

Wire shape:

    PUT /api/smb_history/<node>/    body: [{url, username?, password?}]

Examples:

    # URL only.
    olares-cli files smb history add //host.local/Public

    # With saved credentials.
    olares-cli files smb history add //host.local/Public -u alice -p s3cret

    # CI-friendly.
    printf '%s' "$SMB_PASSWORD" | \
        olares-cli files smb history add //host.local/Public -u alice --password-stdin
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSMBHistoryAdd(cmd.Context(), f, cmd.OutOrStdout(), os.Stdin, args[0], o)
		},
	}
	cmd.Flags().StringVarP(&o.user, "user", "u", "", "SMB username to remember alongside the URL")
	cmd.Flags().StringVarP(&o.password, "password", "p", "", "SMB password to remember (also: --password-stdin)")
	cmd.Flags().BoolVar(&o.passwordStdin, "password-stdin", false, "read the password from the first stdin line (mutually exclusive with --password)")
	cmd.Flags().StringVar(&o.node, "node", "", "node whose history to write (defaults to the first /api/nodes/ entry)")
	return cmd
}

func runSMBHistoryAdd(
	ctx context.Context,
	f *cmdutil.Factory,
	out io.Writer,
	in io.Reader,
	smbURL string,
	o *smbHistoryAddOptions,
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	smbURL = strings.TrimSpace(smbURL)
	if !strings.HasPrefix(smbURL, "//") {
		return fmt.Errorf("smb url %q must start with `//` (e.g. //host.local/Public)", smbURL)
	}
	if o.password != "" && o.passwordStdin {
		return errors.New("--password and --password-stdin are mutually exclusive")
	}
	// Cred handling for `add` is more permissive than `mount`: the
	// favorite can be URL-only. Only read stdin / interactively
	// when the user explicitly asked for it via flags; an empty
	// `o.password` is taken at face value.
	password := o.password
	if o.passwordStdin {
		rd := bufio.NewReader(in)
		line, err := rd.ReadString('\n')
		if err != nil && err != io.EOF {
			return fmt.Errorf("read password from stdin: %w", err)
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			return errors.New("--password-stdin: password is empty")
		}
		password = line
	}
	// If user provided -p / --password-stdin without -u, that's
	// almost certainly a typo — the favorite is keyed by URL and
	// password without username is unusable for SMB auth.
	if (password != "" || o.passwordStdin) && o.user == "" {
		return errors.New("--password / --password-stdin requires --user; SMB auth needs both halves")
	}

	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return err
	}
	httpClient, err := f.HTTPClient(ctx)
	if err != nil {
		return err
	}
	client := &smbmount.Client{
		HTTPClient: httpClient,
		BaseURL:    rp.FilesURL,
	}

	node, err := resolveSMBNode(ctx, client, o.node)
	if err != nil {
		return reformatSMBHTTPErr(err, rp.OlaresID, "smb history add")
	}
	if node == "" {
		return errors.New("could not resolve a node for SMB history; pass --node <name> explicitly")
	}

	entry := smbmount.HistoryEntry{
		URL:      smbURL,
		Username: o.user,
		Password: password,
	}
	if err := client.HistoryUpsert(ctx, node, []smbmount.HistoryEntry{entry}); err != nil {
		return reformatSMBHTTPErr(err, rp.OlaresID, fmt.Sprintf("smb history add %s", smbURL))
	}
	creds := "no credentials"
	if entry.Username != "" && entry.Password != "" {
		creds = fmt.Sprintf("user=%s, password saved", entry.Username)
	} else if entry.Username != "" {
		creds = fmt.Sprintf("user=%s, no password", entry.Username)
	}
	fmt.Fprintf(out, "  ✓ saved favorite %s on node %s (%s)\n", smbURL, node, creds)
	return nil
}

// smbHistoryRmOptions captures flag state for `history rm`.
type smbHistoryRmOptions struct {
	node string
}

func newSMBHistoryRmCommand(f *cmdutil.Factory) *cobra.Command {
	o := &smbHistoryRmOptions{}
	cmd := &cobra.Command{
		Use:   "rm <smb-url>... [--node <node>]",
		Short: "remove one or more SMB favorites by URL",
		Long: `Remove one or more entries from the per-node SMB favorites.

Wire shape:

    DELETE /api/smb_history/<node>/    body: [{url}, ...]

Multiple URLs in a single invocation are batched into one request.

Examples:

    olares-cli files smb history rm //host.local/Public
    olares-cli files smb history rm //a/Public //b/Movies
`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSMBHistoryRm(cmd.Context(), f, cmd.OutOrStdout(), args, o)
		},
	}
	cmd.Flags().StringVar(&o.node, "node", "", "node whose history to write (defaults to the first /api/nodes/ entry)")
	return cmd
}

func runSMBHistoryRm(
	ctx context.Context,
	f *cmdutil.Factory,
	out io.Writer,
	urls []string,
	o *smbHistoryRmOptions,
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	cleaned := make([]string, 0, len(urls))
	for _, u := range urls {
		u = strings.TrimSpace(u)
		if u == "" {
			continue
		}
		if !strings.HasPrefix(u, "//") {
			return fmt.Errorf("smb url %q must start with `//` (e.g. //host.local/Public)", u)
		}
		cleaned = append(cleaned, u)
	}
	if len(cleaned) == 0 {
		return errors.New("no SMB urls given")
	}

	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return err
	}
	httpClient, err := f.HTTPClient(ctx)
	if err != nil {
		return err
	}
	client := &smbmount.Client{
		HTTPClient: httpClient,
		BaseURL:    rp.FilesURL,
	}
	node, err := resolveSMBNode(ctx, client, o.node)
	if err != nil {
		return reformatSMBHTTPErr(err, rp.OlaresID, "smb history rm")
	}
	if node == "" {
		return errors.New("could not resolve a node for SMB history; pass --node <name> explicitly")
	}
	if err := client.HistoryRemove(ctx, node, cleaned); err != nil {
		return reformatSMBHTTPErr(err, rp.OlaresID, "smb history rm")
	}
	for _, u := range cleaned {
		fmt.Fprintf(out, "  ✓ removed favorite %s on node %s\n", u, node)
	}
	return nil
}

// reformatSMBHTTPErr maps smbmount.HTTPError onto user-friendly
// messages, mirroring the rename / rm / cp / chown counterparts.
//
// Status branches:
//   - 401/403: token rejected → suggest `profile login`. Same
//     wording as the other verbs so the user gets one consistent CTA.
//   - 404: target / endpoint not found → echo what we were doing so
//     the user can re-target.
//
// Typed credential errors from the refreshing transport are surfaced
// verbatim — same rationale as reformatChownHTTPErr.
func reformatSMBHTTPErr(err error, olaresID, op string) error {
	if err == nil {
		return nil
	}
	var inv *credential.ErrTokenInvalidated
	if errors.As(err, &inv) {
		return inv
	}
	var nli *credential.ErrNotLoggedIn
	if errors.As(err, &nli) {
		return nli
	}
	var hErr *smbmount.HTTPError
	if errors.As(err, &hErr) {
		switch hErr.Status {
		case 401, 403:
			if olaresID != "" {
				return fmt.Errorf(
					"server rejected the access token (HTTP %d); please run: olares-cli profile login --olares-id %s",
					hErr.Status, olaresID,
				)
			}
			return fmt.Errorf(
				"server rejected the access token (HTTP %d); please re-run `olares-cli profile login`",
				hErr.Status,
			)
		case 404:
			return fmt.Errorf("%s: not found on the server (HTTP 404)", op)
		}
	}
	return err
}
