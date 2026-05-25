package files

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/internal/files/cp"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/credential"
)

// cpOptions are shared between the cp and mv variants — the only
// thing that differs is the wire `action` verb (and a few help-text
// strings), which is captured by the inner action argument to
// runCpMv.
type cpOptions struct {
	recursive bool
	node      string
}

// NewCpCommand: `olares-cli files cp [-r] <src>... <dst>`
//
// Copies one or more remote entries to a remote destination via the
// per-user files-backend's PATCH /api/paste/<node>/ endpoint. The wire
// action is "copy"; mv uses the same code path with action="move".
//
// The CLI takes a Unix-style stance on multi-source / target-shape:
//
//   - Trailing '/' on <dst> means "drop each source into this
//     directory, preserving its basename" — e.g. `cp foo bar baz/`
//     yields baz/foo and baz/bar. This matches `files upload`'s
//     <remote> rule and download's resolveLocalFile.
//   - Without a trailing '/' on <dst>, exactly one <src> is allowed
//     and <dst> is treated as the full target path (rename on the
//     way in). Multi-source + non-dir <dst> is rejected.
//   - --recursive / -r is required for directory sources, same
//     refusal pattern as Unix `cp -r`. -R is accepted as an alias.
//
// The PATCH endpoint returns one task_id per call (the actual byte
// movement runs on the server's task queue), so a multi-source
// invocation prints N task_ids. We don't poll for completion in
// this iteration — that's a separate concern best built once the
// ws / task-status surface stabilises.
func NewCpCommand(f *cmdutil.Factory) *cobra.Command {
	o := &cpOptions{}
	cmd := &cobra.Command{
		Use:   "cp [-r] <src>... <dst>",
		Short: "copy one or more remote files / directories to another remote location",
		Long: `Copy one or more entries between locations on the per-user files-backend.

Wire shape (one PATCH per source):

    PATCH /api/paste/<node>/  {action: "copy", source: "...", destination: "..."}

Both <src>... and <dst> use the same 3-segment frontend path as
` + "`olares-cli files ls`" + ` (e.g. ` + "`drive/Home/Documents/foo.pdf`" + `,
` + "`sync/<repo_id>/notes/`" + `, ...). Cross-volume copies (drive → sync,
external → drive, ...) are supported because the endpoint takes
plain string source/destination paths and the backend handles the
storage-class fan-out.

Destination semantics:

    --- <dst> ENDS with '/' ---
    Drop-into-directory mode. Each <src>'s basename is appended,
    preserving the dir / file marker:

        cp drive/Home/a.pdf drive/Home/Backups/
        # → /drive/Home/Backups/a.pdf

        cp -r drive/Home/old/ drive/Home/Backups/
        # → /drive/Home/Backups/old/

    --- <dst> does NOT end with '/' ---
    Rename / exact-target mode. Exactly one <src> is allowed; <dst>
    becomes the full target path:

        cp drive/Home/a.pdf drive/Home/a-backup.pdf
        # → /drive/Home/a-backup.pdf

    Multi-source with a non-directory <dst> is rejected.

Recursion:

    Directory sources require --recursive / -r (or -R). Without it
    the command refuses to operate, matching Unix ` + "`cp -r`" + ` behavior.

Node selection:

    Each PATCH call carries a {node} URL segment. The default is the
    first entry from /api/nodes/ (same default ` + "`files upload`" + ` uses).
    External / Cache fileTypes contribute their <extend> as a node
    hint per the LarePass web app's dst_node || src_node || default
    cascade. Pass --node to force a specific node for every PATCH
    in the batch.

Examples:

    # One file → directory.
    olares-cli files cp drive/Home/notes.md drive/Home/Documents/

    # Rename on copy.
    olares-cli files cp drive/Home/notes.md drive/Home/notes-2026.md

    # Recursive directory copy.
    olares-cli files cp -r drive/Home/Photos/ drive/Home/Backups/

    # Multiple sources into a directory.
    olares-cli files cp drive/Home/a.pdf drive/Home/b.pdf drive/Home/Archive/

    # Cross-volume copy (drive → sync repo).
    olares-cli files cp drive/Home/notes.md sync/<repo_id>/inbox/
`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCpMv(cmd.Context(), f, cmd.OutOrStdout(), args, cp.ActionCopy, o)
		},
	}
	registerCpMvFlags(cmd, o)
	return cmd
}

