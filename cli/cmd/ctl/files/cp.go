package files

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/internal/files/cp"
	"github.com/beclab/Olares/cli/internal/files/download"
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
//   - --recursive / -r is required for directory sources, same
//     refusal pattern as Unix `cp -r`. -R is accepted as an alias.
//
// Renaming via `cp` (single source + non-dir <dst>) is intentionally
// undocumented in the user-facing help — point users at
// `olares-cli files rename` for in-place basename changes. The
// planner still accepts that shape for backwards compatibility, but
// the help text only promotes the drop-into-directory UX.
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

    <dst> MUST end with '/' (drop-into-directory mode). Each <src>'s
    basename is appended, preserving the dir / file marker:

        cp drive/Home/a.pdf drive/Home/Backups/
        # → /drive/Home/Backups/a.pdf

        cp -r drive/Home/old/ drive/Home/Backups/
        # → /drive/Home/Backups/old/

    Renaming via ` + "`cp`" + ` is not currently supported — for in-place
    basename changes use ` + "`olares-cli files rename`" + ` (or move into a
    directory under a different parent with ` + "`files mv`" + ` after the
    rename).

Recursion:

    Directory sources require --recursive / -r (or -R). Without it
    the command refuses to operate, matching Unix ` + "`cp -r`" + ` behavior.

