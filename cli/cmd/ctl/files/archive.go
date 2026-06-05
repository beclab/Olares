package files

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/beclab/Olares/cli/internal/files/archive"
	"github.com/beclab/Olares/cli/internal/files/download"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/credential"
)

// NewArchiveCommand builds the `olares-cli files archive` verb group —
// the read-only counterpart to the top-level `files compress` /
// `files extract` verbs. Currently exposes:
//
//	files archive entries <archive>            — stream the archive's
//	                                              entry list (NDJSON
//	                                              under the hood; the
//	                                              cobra layer renders
//	                                              it as a tabular or
//	                                              JSON view).
//
//	files archive cat <archive> <inner-path>   — write a single archive
//	                                              member's bytes to
//	                                              stdout or to a local
//	                                              file via -o/--output.
//
// Why a separate verb group for the read-only pair:
//
//   - `compress` and `extract` are state-mutating, asynchronous (return
//     a task_id) and high-frequency — they earn top-level verbs.
//   - `entries` / `cat` are read-only, synchronous (streaming) and
//     primarily diagnostic ("what's in this archive without un-tarring
//     all of it?") — grouping them under `archive` keeps the parent
//     verb list crisp without burying the rare-use diagnostic verbs.
//
// All four endpoints share the `/api/archive/<node>/` wire prefix
// — see `internal/files/archive/client.go` for the shared scaffolding.
func NewArchiveCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive",
		Short: "inspect archives on the per-user files-backend (entries, cat single member)",
		Long: `Inspect archives stored on the per-user files-backend without unpacking
them first. The two sub-verbs share the ` + "`/api/archive/<node>/`" + ` wire prefix
with ` + "`files compress`" + ` / ` + "`files extract`" + `:

    files archive entries <archive>            — stream the archive's entry list
    files archive cat <archive> <inner-path>   — write a single member's bytes

Supported archive formats (same set as ` + "`files compress`" + ` / ` + "`files extract`" + `):

    zip, 7z, tar, tar.gz, tgz, tar.bz2, tar.xz, gzip, bzip2, xz

Passwords (zip / 7z only) are read from stdin via --password-stdin so they
never appear in shell history or process listings.

Examples:

    # Preview an archive's contents without extracting.
    olares-cli files archive entries drive/Home/Backups/2026-Q1.zip

    # JSON output (one object per line) for pipelines.
    olares-cli files archive entries drive/Home/Backups/2026-Q1.zip --json

    # Dump one entry's bytes to stdout (binary-safe).
    olares-cli files archive cat drive/Home/Backups/2026-Q1.zip notes.md

    # Or save the entry to a local file by inferring the destination.
    olares-cli files archive cat drive/Home/Backups/2026-Q1.zip notes.md -o ./notes.md
`,
	}
	for _, sub := range []*cobra.Command{
		newArchiveEntriesCommand(f),
		newArchiveCatCommand(f),
	} {
		sub.SilenceUsage = true
		cmd.AddCommand(sub)
	}
	return cmd
}

// archiveEntriesOptions captures flags exclusive to `archive entries`.
type archiveEntriesOptions struct {
	format         string
	node           string
	passwordStdin  bool
	jsonOutput     bool
	maxEntries     int
}

