package files

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/beclab/Olares/cli/internal/files/download"
	"github.com/beclab/Olares/cli/internal/files/rm"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/credential"
)

type rmOptions struct {
	recursive bool
	force     bool
}

// NewRmCommand: `olares-cli files rm [-r] [-f] <remote-path>...`
//
// Deletes one or more remote entries via the per-user files-backend's
// batch DELETE endpoint. Multiple targets sharing a parent directory
// collapse into a single wire request — the LarePass web app does the
// same, see batchDeleteFileItems in v2/common/utils.ts.
//
// Conventions:
//
//   - --recursive / -r / -R is required to remove directories. A
//     trailing '/' on a target is interpreted as "this is a
//     directory" intent and triggers the same check. Once -r is in
//     play, EVERY target in the same invocation is treated as a
//     directory — the wire dirent carries a trailing slash regardless
//     of whether the user's path string ended in `/` or not. This
//     mirrors `rm -r foo` in Unix shells: the user has declared
//     directory intent, the CLI follows through.
//   - --force / -f skips the interactive confirmation prompt. Without
//     it, we list what would be deleted and ask y/N. In a non-TTY
//     environment (CI, piped stdin) we refuse rather than guessing —
//     the user has to opt in to deletion explicitly.
//   - Removing the root of a volume (`drive/Home/`, `sync/<repo>/`,
//     ...) is rejected by the planner; that operation has to be
//     expressed differently if it's ever needed.
func NewRmCommand(f *cmdutil.Factory) *cobra.Command {
	o := &rmOptions{}
	cmd := &cobra.Command{
		Use:     "rm [-r] [-f] <remote-path>...",
		Aliases: []string{"remove", "delete"},
		Short:   "delete one or more remote files / directories",
		Long: `Delete one or more files or directories on the per-user files-backend.

Wire shape (batch DELETE per parent dir):

    DELETE /api/resources/<encParentDir>/   body: {"dirents": [...]}

Multiple targets that share a parent directory collapse into one
request, matching the LarePass web app's batchDeleteFileItems helper.
Targets across different parents send one request each, in a stable
order (sorted by fileType + extend + parent).

Preflight existence check (runs BEFORE the preview / prompt):

    Every target is Stat'd against the server before the
    confirmation prompt is shown. The check fails fast and aborts
    BEFORE printing any "will delete N entries" line if:

      - a target path doesn't exist on the server (typo / stale path);
      - the user typed ` + "`<target>/`" + ` or passed --recursive, but the
        entry on the server is actually a FILE;
      - the user typed ` + "`<target>`" + ` (no slash) without --recursive,
        but the entry is actually a DIRECTORY (same "pass -r/-R"
        CTA the planner uses).

    Volume roots (` + "`drive/Home/`" + `, ` + "`sync/<repo>/`" + `, ...) are rejected
    upstream by the planner; the preflight only sees real entries.

Confirmation:

    By default ` + "`rm`" + ` lists what it would delete and asks y/N. Pass
    --force / -f to skip the prompt (e.g. in scripts). In a non-TTY
    context (CI, piped stdin) we refuse without --force rather than
    guessing. ` + "`--force`" + ` does NOT bypass the preflight existence
    check — a missing path still aborts the operation, matching the
    safer-than-Unix-` + "`rm -f`" + ` default we've taken everywhere else in
    ` + "`olares-cli files`" + `.

Trailing slash on a target signals "this is a directory" — the
planner errors out without --recursive in that case (Unix-style).
With --recursive both forms (` + "`foo`" + ` and ` + "`foo/`" + `) are accepted, and
the wire dirent ALWAYS gets a trailing slash so the server's POSIX
driver routes the request through the directory-removal path. In
practice this means:

    files rm     <path>      → "<path> is a file"   (no trailing slash on the wire)
    files rm -r  <path>      → "<path> is a folder" (trailing slash on the wire)

regardless of how the user typed the path. Mixing files and folders
in one ` + "`-r`" + ` invocation is unusual; if you have a file to delete
alongside directories, drop the file into a separate ` + "`files rm`" + ` call
without ` + "`-r`" + `.

Examples:

    olares-cli files rm drive/Home/Documents/old.pdf
    olares-cli files rm -r drive/Home/Backups/2024
    olares-cli files rm -r drive/Home/Backups/2024/
    olares-cli files rm -rf drive/Home/junk drive/Home/scratch/
`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRm(cmd.Context(), f, cmd.OutOrStdout(), os.Stdin, args, o)
		},
	}
	cmd.Flags().BoolVarP(&o.recursive, "recursive", "r", false,
		"recursively remove directories (also: -R)")
	// -R is the BSD spelling; same flag, just an alias so users with
	// muscle memory either way get the expected behavior.
	cmd.Flags().BoolVarP(&o.force, "force", "f", false,
		"skip the interactive y/N confirmation")
	cmd.Flags().BoolP("recursive-bsd", "R", false, "alias for -r")
	cmd.Flags().Lookup("recursive-bsd").Hidden = true
	// Wire -R to the same boolean by post-parse fixup so we don't
	// have to define two separate variables.
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		bsd, err := cmd.Flags().GetBool("recursive-bsd")
		if err == nil && bsd {
			o.recursive = true
		}
		return nil
	}
	return cmd
}