// NewMvCommand: `olares-cli files mv [-r] <src>... <dst>`. Same flow
// as cp; the wire action is "move" instead of "copy". Surfaced as a
// separate command (rather than an alias of cp --move) because users
// reach for the verb they know — making mv a typed command also
// keeps the help text honest about what each verb does.
func NewMvCommand(f *cmdutil.Factory) *cobra.Command {
	o := &cpOptions{}
	cmd := &cobra.Command{
		Use:   "mv [-r] <src>... <dst>",
		Short: "move (or rename) one or more remote files / directories",
		Long: `Move one or more entries between locations on the per-user files-backend.

Same wire endpoint as ` + "`files cp`" + ` (PATCH /api/paste/<node>/), but
with action="move" — the server moves the source instead of
copying it. Same Unix-style multi-source / target-shape rules
apply: trailing '/' on <dst> drops sources into the directory by
basename; no trailing slash + a single source = rename.

Examples:

    # Rename a file in place.
    olares-cli files mv drive/Home/notes.md drive/Home/notes-2026.md

    # Move several files into a directory.
    olares-cli files mv drive/Home/a.pdf drive/Home/b.pdf drive/Home/Archive/

    # Recursive directory move.
    olares-cli files mv -r drive/Home/Photos/ drive/Home/Backups/
`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCpMv(cmd.Context(), f, cmd.OutOrStdout(), args, cp.ActionMove, o)
		},
	}
	registerCpMvFlags(cmd, o)
	return cmd
}

// registerCpMvFlags wires the shared flag set onto a cobra command.
// Kept as one helper so cp and mv stay byte-for-byte aligned on flag
// names; if we ever add a cp-only or mv-only flag, split it back
// into per-command setups.
func registerCpMvFlags(cmd *cobra.Command, o *cpOptions) {
	cmd.Flags().BoolVarP(&o.recursive, "recursive", "r", false,
		"recursively copy/move directories (also: -R)")
	// -R is the BSD spelling — same flag, just an alias so users
	// with muscle memory either way get the expected behavior.
	cmd.Flags().BoolP("recursive-bsd", "R", false, "alias for -r")
	cmd.Flags().Lookup("recursive-bsd").Hidden = true
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		bsd, err := cmd.Flags().GetBool("recursive-bsd")
		if err == nil && bsd {
			o.recursive = true
		}
		return nil
	}
	cmd.Flags().StringVar(&o.node, "node", "",
		"override the {node} URL segment for /api/paste/<node>/ "+
			"(defaults to the first node from /api/nodes/, with External/Cache hint applied)")
}

// runCpMv is the shared cp/mv implementation. The verb-level
// difference (action) is the only thing that varies on the wire, so
// the cobra layer factors out one runner instead of two near-identical
// functions.
func runCpMv(
	ctx context.Context,
	f *cmdutil.Factory,
	out io.Writer,
	args []string,
	action cp.Action,
	o *cpOptions,
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if len(args) < 2 {
		// cobra's MinimumNArgs catches this earlier, but the runner
		// stays defensive so a future code path that constructs
		// args programmatically can't slip a bad call through.
		return fmt.Errorf("%s: need at least one <src> and one <dst>", action)
	}

	srcArgs := args[:len(args)-1]
	dstArg := args[len(args)-1]

	srcs := make([]cp.Source, 0, len(srcArgs))
	for _, a := range srcArgs {
		s, err := frontendPathToCpSource(a)
		if err != nil {
			return err
		}
		srcs = append(srcs, s)
	}
	dst, err := frontendPathToCpDestination(dstArg)
	if err != nil {
		return err
	}

	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return err
	}
	httpClient, err := f.HTTPClient(ctx)
	if err != nil {
		return err
	}
	client := &cp.Client{
		HTTPClient: httpClient,
		BaseURL:    rp.FilesURL,
	}

	// Default {node} resolution mirrors `files upload`'s policy:
	// only fetch /api/nodes/ when we actually need it, i.e. when
	// the user didn't pass --node AND no source/destination is
	// External/Cache (those carry their own node hint via the
	// path's <extend>). The skip is purely an optimisation — the
	// fetch is cheap, but avoiding an unnecessary network round-trip
	// keeps cp snappy for the common drive-↔-drive case.
	defaultNode := ""
	if o.node == "" && needsDefaultNode(srcs, dst) {
		nodes, err := client.FetchNodes(ctx)
		if err != nil {
			return reformatCpHTTPErr(err, rp.OlaresID, "fetch /api/nodes/", "")
		}
		// Defense in depth: FetchNodes already errors on empty data,
		// but guard the index so a future regression surfaces as a
		// typed error instead of an "index out of range" panic.
		if len(nodes) == 0 {
			return fmt.Errorf("/api/nodes/ returned no nodes; cannot resolve default {node} for paste")
		}
		defaultNode = nodes[0].Name
		if defaultNode == "" {
			return fmt.Errorf("/api/nodes/ returned a node with empty name; cannot resolve default {node}")
		}
	}

	ops, err := cp.Plan(srcs, dst, action, o.recursive, defaultNode, o.node)
	if err != nil {
		return err
	}

	// Plan summary first — gives the user a chance to ^C before any
	// state-changing wire call goes out.
	fmt.Fprintf(out, "%s %d entr%s:\n", action, len(ops), pluralYies(len(ops)))
	for _, op := range ops {
		fmt.Fprintf(out, "  %s → %s  (node=%s)\n", op.Source, op.Destination, op.Node)
	}

	// Serial PATCHes. Per-call failures abort the rest: paste tasks
	// are queued server-side and there's no transactional rollback,
	// so we let the user see exactly which call failed and re-run
	// from there (rather than carrying on and producing a partial
	// success they have to reason about).
	taskIDs := make([]string, 0, len(ops))
	for _, op := range ops {
		taskID, err := client.PasteOne(ctx, op)
		if err != nil {
			return reformatCpHTTPErr(err, rp.OlaresID, string(action), op.Source+" → "+op.Destination)
		}
		fmt.Fprintf(out, "  ✓ %s → %s  task=%s\n", op.Source, op.Destination, taskID)
		taskIDs = append(taskIDs, taskID)
	}
	fmt.Fprintf(out, "queued %d %s task%s: %s\n",
		len(taskIDs), action, pluralEs(len(taskIDs)),
		strings.Join(taskIDs, ", "))
	return nil
}