Preflight existence check:

    Before any PATCH /api/paste/<node>/ is sent, the CLI Stats every
    source and the destination directory:

      - each <src> MUST exist on the server, AND its trailing slash
        must match the actual file/dir kind (trailing '/' for
        directories, none for files);
      - <dst> MUST exist as a directory on the server (create it
        first with ` + "`olares-cli files mkdir`" + ` if it doesn't yet exist).

    The check fails fast — without the preflight a typo could
    enqueue paste tasks that fail asynchronously on the server's
    task queue with no transactional rollback, leaving the user to
    reason about a partial-success batch.

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
		Short: "move one or more remote files / directories into a destination directory",
		Long: `Move one or more entries between locations on the per-user files-backend.

Same wire endpoint as ` + "`files cp`" + ` (PATCH /api/paste/<node>/), but
with action="move" — the server moves the source instead of
copying it. Same Unix-style multi-source / target-shape rule
applies: <dst> MUST end with '/' (drop-into-directory mode); each
source is dropped into the directory by basename.

Renaming via ` + "`mv`" + ` is not currently supported — for in-place
basename changes use ` + "`olares-cli files rename`" + `.

Preflight existence check (same as ` + "`files cp`" + `):

    Before the PATCH goes out, every <src> is Stat'd (must exist;
    trailing slash must match file/dir kind) and <dst> is Stat'd
    (must exist as a directory). A missing source or destination
    directory aborts the operation before any state change.

Examples:

    # Move one file into a directory.
    olares-cli files mv drive/Home/notes.md drive/Home/Archive/

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

	// Preflight: verify every source exists on the server, the
	// trailing-slash on the source matches the actual file/dir kind,
	// and the destination (or its parent for the undocumented
	// exact-target shape) exists as a directory. This catches the
	// common typo cases ("did you mean Documents/?" / "you forgot
	// to mkdir the target") BEFORE we enqueue any paste tasks —
	// once a paste task is on the queue we have no transactional
	// rollback, and a partial-success batch is harder to reason
	// about than a clean refusal up front.
	//
	// Stat uses the same parent-listing strategy `files cat` and
	// `files download` use (see internal/files/download/stat.go) so
	// it works uniformly across drive / sync / cache / external /
	// cloud namespaces. We pass the same HTTPClient cp uses so the
	// preflight inherits the refreshing-transport's 401/403 retry.
	statClient := &download.Client{
		HTTPClient: httpClient,
		BaseURL:    rp.FilesURL,
	}
	if err := preflightCpMv(ctx, statClient, srcs, dst, action); err != nil {
		return reformatCpHTTPErr(err, rp.OlaresID, string(action)+" preflight", "")
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

// preflightCpMv probes every source and the destination side BEFORE
// any state-changing PATCH /api/paste/<node>/ goes out, and refuses
// the operation early if:
//
//   - a source path doesn't exist on the server (typo / stale path);
//   - a source's trailing-slash intent doesn't match the actual file
//     vs. directory kind on the server (e.g. user typed
//     `cp drive/Home/Photos` without the trailing '/' but Photos is
//     actually a directory — they would have hit "is a directory:
//     pass -r/-R" via the planner only if they ALSO typed `Photos/`,
//     so we surface the kind mismatch here);
//   - in drop-into-dir mode (the documented shape, `<dst>` ends with
//     '/'), the destination directory doesn't exist OR is a file;
//   - in exact-target mode (the undocumented single-source +
//     non-'/'-terminated `<dst>` shape), the destination's parent
//     directory doesn't exist OR is a file.
//
// The Stat strategy is the same parent-listing approach `files cat`
// and `files download` use (see internal/files/download/stat.go).
// Volume roots (depth-0/1 subpaths like `drive/Home/`, `sync/<repo>/`)
// stat as synthetic directories — they always "exist" in the
// files-backend model — so a `cp foo drive/Home/` works without an
// extra round-trip.
//
// Performance:
//   - One Stat per unique source path. Multi-source `cp a b c dst/`
//     sharing the same parent dir still issues N Stats (one per
//     leaf); the redundancy is small for typical N≤10 and not worth
//     a parent-listing cache yet.
//   - One Stat for the destination side (the dst dir itself, or its
//     parent for exact-target mode).
//
// Returns nil on success; otherwise a typed error whose Error()
// names the offending path and the corrective action. HTTP errors
// (auth / network) are passed through verbatim so reformatCpHTTPErr
// can attach the standard CTA.
func preflightCpMv(
	ctx context.Context,
	statClient *download.Client,
	srcs []cp.Source,
	dst cp.Destination,
	action cp.Action,
) error {
	for _, s := range srcs {
		srcDisplay := s.FileType + "/" + s.Extend + s.SubPath
		srcPlain := s.FileType + "/" + s.Extend + s.SubPath
		info, err := statClient.Stat(ctx, srcPlain)
		if err != nil {
			if download.IsNotFound(err) {
				return fmt.Errorf("%s: source %s does not exist on the server",
					action, srcDisplay)
			}
			return err
		}
		// Volume roots stat as synthetic directories; the planner
		// has already rejected SubPath=="/" sources, so a synthetic
		// dir reaching here means the user gave us at least one
		// real path segment — info.IsDir reflects the actual leaf.
		if s.IsDirIntent && !info.IsDir {
			return fmt.Errorf(
				"%s: source %s is a file on the server, not a directory; drop the trailing '/'",
				action, srcDisplay)
		}
		if !s.IsDirIntent && info.IsDir {
			return fmt.Errorf(
				"%s: source %s is a directory on the server; add a trailing '/' and pass -r/-R to %s it recursively",
				action, srcDisplay, action)
		}
	}

	// Destination side. The two shapes correspond to the two
	// destination modes the planner accepts:
	//
	//   - dst.IsDirIntent == true  → drop-into-directory mode.
	//     The dst path ITSELF must exist as a directory.
	//   - dst.IsDirIntent == false → exact-target mode (the
	//     undocumented single-source rename shape kept for
	//     backwards compatibility). The dst's PARENT directory
	//     must exist; the leaf doesn't exist yet by definition
	//     (or, if it does, the backend's auto-rename / overwrite
	//     handling takes over, which is server-side behavior we
	//     don't preflight here).
	if dst.IsDirIntent {
		dstDisplay := dst.FileType + "/" + dst.Extend + dst.SubPath
		dstPlain := dst.FileType + "/" + dst.Extend + dst.SubPath
		info, err := statClient.Stat(ctx, dstPlain)
		if err != nil {
			if download.IsNotFound(err) {
				return fmt.Errorf(
					"%s: destination directory %s does not exist on the server; create it first with `olares-cli files mkdir`",
					action, dstDisplay)
			}
			return err
		}
		if !info.IsDir {
			return fmt.Errorf(
				"%s: destination %s is a file on the server, not a directory; drop the trailing '/' or pick a different target",
				action, dstDisplay)
		}
		return nil
	}

	// Exact-target mode: stat the parent directory only. The
	// planner has already rejected SubPath=="/" via the dir-intent
	// requirement on root targets, so dst.SubPath here is at least
	// "/<leaf>".
	parentSub := parentSubPath(dst.SubPath)
	parentDisplay := dst.FileType + "/" + dst.Extend + parentSub
	parentPlain := dst.FileType + "/" + dst.Extend + parentSub
	info, err := statClient.Stat(ctx, parentPlain)
	if err != nil {
		if download.IsNotFound(err) {
			return fmt.Errorf(
				"%s: destination's parent directory %s does not exist on the server",
				action, parentDisplay)
		}
		return err
	}
	if !info.IsDir {
		return fmt.Errorf(
			"%s: destination's parent %s is a file on the server, not a directory",
			action, parentDisplay)
	}
	return nil
}

// parentSubPath returns the parent of a `<sub-path>` (always a string
// starting with '/'). Examples:
//
//	"/Documents/foo.pdf"  → "/Documents/"
//	"/Documents/sub/"     → "/Documents/"   (treat the input as the dir itself)
//	"/foo.pdf"            → "/"             (parent is the extend root)
//	"/"                   → "/"             (extend root — its own parent)
//
// The trailing '/' on the returned parent is preserved so Stat
// receives the canonical directory-form path; download.Stat trims it
// internally before splitting on '/', but keeping the slash here
// keeps the wire shape consistent with the rest of the codebase.
func parentSubPath(sub string) string {
	s := strings.TrimRight(sub, "/")
	if s == "" {
		return "/"
	}
	i := strings.LastIndex(s, "/")
	if i < 0 {
		return "/"
	}
	return s[:i+1]
}

// reformatCpHTTPErr maps cp.HTTPError / download.HTTPError onto
// user-friendly messages, mirroring rm/download's reformatters.
// Typed credential errors from the refreshing transport are
// surfaced verbatim — see reformatHTTPErr in download.go for the
// rationale.
//
// We branch on both error types because the cp cobra flow now goes
// through TWO packages:
//   - cp.Client.PasteOne for the PATCH calls (cp.HTTPError);
//   - download.Client.Stat for the preflight existence checks
//     (download.HTTPError).
//
// Status code mapping (401/403 → re-login CTA, 404 → "not found")
// is the same for both — we keep one switch block but two errors.As
// branches so the type-specific status / URL fields stay accessible
// for future per-package diagnostics.
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
	var status int
	var cpErr *cp.HTTPError
	if errors.As(err, &cpErr) {
		status = cpErr.Status
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
		if target == "" {
			return fmt.Errorf("%s: not found on the server (HTTP 404)", op)
		}
		return fmt.Errorf("%s %s: not found on the server (HTTP 404)", op, target)
	}
	return err
}