func runRm(
	ctx context.Context,
	f *cmdutil.Factory,
	out io.Writer,
	in io.Reader,
	args []string,
	o *rmOptions,
) error {
	if ctx == nil {
		ctx = context.Background()
	}

	targets := make([]rm.Target, 0, len(args))
	touchesCommon := false
	for _, a := range args {
		t, err := frontendPathToRmTarget(a)
		if err != nil {
			return err
		}
		if isCommonFrontendPath(t.FileType, t.Extend) {
			touchesCommon = true
		}
		targets = append(targets, t)
	}
	if err := requireCommonBackendVersion(ctx, f, touchesCommon); err != nil {
		return err
	}

	groups, err := rm.Plan(targets, o.recursive)
	if err != nil {
		return err
	}

	// Resolve profile + HTTP client up front. Earlier revisions of
	// rm deferred this until after the confirmation prompt so a
	// `-f` script could short-circuit on a Plan-level rejection
	// without a network round-trip, but the preflight existence
	// check below NEEDS the HTTP client BEFORE the preview /
	// confirmation — printing a "will delete X" line for a path
	// that doesn't even exist on the server is misleading, and
	// proceeding to the DELETE would just produce a 404 the user
	// has to interpret. We accept the extra `/api/refresh` round-
	// trip on the cold path (the refreshing transport caches the
	// token within a process) so the preflight can give a clean,
	// path-named rejection before any state is shown as "about to
	// be deleted".
	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return err
	}
	httpClient, err := f.HTTPClient(ctx)
	if err != nil {
		return err
	}
	client := &rm.Client{
		HTTPClient: httpClient,
		BaseURL:    rp.FilesURL,
	}

	// Preflight: Stat every target to verify it exists on the server
	// and that the trailing-slash / -r intent matches the actual
	// file vs. directory kind. This refuses the typical "I typed
	// the wrong path" and "I forgot -r" cases BEFORE we print a
	// destructive-looking preview line. See preflightRm for the full
	// table of refusal arms.
	//
	// Stat reuses the same parent-listing strategy `files cat` /
	// `files download` / `files cp` use (see
	// internal/files/download/stat.go) so the check works uniformly
	// across drive / sync / cache / external / cloud namespaces.
	statClient := &download.Client{
		HTTPClient: httpClient,
		BaseURL:    rp.FilesURL,
	}
	if err := preflightRm(ctx, statClient, targets, o.recursive); err != nil {
		return reformatRmHTTPErr(err, rp.OlaresID, nil)
	}

	// Show the user the exact set of operations before any wire call
	// goes out — even with --force; deletion is destructive enough
	// that a one-line "deleting N entries" line is worth printing.
	totalDirents := 0
	for _, g := range groups {
		totalDirents += len(g.Dirents)
	}
	fmt.Fprintf(out, "will delete %d entr%s in %d batch%s:\n",
		totalDirents, pluralYies(totalDirents),
		len(groups), pluralEs(len(groups)),
	)
	for _, g := range groups {
		parent := g.ParentSubPath
		if parent == "" {
			parent = "/"
		}
		fmt.Fprintf(out, "  %s/%s%s\n", g.FileType, g.Extend, parent)
		for _, d := range g.Dirents {
			fmt.Fprintf(out, "    %s\n", d)
		}
	}

	if !o.force {
		if !term.IsTerminal(int(syscall.Stdin)) {
			return errors.New("refusing to delete without --force in a non-interactive context (no TTY)")
		}
		fmt.Fprintf(out, "proceed with deletion? [y/N]: ")
		ok, err := readYesNo(in)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintf(out, "aborted\n")
			return nil
		}
	}

	// Serial DELETE per group. Per-group failures abort the rest:
	// the user should see exactly which group failed so they can
	// re-run on a narrower set, rather than getting a partial-success
	// state with no clear "what's left".
	for _, g := range groups {
		if err := client.DeleteBatch(ctx, g); err != nil {
			return reformatRmHTTPErr(err, rp.OlaresID, g)
		}
		fmt.Fprintf(out, "  ✓ %s/%s%s (%d entr%s)\n",
			g.FileType, g.Extend, g.ParentSubPath,
			len(g.Dirents), pluralYies(len(g.Dirents)),
		)
	}
	fmt.Fprintf(out, "done: deleted %d entr%s\n", totalDirents, pluralYies(totalDirents))
	return nil
}