// frontendPathToCpSource converts a user-supplied path into the cp
// package's Source shape. The trailing '/' on the input is preserved
// as the IsDirIntent signal so cp.Plan can demand --recursive for it
// (Unix-style); subPath always starts with '/' from ParseFrontendPath
// so the wire builder can blindly concatenate.
func frontendPathToCpSource(raw string) (cp.Source, error) {
	fp, err := ParseFrontendPath(raw)
	if err != nil {
		return cp.Source{}, err
	}
	if strings.Trim(fp.SubPath, "/") == "" {
		return cp.Source{}, fmt.Errorf("refusing to use the root of %s/%s as a copy source",
			fp.FileType, fp.Extend)
	}
	return cp.Source{
		FileType:    fp.FileType,
		Extend:      fp.Extend,
		SubPath:     fp.SubPath,
		IsDirIntent: strings.HasSuffix(fp.SubPath, "/"),
	}, nil
}

// frontendPathToCpDestination is the same conversion for the <dst>
// arg. We intentionally do NOT reject a root destination (e.g. `cp
// foo drive/Home/`) — that's the legitimate "drop a file into the
// volume root" UX. Plan's per-source basename + parent join handles
// the dst.SubPath == "/" case correctly.
func frontendPathToCpDestination(raw string) (cp.Destination, error) {
	fp, err := ParseFrontendPath(raw)
	if err != nil {
		return cp.Destination{}, err
	}
	return cp.Destination{
		FileType:    fp.FileType,
		Extend:      fp.Extend,
		SubPath:     fp.SubPath,
		IsDirIntent: strings.HasSuffix(fp.SubPath, "/"),
	}, nil
}

// needsDefaultNode tells us whether ANY (src, dst) pair would fall
// through to the defaultNode in cp.ResolveNode's cascade. If every
// pair has an External/Cache side that supplies a node hint, we can
// skip the /api/nodes/ round-trip entirely.
func needsDefaultNode(srcs []cp.Source, dst cp.Destination) bool {
	for _, s := range srcs {
		// Mirror cp.ResolveNode's "dst External/Cache wins" check
		// without reaching into the package's internals.
		if isPasteMultiNode(dst.FileType) && dst.Extend != "" {
			continue
		}
		if isPasteMultiNode(s.FileType) && s.Extend != "" {
			continue
		}
		return true
	}
	return false
}

// isPasteMultiNode is a CLI-side mirror of cp.pasteMultiNodeFileTypes.
// We don't export the full set from cp because the only public use
// case is this skip-fetch optimisation; if a third caller ever needs
// it, promote it to an exported helper there.
func isPasteMultiNode(fileType string) bool {
	switch fileType {
	case "external", "cache":
		return true
	default:
		return false
	}
}

// reformatCpHTTPErr maps cp.HTTPError onto user-friendly messages,
// mirroring rm/download's reformatters. Typed credential errors from
// the refreshing transport are surfaced verbatim — see
// reformatHTTPErr in download.go for the rationale.
func reformatCpHTTPErr(err error, olaresID, op, target string) error {
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
	var hErr *cp.HTTPError
	if errors.As(err, &hErr) {
		switch hErr.Status {
		case 401, 403:
			if olaresID != "" {
				return fmt.Errorf("server rejected the access token (HTTP %d); please run: olares-cli profile login --olares-id %s",
					hErr.Status, olaresID)
			}
			return fmt.Errorf("server rejected the access token (HTTP %d); please re-run `olares-cli profile login`", hErr.Status)
		case 404:
			if target == "" {
				return fmt.Errorf("%s: not found on the server (HTTP 404)", op)
			}
			return fmt.Errorf("%s %s: not found on the server (HTTP 404)", op, target)
		}
	}
	return err
}