// newArchiveEntriesCommand returns the `files archive entries
// <archive>` sub-command. Streams the archive's entry list as
// NDJSON under the hood; renders either a human-readable table
// (default) or one JSON object per line (--json) on stdout.
func newArchiveEntriesCommand(f *cmdutil.Factory) *cobra.Command {
	o := &archiveEntriesOptions{}
	cmd := &cobra.Command{
		Use:   "entries <archive-path>",
		Short: "stream the entry list of an archive on the per-user files-backend",
		Long: `Stream every member of an archive stored on the per-user files-backend
without unpacking it first.

Wire shape:

    GET /api/archive/<node>/entries?source=<archive>
    Content-Type: application/x-ndjson; one object per line.

Output modes:

    Default (--json=false): a human-readable table — kind, size, modified
    time, encrypted flag, and the in-archive path.

    --json: one JSON object per line, matching the wire shape verbatim
    (path / size / modified / is_dir / encrypted). Useful in pipelines
    (e.g. ` + "`olares-cli files archive entries ... --json | jq '.path'`" + `).

Format detection:

    The server infers the archive container from the source's filename
    extension (.zip / .7z / .tar.gz / .tgz / ...) — no client-side hint
    travels on the wire. We still accept --format locally so the CLI can
    pre-validate flag combinations (e.g. --password-stdin only on zip /
    7z); pass it when the file has no canonical extension and you need
    those checks to fire.

Preview constraints (mirrors the LarePass web app):

    Bare single-stream compressors — bzip2 (.bz2 / .bzip2) and xz (.xz)
    — have no listable entry table, so previewing them is rejected up
    front. To get their single decompressed payload, unpack the archive
    with ` + "`olares-cli files extract`" + ` instead. The tar.* compounds
    (tar.gz / tar.bz2 / tar.xz / tgz) are real tar containers and remain
    fully previewable.

Passwords (zip / 7z only):

    Use --password-stdin to read the password from STDIN (echo or here-
    doc). Passing it on the command line would leak through shell
    history / ` + "`ps`" + ` listings.

Examples:

    # Tabular preview.
    olares-cli files archive entries drive/Home/Backups/2026-Q1.zip

    # JSON pipeline.
    olares-cli files archive entries drive/Home/Backups/2026-Q1.zip \
        --json | jq '.path'

    # Encrypted 7z.
    echo "s3cret" | olares-cli files archive entries \
        drive/Home/Vault/data.7z --password-stdin
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runArchiveEntries(cmd.Context(), f, cmd.OutOrStdout(), args[0], o)
		},
	}
	cmd.Flags().StringVar(&o.format, "format", "",
		"archive format (one of: "+archive.JoinFormats()+"); derived from the source extension when omitted")
	cmd.Flags().StringVar(&o.node, "node", "",
		"override the {node} URL segment for /api/archive/<node>/ (defaults to the first /api/nodes/ entry)")
	cmd.Flags().BoolVar(&o.passwordStdin, "password-stdin", false,
		"read the archive password from STDIN (zip / 7z only); avoids leaking through shell history")
	cmd.Flags().BoolVar(&o.jsonOutput, "json", false,
		"emit one JSON object per line instead of the human-readable table")
	cmd.Flags().IntVar(&o.maxEntries, "max-entries", 0,
		"stop after this many entries (0 = no limit); useful for previews of huge archives")
	return cmd
}

func runArchiveEntries(
	ctx context.Context,
	f *cmdutil.Factory,
	out io.Writer,
	archiveArg string,
	o *archiveEntriesOptions,
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := requireArchiveBackendVersion(ctx, f); err != nil {
		return err
	}

	src, srcWire, err := parseArchiveSource(archiveArg, "archive entries")
	if err != nil {
		return err
	}
	if err := requireCommonBackendVersion(ctx, f, isCommonFrontendPath(src.FileType, src.Extend)); err != nil {
		return err
	}

	format := o.format
	if format == "" {
		format = archive.FormatFromExtension(archiveArg)
		if format == "" {
			return fmt.Errorf(
				"entries: cannot derive --format from %q; pass --format (one of: %s)",
				archiveArg, archive.JoinFormats())
		}
	}
	if !archive.IsSupportedFormat(format) {
		return fmt.Errorf("entries: unsupported --format %q; valid formats: %s",
			format, archive.JoinFormats())
	}
	// Bare single-stream compressors (bzip2 / xz) have no listable
	// entry table — there is nothing to enumerate. Refuse the
	// preview up front instead of streaming an empty/misleading
	// listing. Mirrors LarePass's unsupportedArchivePreviewExtensions
	// gate. (The tar.* compound formats remain previewable.)
	if !archive.SupportsPreview(format) {
		return fmt.Errorf(
			"entries: previewing %q archives is not supported — bzip2 / xz are raw single-stream "+
				"compressors with no listable entries; unpack it with `olares-cli files extract` instead",
			format)
	}

	password, err := readArchivePasswordStdin(o.passwordStdin)
	if err != nil {
		return err
	}
	if password != "" && !archive.SupportsPassword(format) {
		return fmt.Errorf(
			"entries: --password-stdin is only supported on passwordable formats (zip, 7z); got format %q",
			format)
	}

	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return err
	}
	// Streaming verbs use HTTPClientWithoutTimeout — a large archive
	// (think 10k+ entries on a slow disk) easily exceeds the standard
	// 30 s budget. Same rationale as `files cat`.
	httpClient, err := f.HTTPClientWithoutTimeout(ctx)
	if err != nil {
		return err
	}
	cli := &archive.Client{
		HTTPClient: httpClient,
		BaseURL:    rp.FilesURL,
	}

	node, err := resolveArchiveNode(ctx, f, rp, []frontendPathLike{src}, o.node)
	if err != nil {
		return err
	}

	if err := preflightArchiveSource(ctx, rp, httpClient, srcWire, "entries"); err != nil {
		return reformatArchiveHTTPErr(err, rp.OlaresID, "entries preflight", srcWire)
	}

	// The column header in tabular mode is printed LAZILY — on the
	// first streamed entry, inside the cb — rather than up front.
	// That keeps the password-retry loop below clean: an encrypted
	// archive fails to open BEFORE the walk starts (count stays 0,
	// nothing printed), so a re-attempt with a freshly prompted
	// password starts from a blank slate instead of duplicating a
	// header row. Mirrors TermiPass's preview password re-prompt.
	count := 0
	headerPrinted := false
	var total int
	pw := password
	for attempt := 0; ; attempt++ {
		count = 0
		t, serr := cli.StreamEntries(ctx, archive.EntriesOptions{
			Source: srcWire,
			Format: format,
			Node:   node,
		}, pw, func(e archive.Entry) error {
			count++
			if o.maxEntries > 0 && count > o.maxEntries {
				// Sentinel error to abort the stream — see
				// archive.StreamEntries' cb contract. We pick a
				// distinctive value so the caller doesn't mistake
				// it for a real error.
				return errArchiveEntriesMaxReached
			}
			if o.jsonOutput {
				b, jerr := json.Marshal(e)
				if jerr != nil {
					return fmt.Errorf("marshal entry %s: %w", e.Path, jerr)
				}
				fmt.Fprintf(out, "%s\n", b)
				return nil
			}
			if !headerPrinted {
				fmt.Fprintf(out, "%-6s %12s %20s %-3s %s\n", "KIND", "SIZE", "MODIFIED", "ENC", "PATH")
				headerPrinted = true
			}
			kind := "FILE"
			if e.IsDir {
				kind = "DIR"
			}
			enc := "no"
			if e.Encrypted {
				enc = "yes"
			}
			modStr := "-"
			if e.Modified > 0 {
				modStr = time.Unix(e.Modified, 0).UTC().Format(time.RFC3339)
			}
			sizeStr := "-"
			if !e.IsDir {
				sizeStr = formatBytes(e.Size)
			}
			fmt.Fprintf(out, "%-6s %12s %20s %-3s %s\n", kind, sizeStr, modStr, enc, e.Path)
			return nil
		})
		total = t

		if serr == nil {
			break
		}
		if errors.Is(serr, errArchiveEntriesMaxReached) {
			// Soft abort — the user asked for a head-style preview.
			// In --json mode this notice must NOT go to `out`: it
			// would interleave a non-JSON line into the NDJSON stream
			// and break line-oriented consumers (jq et al.). Mirror
			// the completion-footer policy below — divert it to
			// stderr so the user still sees it while stdout stays a
			// clean record stream. Tabular mode keeps it on `out`.
			notice := out
			if o.jsonOutput {
				notice = os.Stderr
			}
			fmt.Fprintf(notice, "\n(stopped after %d entries; pass --max-entries 0 for the full list)\n",
				o.maxEntries)
			return nil
		}
		// Only a pre-walk password failure (count == 0, nothing
		// printed yet) is safely retryable; retrying after rows have
		// been emitted would duplicate output. On a TTY we prompt
		// for the password and loop; otherwise we fall through to
		// the reformatter (which points at --password-stdin).
		kind := archive.ClassifyPasswordError(serr)
		if kind != archive.PasswordErrorNone && count == 0 && attempt < maxArchivePasswordPrompts {
			newPw, ok, perr := promptArchivePasswordInteractive(kind, attempt)
			if perr != nil {
				return perr
			}
			if ok {
				pw = newPw
				continue
			}
		}
		return reformatArchiveHTTPErr(serr, rp.OlaresID, "entries", srcWire)
	}

	if !o.jsonOutput {
		// Footer line in tabular mode helps the user tell that the
		// stream completed cleanly vs. was truncated mid-walk. JSON
		// mode skips this to keep `| jq` pipelines clean.
		if total > 0 && total != count {
			fmt.Fprintf(out, "\n(streamed %d entries; server reports %d total)\n", count, total)
		} else {
			fmt.Fprintf(out, "\n(streamed %d entries)\n", count)
		}
	}
	return nil
}

// errArchiveEntriesMaxReached is the sentinel the entries cb returns
// when --max-entries N is set and the walk has reached N. Using a
// named error keeps the abort path distinguishable from real
// failures via errors.Is.
var errArchiveEntriesMaxReached = errors.New("max-entries reached")

// archiveCatOptions captures flags exclusive to `archive cat`.
type archiveCatOptions struct {
	format        string
	node          string
	passwordStdin bool
	output        string
}

// newArchiveCatCommand returns the `files archive cat <archive>
// <inner-path>` sub-command. Streams the bytes of a single archive
// member to stdout (or to a local file via -o/--output).
func newArchiveCatCommand(f *cmdutil.Factory) *cobra.Command {
	o := &archiveCatOptions{}
	cmd := &cobra.Command{
		Use:   "cat <archive-path> <inner-path>",
		Short: "stream a single archive member's bytes to stdout",
		Long: `Stream the raw bytes of a single member of an archive on the per-user
files-backend, without extracting the whole archive first.

Wire shape:

    GET /api/archive/<node>/entry?source=<archive>&path=<inner-path>
    Content-Type: application/octet-stream

The transfer is binary-safe (no buffering, no transformation), so
piping into ` + "`less`" + ` / ` + "`hexdump`" + ` / ` + "`head -c`" + ` works as expected. Pass
-o/--output to write to a local file instead of stdout.

Format detection:

    The server infers the archive container from the source's filename
    extension (.zip / .7z / .tar.gz / .tgz / ...) — no client-side hint
    travels on the wire. We still accept --format locally so the CLI can
    pre-validate flag combinations (e.g. --password-stdin only on zip /
    7z); pass it when the file has no canonical extension.

Read constraints (mirrors the LarePass web app):

    Bare single-stream compressors — bzip2 (.bz2 / .bzip2) and xz (.xz)
    — carry no entry table, so there is no inner member to address by
    path; ` + "`cat`" + ` is rejected up front. Unpack such an archive with
    ` + "`olares-cli files extract`" + ` to get its single decompressed file.
    The tar.* compounds remain fully readable.

Passwords (zip / 7z only):

    Use --password-stdin to read the password from STDIN. The
    server returns HTTP 4xx with a typed code (password_required /
    password_invalid) when authentication fails; the CLI surfaces
    that as a friendly message pointing at --password-stdin.

Examples:

    # Cat a Markdown file out of a zip.
    olares-cli files archive cat drive/Home/Backups/2026-Q1.zip notes.md

    # Save a single binary asset to disk.
    olares-cli files archive cat drive/Home/Vault/data.7z bin/payload \
        --password-stdin -o ./payload < pw.txt

    # Pipe through tools.
    olares-cli files archive cat drive/Home/Backups/2026-Q1.zip logs/today.log | \
        tail -n 50
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runArchiveCat(cmd.Context(), f, cmd.OutOrStdout(), args[0], args[1], o)
		},
	}
	cmd.Flags().StringVar(&o.format, "format", "",
		"archive format (one of: "+archive.JoinFormats()+"); derived from the source extension when omitted")
	cmd.Flags().StringVar(&o.node, "node", "",
		"override the {node} URL segment for /api/archive/<node>/ (defaults to the first /api/nodes/ entry)")
	cmd.Flags().BoolVar(&o.passwordStdin, "password-stdin", false,
		"read the archive password from STDIN (zip / 7z only)")
	cmd.Flags().StringVarP(&o.output, "output", "o", "",
		"write the entry's bytes to this local file path instead of stdout")
	return cmd
}

func runArchiveCat(
	ctx context.Context,
	f *cmdutil.Factory,
	out io.Writer,
	archiveArg, innerPath string,
	o *archiveCatOptions,
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := requireArchiveBackendVersion(ctx, f); err != nil {
		return err
	}
	innerPath = strings.TrimSpace(innerPath)
	if innerPath == "" {
		return errors.New("cat: <inner-path> must not be empty")
	}
	// Allow the user to type a leading '/' but normalise it away —
	// the server stores entry paths without a leading slash, and
	// silently failing on the wire would be a poor UX.
	innerPath = strings.TrimPrefix(innerPath, "/")

	src, srcWire, err := parseArchiveSource(archiveArg, "archive cat")
	if err != nil {
		return err
	}
	if err := requireCommonBackendVersion(ctx, f, isCommonFrontendPath(src.FileType, src.Extend)); err != nil {
		return err
	}

	format := o.format
	if format == "" {
		format = archive.FormatFromExtension(archiveArg)
		if format == "" {
			return fmt.Errorf(
				"cat: cannot derive --format from %q; pass --format (one of: %s)",
				archiveArg, archive.JoinFormats())
		}
	}
	if !archive.IsSupportedFormat(format) {
		return fmt.Errorf("cat: unsupported --format %q; valid formats: %s",
			format, archive.JoinFormats())
	}
	// Bare single-stream compressors (bzip2 / xz) carry no entry
	// table, so there is no inner member to address by path. Refuse
	// the read up front. Mirrors LarePass's
	// unsupportedArchivePreviewExtensions gate. To get the single
	// decompressed payload, extract the archive with
	// `olares-cli files extract` instead.
	if !archive.SupportsPreview(format) {
		return fmt.Errorf(
			"cat: reading members of %q archives is not supported — bzip2 / xz are raw single-stream "+
				"compressors with no addressable entries; unpack it with `olares-cli files extract` instead",
			format)
	}

	password, err := readArchivePasswordStdin(o.passwordStdin)
	if err != nil {
		return err
	}
	if password != "" && !archive.SupportsPassword(format) {
		return fmt.Errorf(
			"cat: --password-stdin is only supported on passwordable formats (zip, 7z); got format %q",
			format)
	}

	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return err
	}
	httpClient, err := f.HTTPClientWithoutTimeout(ctx)
	if err != nil {
		return err
	}
	cli := &archive.Client{
		HTTPClient: httpClient,
		BaseURL:    rp.FilesURL,
	}

	node, err := resolveArchiveNode(ctx, f, rp, []frontendPathLike{src}, o.node)
	if err != nil {
		return err
	}

	if err := preflightArchiveSource(ctx, rp, httpClient, srcWire, "cat"); err != nil {
		return reformatArchiveHTTPErr(err, rp.OlaresID, "cat preflight", srcWire)
	}

	// Pick the destination writer. -o opens a local file (tmp +
	// rename, mirroring download's atomicity guarantee); without it
	// we write straight to `out` (stdout in the common case).
	w, finalize, err := openArchiveCatDestination(o.output, out)
	if err != nil {
		return err
	}
	// Rollback on any early return / stream failure. finalize is
	// single-shot, so once we commit below this deferred call is a
	// no-op and won't clobber the committed (or failed-but-preserved)
	// tmp.
	defer func() { _ = finalize(false) }()

	// The entry endpoint reports a missing / wrong password with an
	// HTTP 4xx BEFORE any bytes are copied into `w` (the status is
	// checked ahead of io.Copy), so the destination writer is still
	// at offset 0 on a password failure and is safe to reuse across
	// the prompt-and-retry loop — mirroring TermiPass's preview
	// password re-prompt.
	var dl archive.EntryDownload
	err = withArchivePasswordRetry(password, func(pw string) error {
		var e error
		dl, e = cli.StreamEntry(ctx, archive.EntryOptions{
			Source: srcWire,
			Path:   innerPath,
			Format: format,
			Node:   node,
		}, pw, w)
		return e
	})
	if err != nil {
		return reformatArchiveHTTPErr(err, rp.OlaresID, "cat", innerPath)
	}

	// Commit the tmp → final rename on success. A rename / close
	// failure here is a local-filesystem error (not an HTTP one) and
	// MUST surface — otherwise we'd print a "wrote ..." line for
	// bytes that never landed at the destination.
	if cerr := finalize(true); cerr != nil {
		return cerr
	}

	if o.output != "" {
		// Only print a status line in -o mode so stdout passthrough
		// stays clean for pipelines.
		fmt.Fprintf(out, "wrote %s (%s) to %s\n",
			innerPath, formatBytes(dl.BytesWritten), o.output)
	}
	return nil
}

// openArchiveCatDestination picks the writer for `archive cat`.
// When `output` is empty (no -o) OR the explicit Unix stdout alias
// "-", we write straight to `stdout` (the cobra command's output
// writer) with a no-op finalize. Otherwise we open `output.tmp`
// and atomically rename it on success.
//
// Returns (writer, finalize, err). `finalize(true)` commits the
// tmp → final rename and RETURNS the rename / close error so the
// caller never reports a write that didn't land on disk;
// `finalize(false)` rolls back by deleting the tmp (so a failed
// stream doesn't leave a half-written file behind).
//
// The finalize is single-shot: the first call (commit or rollback)
// wins and subsequent calls are no-ops. That lets the caller wire
// `defer finalize(false)` for the early-return / stream-failure
// paths AND call `finalize(true)` explicitly on success without the
// deferred rollback clobbering a committed (or a failed-but-
// preserved) tmp. Crucially, on a commit whose rename FAILS we
// leave the tmp in place — the streamed bytes are intact, only the
// rename didn't happen — and surface that path in the error so the
// user can recover instead of silently losing the download.
//
// We don't refuse to overwrite an existing file here: the
// caller's pattern is "I just told you where to put it", same
// shape as `cat > file` in the shell. If a guard is needed, add
// an --overwrite flag later (mirroring download's policy).
func openArchiveCatDestination(output string, stdout io.Writer) (io.Writer, func(commit bool) error, error) {
	if output == "" || output == "-" {
		// Default (no -o) and the explicit "-" alias both stream to
		// the command's stdout — binary-safe passthrough, same
		// behavior as `cat` / `kubectl ... -f -`. No temp file, so
		// finalize is a no-op.
		return stdout, func(bool) error { return nil }, nil
	}
	tmp := output + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, nil, fmt.Errorf("open %s: %w", tmp, err)
	}
	// done guards single-shot semantics: once a commit or rollback
	// has run, later calls (e.g. the deferred rollback after a
	// successful or failed commit) are no-ops so they can't delete a
	// committed file or a failed-but-preserved tmp.
	done := false
	finalize := func(commit bool) error {
		if done {
			return nil
		}
		done = true
		if commit {
			if cerr := f.Close(); cerr != nil {
				// Couldn't flush/close the tmp — the bytes may be
				// incomplete, so don't rename over the destination.
				// Keep the tmp for inspection and report the failure.
				return fmt.Errorf("finalize %s: closing temp %s failed: %w", output, tmp, cerr)
			}
			if rerr := os.Rename(tmp, output); rerr != nil {
				// The stream fully landed in tmp; only the rename
				// failed (cross-device, perms, destination dir gone,
				// …). Preserve tmp so the download isn't lost and
				// surface where it is.
				return fmt.Errorf("finalize %s: rename from temp failed: %w (the downloaded bytes are preserved at %s)",
					output, rerr, tmp)
			}
			return nil
		}
		// Rollback: stream failed before commit — drop the partial tmp.
		_ = f.Close()
		_ = os.Remove(tmp)
		return nil
	}
	return f, finalize, nil
}

// readArchivePasswordStdin reads a single line of password from
// STDIN when --password-stdin is set. We trim the trailing
// newline (so `echo "pw" | ...` works) but preserve embedded
// spaces (a passphrase with spaces is legitimate).
//
// When STDIN is a TTY we print a one-line prompt before reading
// — same pattern as `git`'s credential helper. When STDIN is
// piped (the common case in scripts), we read silently.
//
// If --password-stdin was NOT passed, we return "" without
// touching STDIN.
func readArchivePasswordStdin(enabled bool) (string, error) {
	if !enabled {
		return "", nil
	}
	if term.IsTerminal(int(os.Stdin.Fd())) {
		fmt.Fprint(os.Stderr, "archive password: ")
		// Use term.ReadPassword to suppress echo. Returned bytes
		// don't include the trailing newline.
		pw, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr) // newline after the silent input
		if err != nil {
			return "", fmt.Errorf("read password from terminal: %w", err)
		}
		return string(pw), nil
	}
	// Piped STDIN: read a single line (or the whole stream if no
	// newline). We treat the leading line as the password — this
	// matches `docker login --password-stdin` and is the shape
	// scripts already use.
	br := bufio.NewReader(os.Stdin)
	line, err := br.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("read password from stdin: %w", err)
	}
	return strings.TrimRight(line, "\r\n"), nil
}

// maxArchivePasswordPrompts caps the interactive retry loop so a
// misbehaving server that always answers "password incorrect"
// can't trap the user in an endless prompt. TermiPass's GUI loops
// until the user clicks Cancel; on the CLI an empty entry (or
// Ctrl-D) is the cancel gesture, and this cap is the backstop.
const maxArchivePasswordPrompts = 5

// promptArchivePasswordInteractive asks the terminal for an
// archive password during a retry — the server reported the
// archive is encrypted (PasswordErrorRequired) or that the
// supplied password was wrong (PasswordErrorInvalid). This mirrors
// TermiPass's requestArchivePassword dialog, which re-opens
// whenever isArchivePasswordError fires.
//
// Returns ok=false (and a nil error) when:
//
//   - STDIN is not a TTY — a scripted / piped invocation can't be
//     prompted, so the caller surfaces the original error with a
//     --password-stdin hint instead; or
//   - the user cancels by entering an empty password or hitting
//     Ctrl-D (EOF).
//
// `attempt` is the zero-based retry index, used only to vary the
// lead-in line (first ask vs. subsequent re-asks).
func promptArchivePasswordInteractive(kind archive.PasswordErrorKind, attempt int) (string, bool, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return "", false, nil
	}
	switch {
	case kind == archive.PasswordErrorInvalid:
		fmt.Fprintln(os.Stderr, "archive password is incorrect.")
	case attempt == 0:
		fmt.Fprintln(os.Stderr, "this archive is password-protected.")
	}
	fmt.Fprint(os.Stderr, "enter archive password (empty to cancel): ")
	pw, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr) // newline after the silent input
	if err != nil {
		if errors.Is(err, io.EOF) {
			return "", false, nil // Ctrl-D = cancel
		}
		return "", false, fmt.Errorf("read password from terminal: %w", err)
	}
	s := string(pw)
	if s == "" {
		return "", false, nil // empty entry = cancel
	}
	return s, true, nil
}

// withArchivePasswordRetry runs `attempt(password)` and, when it
// fails with a password-required / password-incorrect error AND
// STDIN is a TTY, re-prompts for a password and retries — the CLI
// analogue of TermiPass's isArchivePasswordError loop.
//
// `initial` is the password parsed from --password-stdin (often
// "" — the user runs extract without one and discovers the archive
// is encrypted only when the server says so).
//
// The loop returns:
//
//   - nil once `attempt` succeeds;
//   - the LAST underlying error when the failure isn't a password
//     problem, when STDIN isn't a TTY, when the user cancels, or
//     when the retry cap is hit. The caller is responsible for
//     reformatting that error (reformatArchiveHTTPErr), which also
//     maps the numeric password codes to a --password-stdin hint
//     for the non-TTY path.
//
// `attempt` MUST be side-effect-free up to the point a password
// error can occur — archive password failures happen before any
// bytes are written / streamed, so the retried call starts from a
// clean slate.
func withArchivePasswordRetry(initial string, attempt func(password string) error) error {
	password := initial
	for i := 0; ; i++ {
		err := attempt(password)
		if err == nil {
			return nil
		}
		kind := archive.ClassifyPasswordError(err)
		if kind == archive.PasswordErrorNone || i >= maxArchivePasswordPrompts {
			return err
		}
		newPw, ok, perr := promptArchivePasswordInteractive(kind, i)
		if perr != nil {
			return perr
		}
		if !ok {
			return err
		}
		password = newPw
	}
}

// frontendPathLike is the small interface needed by
// resolveArchiveNode — the cobra layer's source/destination structs
// expose FileType / Extend, and we don't need anything more.
// Keeping it tiny lets compress.go / extract.go / archive.go all
// reuse the same node resolver without sharing concrete types.
type frontendPathLike interface {
	GetFileType() string
	GetExtend() string
}

// archivePath is a small concrete value type the cobra layer
// uses both to drive resolveArchiveNode and to render the
// canonical wire path. We don't promote it to a top-level type
// because nothing outside this file needs to see it.
type archivePath struct {
	FileType string
	Extend   string
	SubPath  string
}

func (p archivePath) GetFileType() string { return p.FileType }
func (p archivePath) GetExtend() string   { return p.Extend }

// archiveAllowedNamespaces is the explicit allow-list for the four
// archive verbs (compress / extract / archive entries / archive
// cat). Anything outside this set is rejected client-side so the
// user gets an actionable error instead of an opaque 404 / 500
// from the backend (or worse, a silent success that lands the
// archive somewhere unexpected).
//
// The shape mirrors permission.SupportedFileTypes for `chown`:
// each entry is a (fileType, extend, label) tuple. Empty `extend`
// means the fileType is allowed regardless of which volume root
// the path picks (e.g. cache/<any-node>/...).
//
// Allowed:
//
//   - drive/Home/<...>       — the Home volume on the user's PVC
//   - drive/Data/<...>       — the Data volume on the user's PVC
//   - drive/Common/<...>     — the app common data area (JuiceFS
//                              /rootfs/Common); Olares >= 1.12.6
//   - cache/<node>/<...>     — the per-node Cache volume
//   - external/<node>/<...>  — external mounts (USB, SMB, …)
//
// drive/Common joins the allow-list because TermiPass's
// `archiveSupportedDriveTypes` (utils/interface/archive.ts) includes
// DriveType.Common — the LarePass GUI offers compress / extract on the
// common data area just like Home / Data.
//
// Rejected: sync (Seafile libraries — backend doesn't support
// streaming compress / extract today), and every cloud-account
// drive (awss3, dropbox, google, tencent — same reason: cloud
// stores don't share the local-FS path semantics the archive
// endpoints assume).
var archiveAllowedNamespaces = []struct {
	FileType string // drive / cache / external
	Extend   string // empty = any (cache, external); non-empty = exact
	Label    string // human-readable hint for error messages
}{
	{FileType: "drive", Extend: "Home", Label: "drive/Home/<sub>"},
	{FileType: "drive", Extend: "Data", Label: "drive/Data/<sub>"},
	{FileType: "drive", Extend: "Common", Label: "drive/Common/<sub>"},
	{FileType: "cache", Extend: "", Label: "cache/<node>/<sub>"},
	{FileType: "external", Extend: "", Label: "external/<node>/<volume>/<sub>"},
}

// validateArchiveNamespace enforces the archive allow-list. Called
// from every path-parse helper in compress.go / extract.go /
// archive.go so the rejection fires before any wire request.
//
// `verb` is the user-facing operation name ("compress" /
// "extract" / "archive entries" / "archive cat") used in the
// error message. Putting it in the message text helps the user
// recognise which command tripped the guard when they're
// running compound pipelines.
func validateArchiveNamespace(verb, fileType, extend string) error {
	for _, ns := range archiveAllowedNamespaces {
		if ns.FileType != fileType {
			continue
		}
		if ns.Extend == "" || ns.Extend == extend {
			return nil
		}
	}
	return archiveNamespaceError(verb, fileType, extend)
}

// archiveNamespaceError builds the per-namespace rejection
// message. Split out so each unsupported namespace gets a hint
// pointing at the right next step:
//
//   - sync     → Seafile libraries don't go through the archive
//                endpoint; suggest `files repos` for inspection,
//                or staging the content into drive/Home first.
//   - awss3 / dropbox / google / tencent → cloud object stores
//                don't expose the local-FS path semantics archive
//                needs; the LarePass app handles compress/extract
//                in-browser for these.
//   - drive with a non-Home/Data/Common extend (e.g. `drive/Shared`,
//                `drive/Trash`) → out of allow-list explicitly.
//
// The error always lists the full allow-list at the tail so the
// user has a quick reference of what would work.
func archiveNamespaceError(verb, fileType, extend string) error {
	allowed := archiveAllowedNamespacesList()
	switch fileType {
	case "sync":
		return fmt.Errorf(
			"`files %s` does not support the %q namespace today: Seafile libraries are not routed through /api/archive/. "+
				"Stage the content into drive/Home or drive/Data first, or operate on the library via `files repos`. "+
				"Allowed: %s",
			verb, fileType, allowed)
	case "awss3", "dropbox", "google", "tencent":
		return fmt.Errorf(
			"`files %s` does not support the %q cloud-drive namespace: cloud object stores don't expose the local-FS path semantics the archive endpoints need. "+
				"Use the LarePass app for in-browser compress/extract on cloud drives. "+
				"Allowed: %s",
			verb, fileType, allowed)
	case "drive":
		return fmt.Errorf(
			"`files %s` only supports the %q drive volumes Home / Data / Common; got %q/%q. "+
				"Allowed: %s",
			verb, fileType, fileType, extend, allowed)
	}
	return fmt.Errorf(
		"`files %s` does not support the %q namespace. Allowed: %s",
		verb, fileType, allowed)
}

// archiveAllowedNamespacesList renders the allow-list as a
// human-readable string for error messages.
func archiveAllowedNamespacesList() string {
	parts := make([]string, 0, len(archiveAllowedNamespaces))
	for _, ns := range archiveAllowedNamespaces {
		parts = append(parts, ns.Label)
	}
	return strings.Join(parts, ", ")
}

// parseArchiveSource is the canonical "user-facing path string →
// (archivePath, wire path)" converter. Used by extract / entries
// / cat — verbs that take ONE archive-source argument. Compress
// has its own multi-source / single-dst handling in compress.go.
//
// `verb` names the cobra command driving the parse ("extract" /
// "archive entries" / "archive cat") and is threaded into the
// namespace-allow-list error message so the user can tell which
// command tripped the guard.
//
// Validation:
//
//   - The path must parse as a FrontendPath (same parser the
//     other verbs use; rejects empty, malformed, unknown fileType).
//   - The path's namespace must be in archiveAllowedNamespaces —
//     archive endpoints don't support sync / cloud drives today,
//     so reject early with an actionable hint instead of letting
//     the backend surface an opaque error.
//   - The path must NOT be a volume root — archive operations on
//     `drive/Home/` would be a wildly ambiguous "compress my
//     entire Home" or "extract this archive that is the Home
//     root", neither of which is what users mean.
//   - A trailing '/' is allowed BUT we strip it before passing
//     to the wire — the server's entry endpoint doesn't care
//     about the trailing slash on a file path.
func parseArchiveSource(raw, verb string) (archivePath, string, error) {
	fp, err := ParseFrontendPath(raw)
	if err != nil {
		return archivePath{}, "", err
	}
	if err := validateArchiveNamespace(verb, fp.FileType, fp.Extend); err != nil {
		return archivePath{}, "", err
	}
	if strings.Trim(fp.SubPath, "/") == "" {
		return archivePath{}, "", fmt.Errorf(
			"refusing to use the root of %s/%s as an archive source; archive operations require a file path",
			fp.FileType, fp.Extend)
	}
	// Archive source MUST be a file. We can't fully validate that
	// here (would require a Stat), so just refuse the directory
	// marker locally — the preflight Stat catches the
	// "actually-a-dir" case before the wire request.
	if strings.HasSuffix(fp.SubPath, "/") {
		return archivePath{}, "", fmt.Errorf(
			"refusing to use %s as an archive source: trailing '/' marks it as a directory, but archive operations require a file",
			fp.String())
	}
	p := archivePath{
		FileType: fp.FileType,
		Extend:   fp.Extend,
		SubPath:  fp.SubPath,
	}
	return p, archive.BuildWirePath(p.FileType, p.Extend, p.SubPath), nil
}

// resolveArchiveNode mirrors the cp package's ResolveNode
// cascade for the archive endpoints:
//
//	flagNode (--node) → wins outright
//	→ first External/Cache extend among the supplied paths
//	→ first /api/nodes/ entry
//
// The cascade differs slightly from cp: archive doesn't have a
// "destination" concept for entries / cat (single-source verbs),
// and compress / extract pass BOTH source and destination so
// the helper accepts a slice of paths to check.
//
// External / Cache paths short-circuit the /api/nodes/ fetch,
// matching the cp optimisation — a drive-only call doesn't need
// the round-trip when --node is also unset.
func resolveArchiveNode(
	ctx context.Context,
	f *cmdutil.Factory,
	rp *credential.ResolvedProfile,
	paths []frontendPathLike,
	flagNode string,
) (string, error) {
	if flagNode != "" {
		return flagNode, nil
	}
	for _, p := range paths {
		if isPasteMultiNode(p.GetFileType()) && p.GetExtend() != "" {
			return p.GetExtend(), nil
		}
	}
	// Default-node fetch via /api/nodes/. We reach for the
	// download.Client here because its HTTPClient (the factory-
	// provided refreshing transport) is the same one the archive
	// verbs will use for the actual call — the round-trip stays
	// inside the auth-aware transport. We could declare a typed
	// helper inside the archive package but the cp / upload
	// packages already prove the GetNodes endpoint is small enough
	// to inline here.
	httpClient, err := f.HTTPClient(ctx)
	if err != nil {
		return "", err
	}
	endpoint := rp.FilesURL + "/api/nodes/"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("GET %s: %w", endpoint, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return "", &download.HTTPError{
			Status: resp.StatusCode,
			Body:   string(body),
			URL:    endpoint,
			Method: http.MethodGet,
		}
	}
	var env struct {
		Data struct {
			Nodes []struct {
				Name   string `json:"name"`
				Master bool   `json:"master"`
			} `json:"nodes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return "", fmt.Errorf("decode /api/nodes/: %w", err)
	}
	if len(env.Data.Nodes) == 0 {
		return "", errors.New("/api/nodes/ returned no nodes; pass --node to override")
	}
	// Prefer the master node when one is explicitly flagged — the
	// task queue lives on the master and routing there avoids a
	// cross-node hop. Falls back to the first node when no master
	// is flagged (older deployments).
	for _, n := range env.Data.Nodes {
		if n.Master && n.Name != "" {
			return n.Name, nil
		}
	}
	if env.Data.Nodes[0].Name == "" {
		return "", errors.New("/api/nodes/ returned a node with empty name; cannot resolve default {node}")
	}
	return env.Data.Nodes[0].Name, nil
}

// preflightArchiveSource Stats a wire-shape source path BEFORE
// any state-changing or stream-opening call goes out. Used by
// extract / entries / cat: all three need the archive file to
// EXIST and to be a file (not a directory).
//
// Volume roots get a free pass via download.Stat's synthetic
// dir record; archive sources should always reach this with a
// real leaf path because parseArchiveSource refuses volume
// roots.
//
// We pass the same HTTPClient archive uses so the preflight
// inherits the refreshing-transport's 401/403 retry.
func preflightArchiveSource(
	ctx context.Context,
	rp *credential.ResolvedProfile,
	httpClient *http.Client,
	wirePath, op string,
) error {
	statClient := &download.Client{
		HTTPClient: httpClient,
		BaseURL:    rp.FilesURL,
	}
	// Drop the leading '/' before handing to Stat — Stat expects
	// the plain `<fileType>/<extend>/<sub>` shape, same as cp's
	// preflight.
	plain := strings.TrimPrefix(wirePath, "/")
	info, err := statClient.Stat(ctx, plain)
	if err != nil {
		if download.IsNotFound(err) {
			return fmt.Errorf("%s: archive %s does not exist on the server", op, plain)
		}
		return err
	}
	if info.IsDir {
		return fmt.Errorf(
			"%s: %s is a directory on the server, not a file; archive operations require an archive file",
			op, plain)
	}
	return nil
}

// reformatArchiveHTTPErr is the shared error reformatter for the
// archive verbs. Same shape as reformatCpHTTPErr (cp/mv) and
// reformatHTTPErr (download/cat) — branch on credential errors
// first (preserves the "profile login" CTA), then on typed
// HTTPError statuses, then on the entries-stream / entry-error
// typed errors that only apply to this verb group.
func reformatArchiveHTTPErr(err error, olaresID, op, target string) error {
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

	// In-band NDJSON stream errors carry an enum code we can
	// translate into actionable CTAs. We branch on them before
	// the generic HTTPError mapping because the typed error is
	// the more specific diagnosis.
	if se, ok := archive.IsEntriesStreamError(err); ok {
		return formatArchiveEntriesStreamError(se, op, target)
	}

	// Single-entry endpoint errors wrap an *HTTPError but add the
	// `code` enum on top. errors.As fires on both — branch on
	// EntryError first so the friendlier CTA wins.
	var entryErr *archive.EntryError
	if errors.As(err, &entryErr) {
		return formatArchiveEntryError(entryErr, op, target)
	}

	// Numeric-coded password errors from the POST compress /
	// extract endpoints (and entries' pre-walk JSON 4xx). We reach
	// this point only after the interactive retry loop declined to
	// run (STDIN isn't a TTY) or the user cancelled, so point them
	// at --password-stdin — the scriptable way to supply it.
	switch archive.ClassifyPasswordError(err) {
	case archive.PasswordErrorRequired:
		return fmt.Errorf("%s %s: archive requires a password; pass it via --password-stdin", op, target)
	case archive.PasswordErrorInvalid:
		return fmt.Errorf("%s %s: archive password is incorrect; supply the right one via --password-stdin", op, target)
	}

	// Generic HTTPError mapping mirrors reformatCpHTTPErr.
	var status int
	var url string
	var aErr *archive.HTTPError
	if errors.As(err, &aErr) {
		status = aErr.Status
		url = aErr.URL
	}
	if status == 0 {
		var dlErr *download.HTTPError
		if errors.As(err, &dlErr) {
			status = dlErr.Status
			url = dlErr.URL
		}
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
	_ = url // future use: include in diagnostic line when --verbose lands
	return err
}

// formatArchiveEntriesStreamError maps an in-band NDJSON error
// onto a friendly cobra-layer message. The enum codes match
// the spec (password_required / password_invalid /
// archive_corrupt / volume_missing / canceled / not_found /
// internal).
func formatArchiveEntriesStreamError(e *archive.EntriesStreamError, op, target string) error {
	switch e.Code {
	case archive.CodePasswordRequired:
		return fmt.Errorf("%s %s: archive requires a password; pass --password-stdin", op, target)
	case archive.CodePasswordInvalid:
		return fmt.Errorf("%s %s: archive password is incorrect; re-run with --password-stdin and the right password", op, target)
	case archive.CodeArchiveCorrupt:
		return fmt.Errorf("%s %s: archive appears to be corrupt or truncated (server reports: %s)",
			op, target, e.Message)
	case archive.CodeVolumeMissing:
		return fmt.Errorf("%s %s: a multi-volume archive's part is missing on the server (server reports: %s); "+
			"ensure all .z01 / .z02 / ... files are uploaded next to the main archive",
			op, target, e.Message)
	case archive.CodeNotFound:
		return fmt.Errorf("%s %s: not found on the server (server reports: %s)", op, target, e.Message)
	case archive.CodeCanceled:
		return fmt.Errorf("%s %s: server cancelled the walk (server reports: %s)", op, target, e.Message)
	}
	// Unknown code — fall through to the raw error.
	return e
}

// formatArchiveEntryError maps an EntryError (HTTPError +
// typed code) onto a friendly cobra-layer message. The enum
// values are mostly shared with the entries stream; entry adds
// `entry_too_large` (HTTP 413) for hosts that cap single-shot
// reads.
func formatArchiveEntryError(e *archive.EntryError, op, target string) error {
	if e.Code == "" {
		return e
	}
	switch e.Code {
	case archive.CodePasswordRequired:
		return fmt.Errorf("%s %s: archive requires a password; pass --password-stdin", op, target)
	case archive.CodePasswordInvalid:
		return fmt.Errorf("%s %s: archive password is incorrect; re-run with --password-stdin and the right password", op, target)
	case archive.CodeNotFound:
		return fmt.Errorf("%s %s: entry not found inside the archive (use `files archive entries` to list its members)",
			op, target)
	case archive.CodeArchiveCorrupt:
		return fmt.Errorf("%s %s: archive appears to be corrupt (server reports: %s)", op, target, e.Message)
	case archive.CodeVolumeMissing:
		return fmt.Errorf("%s %s: a multi-volume archive part is missing on the server (server reports: %s); "+
			"ensure all .z01 / .z02 / ... files are uploaded next to the main archive",
			op, target, e.Message)
	case archive.CodeEntryTooLarge:
		return fmt.Errorf("%s %s: entry exceeds the server's single-shot read limit (HTTP 413; server reports: %s); "+
			"extract the whole archive with `files extract` instead",
			op, target, e.Message)
	}
	return e
}