// frontendPathToRmTarget converts a user-supplied path (e.g.
// "drive/Home/Documents/foo.pdf" or "drive/Home/Backups/") into the
// canonical rm.Target shape. The trailing slash is preserved as the
// directory-intent signal so the planner can require --recursive for
// it.
//
// Errors when the path resolves to "the root of <fileType>/<extend>"
// — that case is intentionally unsupported because the user almost
// never means "wipe my Home/repo/bucket" and the cost of accidentally
// allowing it would be enormous.
func frontendPathToRmTarget(raw string) (rm.Target, error) {
	fp, err := ParseFrontendPath(raw)
	if err != nil {
		return rm.Target{}, err
	}
	sub := fp.SubPath
	isDir := strings.HasSuffix(sub, "/")
	clean := strings.Trim(sub, "/")
	if clean == "" {
		return rm.Target{}, fmt.Errorf("refusing to delete the root of %s/%s", fp.FileType, fp.Extend)
	}

	idx := strings.LastIndex(clean, "/")
	var (
		parentSub string
		name      string
	)
	if idx < 0 {
		// Direct child of Extend root: parent is "/".
		parentSub = "/"
		name = clean
	} else {
		parentSub = "/" + clean[:idx] + "/"
		name = clean[idx+1:]
	}
	return rm.Target{
		FileType:      fp.FileType,
		Extend:        fp.Extend,
		ParentSubPath: parentSub,
		Name:          name,
		IsDirIntent:   isDir,
	}, nil
}

// readYesNo reads one line from `in` and returns true when it starts
// with 'y' or 'Y'. Anything else (including EOF) is "no". We
// deliberately don't accept `yes`/`no` as full words separately —
// matches `rm -i`'s permissive behavior.
func readYesNo(in io.Reader) (bool, error) {
	br := bufio.NewReader(in)
	line, err := br.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return false, nil
	}
	switch strings.ToLower(line)[0] {
	case 'y':
		return true, nil
	default:
		return false, nil
	}
}

