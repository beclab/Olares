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

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/credential"
	"github.com/beclab/Olares/cli/internal/files/rm"
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
//     directory" intent and triggers the same check.
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

Confirmation:

    By default ` + "`rm`" + ` lists what it would delete and asks y/N. Pass
    --force / -f to skip the prompt (e.g. in scripts). In a non-TTY
    context (CI, piped stdin) we refuse without --force rather than
    guessing.

Trailing slash on a target signals "this is a directory" — the
planner errors out without --recursive in that case (Unix-style).
With --recursive both forms (` + "`foo`" + ` and ` + "`foo/`" + `) are accepted.

Examples:

    olares-cli files rm drive/Home/Documents/old.pdf
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
	for _, a := range args {
		t, err := frontendPathToRmTarget(a)
		if err != nil {
			return err
		}
		targets = append(targets, t)
	}

	groups, err := rm.Plan(targets, o.recursive)
	if err != nil {
		return err
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

// reformatRmHTTPErr maps rm.HTTPError onto user-friendly messages,
// same spirit as the download counterpart. Typed credential errors
// from the refreshing transport are surfaced verbatim — see
// reformatHTTPErr in download.go for the rationale.
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
	var hErr *rm.HTTPError
	if errors.As(err, &hErr) {
		switch hErr.Status {
		case 401, 403:
			if olaresID != "" {
				return fmt.Errorf("server rejected the access token (HTTP %d); please run: olares-cli profile login --olares-id %s",
					hErr.Status, olaresID)
			}
			return fmt.Errorf("server rejected the access token (HTTP %d); please re-run `olares-cli profile login`", hErr.Status)
		case 404:
			return fmt.Errorf("delete %s/%s%s: not found on the server (HTTP 404)",
				g.FileType, g.Extend, g.ParentSubPath)
		}
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