// preflightRm probes every target BEFORE any state-changing DELETE
// /api/resources/<parent>/ goes out, and refuses the operation early
// if:
//
//   - a target path doesn't exist on the server (typo / stale path);
//   - the user typed `<target>/` (trailing slash) or passed `-r`, but
//     the actual entry is a FILE on the server. Both signal directory
//     intent (the planner promotes `-r` invocations to dir dirents
//     regardless of trailing slash), so a file at that path would
//     fail confusingly server-side;
//   - the user typed `<target>` (no slash) and did NOT pass `-r`, but
//     the actual entry is a DIRECTORY on the server. The planner
//     would have sent a file-shape dirent `/foo` which the backend
//     refuses; preflight surfaces the kind mismatch with the same
//     "pass -r/-R" CTA the planner uses.
//
// Stat uses the parent-listing strategy `files cat` / `files cp` use
// (see internal/files/download/stat.go). Volume roots are not
// reachable here because frontendPathToRmTarget already rejects them.
//
// The effective directory-intent rule matches the planner exactly:
// `IsDirIntent || recursive` is taken as "the user means directory".
// That way preflight and Plan agree on which kind to enforce, and a
// `rm -r foo` (no trailing slash) gets the same treatment as
// `rm -r foo/` — the user has declared dir intent via -r.
//
// Fail-fast: the FIRST refusal stops the batch. We don't continue
// past a mismatch because the preview that follows ("will delete N
// entries in M batches") would be lying — at least one of the
// claimed deletions can't actually run.
//
// Network cost: one Stat per target. Multi-target `rm a b c` against
// the same parent dir still issues N Stats (one per leaf) because
// download.Client.Stat doesn't cache parent listings across calls —
// a documented future optimization, not the current contract.
//
// HTTP errors (auth / network) are passed through verbatim so
// reformatRmHTTPErr can attach the standard CTA.
func preflightRm(
	ctx context.Context,
	statClient *download.Client,
	targets []rm.Target,
	recursive bool,
) error {
	for _, t := range targets {
		plain := t.FileType + "/" + t.Extend + t.ParentSubPath + t.Name
		display := plain
		if t.IsDirIntent && !strings.HasSuffix(display, "/") {
			display += "/"
		}
		info, err := statClient.Stat(ctx, plain)
		if err != nil {
			if download.IsNotFound(err) {
				return fmt.Errorf(
					"rm: target %s does not exist on the server",
					display)
			}
			return err
		}
		// Effective dir intent: explicit trailing slash OR -r flag.
		// Matches the `t.IsDirIntent || recursive` rule the planner
		// uses to decide whether to add a trailing slash to the
		// wire dirent. We keep the two rules in lockstep so the
		// preflight and the Plan can never disagree on "is this a
		// directory delete".
		effectiveDir := t.IsDirIntent || recursive
		if effectiveDir && !info.IsDir {
			// The corrective CTA must name ONLY the inputs the user
			// actually typed. Three combinations reach this branch:
			//
			//   IsDirIntent  recursive   user-input → CTA
			//   ------------ ----------- ------------------------------
			//   true         true        `rm -r foo/`  → drop trailing
			//                                            '/' AND -r/-R
			//   true         false       `rm foo/`     → drop trailing
			//                                            '/'  (unreach-
			//                                            able in practice
			//                                            — the planner
			//                                            rejects trailing
			//                                            slash without -r
			//                                            upstream, but we
			//                                            keep the arm
			//                                            sound for defence
			//                                            in depth)
			//   false        true        `rm -r foo`   → drop -r/-R
			//
			// Telling a user who never typed a trailing slash to
			// "drop the trailing '/'" sends them on a confused
			// hunt for one in their command line — the previous
			// unconditional message did exactly that.
			cta := "drop the -r/-R flag"
			switch {
			case t.IsDirIntent && recursive:
				cta = "drop the trailing '/' and the -r/-R flag"
			case t.IsDirIntent:
				cta = "drop the trailing '/'"
			}
			return fmt.Errorf(
				"rm: target %s is a file on the server, not a directory; %s",
				display, cta)
		}
		if !effectiveDir && info.IsDir {
			return fmt.Errorf(
				"rm: target %s is a directory on the server; pass -r/-R to remove it recursively",
				display)
		}
	}
	return nil
}

// reformatRmHTTPErr maps rm.HTTPError / download.HTTPError onto
// user-friendly messages, same spirit as the download counterpart.
// Typed credential errors from the refreshing transport are
// surfaced verbatim — see reformatHTTPErr in download.go for the
// rationale.
//
// Like cp/mv, rm now goes through TWO packages:
//   - rm.Client.DeleteBatch for the DELETE calls (rm.HTTPError);
//   - download.Client.Stat for the preflight existence checks
//     (download.HTTPError).
//
// Both error types are mapped to the same status-code switch. The
// `g` parameter is non-nil only when the call originated from
// DeleteBatch (we know the group's parent path then); the preflight
// path passes nil and falls back to a parent-less "not found"
// message — that's fine because the preflight's own error string
// already names the offending target verbatim.
func reformatRmHTTPErr(err error, olaresID string, g *rm.Group) error {
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
	var status int
	var rmErr *rm.HTTPError
	if errors.As(err, &rmErr) {
		status = rmErr.Status
	}
	var dlErr *download.HTTPError
	if status == 0 && errors.As(err, &dlErr) {
		status = dlErr.Status
	}
	switch status {
	case 401, 403:
		if olaresID != "" {
			return fmt.Errorf("server rejected the access token (HTTP %d); please run: olares-cli profile login --olares-id %s",
				status, olaresID)
		}
		return fmt.Errorf("server rejected the access token (HTTP %d); please re-run `olares-cli profile login`", status)
	case 404:
		if g != nil {
			return fmt.Errorf("delete %s/%s%s: not found on the server (HTTP 404)",
				g.FileType, g.Extend, g.ParentSubPath)
		}
		return fmt.Errorf("rm: not found on the server (HTTP 404)")
	}
	return err
}

// pluralEs handles "batch" / "batches" — same pattern as pluralYies.
func pluralEs(n int) string {
	if n == 1 {
		return ""
	}
	return "es"
}
