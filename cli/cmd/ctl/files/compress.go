package files

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/internal/files/archive"
	"github.com/beclab/Olares/cli/internal/files/download"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// compressOptions holds every flag the compress verb accepts.
// Shared with no other verb — compress's option surface is
// distinct enough from extract's (level / volumeSizeMB don't apply
// to extract) that a single struct per verb is clearer than a
// merged one with "extract-only / compress-only" guards.
type compressOptions struct {
	format           string
	level            int
	volumeSizeMB     int    // explicit MiB (legacy/precise knob)
	volumeSize       string // unit-aware: "100MB" / "1.5GB" / bare MiB
	preserveSymlinks bool
	conflict         string
	passwordStdin    bool
	node             string
	wait             bool
	pollInterval     int // seconds; 0 = client default
}

// NewCompressCommand: `olares-cli files compress <src>... <dst>`.
//
// Builds one archive out of one or more remote sources, on the
// per-user files-backend's `/api/archive/<node>/compress` endpoint.
// The wire call returns one task_id and the actual byte writing
// happens on the server's task queue (same async shape as
// `files cp` / `files mv`); pass --wait to block until the task
// reaches a terminal status.
//
// CLI verb shape mirrors `cp` for the source / destination
// positional args:
//
//	compress <src>... <dst>
//
// Where:
//
//   - <src>... are the existing remote paths to archive (files or
//     directories; directories are recursively included).
//   - <dst> is the new archive's remote path (must NOT exist;
//     conflict policy decides what happens if it does).
//
// Format detection:
//
//   - --format wins outright.
//   - Otherwise the destination's filename suffix is inspected via
//     archive.FormatFromExtension (`.zip` → "zip", `.tar.gz` →
//     "tar.gz", ...).
//   - If both fail, the command refuses to run; the user has to
//     pick a format explicitly.
//
// Password handling: --password-stdin reads from STDIN. The
// password is sent via `X-Archive-Password` (zip / 7z only); other
// formats refuse the flag client-side.
func NewCompressCommand(f *cmdutil.Factory) *cobra.Command {
	o := &compressOptions{
		level: -1, // sentinel for "unset — let the backend decide"
	}
	cmd := &cobra.Command{
		Use:   "compress <src>... <dst>",
		Short: "compress one or more remote entries into a new archive",
		Long: `Pack one or more entries from the per-user files-backend into a
single archive file. The actual byte-writing happens asynchronously on
the server's task queue.

Wire shape (one POST per invocation):

    POST /api/archive/<node>/compress
        body: {sources, destination, format, level?, volumeSizeMB?,
               preserveSymlinks, conflict}
        headers: X-Archive-Password (zip / 7z only)

The wire endpoint returns one task_id; pass --wait to poll until
the task finishes.

Source / destination paths use the same 3-segment frontend path as
` + "`olares-cli files ls`" + ` (e.g. ` + "`drive/Home/Documents/`" + `,
` + "`sync/<repo_id>/notes/`" + `). Each <src> may be a file or a
directory. Directories are recursively included.

Supported formats:

    zip, 7z, tar, tar.gz, tgz, tar.bz2, tar.xz, gzip, bzip2, xz

The format is derived from <dst>'s filename suffix when --format
is omitted (.zip / .7z / .tar.gz / .tgz / .tar.bz2 / .tar.xz /
.tar / .gz / .gzip / .bz2 / .bzip2 / .xz). Pass --format when the destination has
no canonical suffix.

Format constraints (mirrors the LarePass web app):

    - Passwords / split volumes: zip and 7z only.
    - Single-file compressors (gzip / bzip2 / xz) wrap exactly ONE
      file's bytes — they cannot pack a directory or multiple
      sources. To compress a directory or several files, use a
      container format (zip, 7z, tar, tar.gz, tar.bz2, tar.xz, tgz).
      The tar.* compounds tar first, then compress, so they carry
      directories and multiple members just fine.

Knobs (most apply only to specific formats):

    --level N           Compression level 0..9 (codec-defined; 0 =
                        store, 9 = max). Omit to use the backend's
                        default.

    --volume-size SIZE  Split-archive volume size with a unit
                        suffix: MB / GB (e.g. 100MB, 1.5GB).
                        A bare number is MiB. Rounded up,
                        floored at 1 MiB. zip / 7z only. Preferred
                        over --volume-size-mb.

    --volume-size-mb M  Split-archive volume size in raw MiB (zip /
                        7z only). Back-compat alias for --volume-size;
                        pass at most one of the two.

    --preserve-symlinks Archive symlinks as symlinks instead of
                        dereferencing them at compress time.

    --conflict POLICY   On-collision policy at <dst>:
                        rename (default) / overwrite / skip.

    --password-stdin    Read the archive password from STDIN (zip
                        / 7z only). Avoids leaking through shell
                        history or ` + "`ps`" + `. For 7z this also enables
                        header encryption (mhe=on server-side).

    --wait              Block until the task reaches a terminal
                        status, printing periodic progress lines.

    --node              Override the {node} URL segment. Defaults
                        to the first /api/nodes/ entry, with the
                        External/Cache hint applied (same cascade
                        as ` + "`files cp`" + `).

Preflight existence check:

    Every <src> is Stat'd before the POST goes out. A missing
    source or wrong file-vs-dir intent aborts the operation
    cleanly before the task hits the queue. Same fail-fast spirit
    as ` + "`files cp`" + ` / ` + "`files mv`" + `.

Examples:

    # Two files into a zip.
    olares-cli files compress drive/Home/a.pdf drive/Home/b.pdf \
        drive/Home/out.zip

    # Whole directory into a tar.gz at max compression.
    olares-cli files compress drive/Home/Photos/ \
        drive/Home/photos.tar.gz --level 9

    # Encrypted 7z with header encryption.
    echo "s3cret" | olares-cli files compress \
        drive/Home/Secrets/ drive/Home/secrets.7z --password-stdin

    # Split-volume zip (100 MiB volumes; unit-aware flag).
    olares-cli files compress drive/Home/Backups/ \
        drive/Home/backup.zip --volume-size 100MB

    # Same, with a fractional GiB.
    olares-cli files compress drive/Home/Backups/ \
        drive/Home/backup.zip --volume-size 1.5GB

    # Block until the task completes.
    olares-cli files compress drive/Home/Reports/ \
        drive/Home/reports.zip --wait
`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCompress(cmd.Context(), f, cmd.OutOrStdout(), args, o)
		},
	}
	cmd.Flags().StringVar(&o.format, "format", "",
		"archive format (one of: "+archive.JoinFormats()+"); derived from <dst>'s extension when omitted")
	cmd.Flags().IntVar(&o.level, "level", -1,
		"compression level 0..9 (0=store, 9=max); leave unset to use the backend's codec default")
	cmd.Flags().IntVar(&o.volumeSizeMB, "volume-size-mb", 0,
		"split-archive volume size in MiB (zip / 7z only; 0 = single volume). Prefer --volume-size for unit suffixes")
	cmd.Flags().StringVar(&o.volumeSize, "volume-size", "",
		"split-archive volume size with a unit suffix: MB / GB (e.g. 100MB, 1.5GB; bare number = MiB). zip / 7z only")
	cmd.Flags().BoolVar(&o.preserveSymlinks, "preserve-symlinks", false,
		"archive symlinks as symlinks (default: dereference at compress time)")
	cmd.Flags().StringVar(&o.conflict, "conflict", string(archive.ConflictDefault),
		"on-collision policy at <dst>: rename (default) / overwrite / skip")
	cmd.Flags().BoolVar(&o.passwordStdin, "password-stdin", false,
		"read the archive password from STDIN (zip / 7z only); avoids leaking through shell history")
	cmd.Flags().StringVar(&o.node, "node", "",
		"override the {node} URL segment for /api/archive/<node>/ (defaults to the first /api/nodes/ entry)")
	cmd.Flags().BoolVar(&o.wait, "wait", false,
		"block until the compress task finishes, printing periodic progress updates")
	cmd.Flags().IntVar(&o.pollInterval, "poll-interval", 0,
		"task-status poll interval in seconds when --wait is set (0 = client default ~2s)")
	return cmd
}

func runCompress(
	ctx context.Context,
	f *cmdutil.Factory,
	out io.Writer,
	args []string,
	o *compressOptions,
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := requireArchiveBackendVersion(ctx, f); err != nil {
		return err
	}
	if len(args) < 2 {
		// cobra's MinimumNArgs catches this earlier; guard the
		// runner so a future programmatic call can't slip through.
		return errors.New("compress: need at least one <src> and one <dst>")
	}

	srcArgs := args[:len(args)-1]
	dstArg := args[len(args)-1]

	srcs, srcWires, err := parseCompressSources(srcArgs)
	if err != nil {
		return err
	}
	dst, dstWire, err := parseCompressDestination(dstArg)
	if err != nil {
		return err
	}
	touchesCommon := isCommonFrontendPath(dst.FileType, dst.Extend)
	for _, s := range srcs {
		if isCommonFrontendPath(s.FileType, s.Extend) {
			touchesCommon = true
		}
	}
	if err := requireCommonBackendVersion(ctx, f, touchesCommon); err != nil {
		return err
	}

	conflict, err := archive.ParseConflict(o.conflict)
	if err != nil {
		return fmt.Errorf("compress: %w", err)
	}
	if err := archive.ValidateLevel(o.level); err != nil {
		return fmt.Errorf("compress: %w", err)
	}

	format := o.format
	if format == "" {
		format = archive.FormatFromExtension(dstArg)
		if format == "" {
			return fmt.Errorf(
				"compress: cannot derive --format from destination %q; pass --format (one of: %s)",
				dstArg, archive.JoinFormats())
		}
	}
	if err := archive.ValidateFormat(format, "compress"); err != nil {
		return err
	}

	// Single-file compressors (gzip / bzip2 / xz) wrap exactly one
	// file's bytes — they cannot pack multiple sources. Reject the
	// multi-source case up front (the directory case needs a Stat,
	// so it is enforced in the preflight below). Mirrors LarePass's
	// canCompressItemsWithFormat gate.
	if err := archive.ValidateSingleFileCompression(format, len(srcWires), false); err != nil {
		return fmt.Errorf("compress: %w", err)
	}

	password, err := readArchivePasswordStdin(o.passwordStdin)
	if err != nil {
		return err
	}
	if password != "" && !archive.SupportsPassword(format) {
		return fmt.Errorf(
			"compress: --password-stdin is only supported on passwordable formats (zip, 7z); got format %q",
			format)
	}
	// Resolve the split-volume size from the two mutually-exclusive
	// flags. --volume-size (unit-aware) is the preferred knob;
	// --volume-size-mb (raw MiB) stays for back-compat. Passing both
	// is a usage error — refuse rather than silently pick one.
	volumeMB, err := resolveVolumeSizeMB(o.volumeSize, o.volumeSizeMB)
	if err != nil {
		return fmt.Errorf("compress: %w", err)
	}
	if volumeMB > 0 && !archive.SupportsMultiVolume(format) {
		return fmt.Errorf(
			"compress: split volumes are only supported on multi-volume formats (zip, 7z); got format %q",
			format)
	}

	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return err
	}
	httpClient, err := f.HTTPClient(ctx)
	if err != nil {
		return err
	}
	cli := &archive.Client{
		HTTPClient: httpClient,
		BaseURL:    rp.FilesURL,
	}

	// Node cascade: srcs + dst share the same set so the External
	// / Cache hint can be drawn from any of them. Mirrors cp's
	// dst-wins-over-src cascade by listing dst first.
	all := make([]frontendPathLike, 0, len(srcs)+1)
	all = append(all, dst)
	for _, s := range srcs {
		all = append(all, s)
	}
	node, err := resolveArchiveNode(ctx, f, rp, all, o.node)
	if err != nil {
		return err
	}

	// Preflight: every <src> must exist, AND the dst's parent
	// directory must exist (so the archive can land there). The
	// dst itself MUST NOT exist — but the backend's conflict
	// policy handles that, so we don't probe it here (and don't
	// want to: a "rename" conflict policy is the user's "I
	// expect a collision is fine" declaration).
	statClient := &download.Client{
		HTTPClient: httpClient,
		BaseURL:    rp.FilesURL,
	}
	if err := preflightCompress(ctx, statClient, format, srcs, srcWires, dst, dstWire); err != nil {
		return reformatArchiveHTTPErr(err, rp.OlaresID, "compress preflight", "")
	}

	fmt.Fprintf(out, "compress %d source%s → %s (format=%s, node=%s):\n",
		len(srcWires), pluralEs(len(srcWires)), dstWire, format, node)
	for _, s := range srcWires {
		fmt.Fprintf(out, "  - %s\n", s)
	}

	taskID, err := cli.Compress(ctx, archive.CompressOptions{
		Sources:          srcWires,
		Destination:      dstWire,
		Format:           format,
		Level:            o.level,
		VolumeSizeMB:     volumeMB,
		PreserveSymlinks: o.preserveSymlinks,
		Conflict:         conflict,
		Node:             node,
	}, password)
	if err != nil {
		return reformatArchiveHTTPErr(err, rp.OlaresID, "compress", dstWire)
	}

	fmt.Fprintf(out, "queued compress task: %s\n", taskID)
	if !o.wait {
		fmt.Fprintf(out, "(pass --wait to block until completion; "+
			"manage it with `olares-cli files task {cancel,pause,resume} %s --node %s`)\n",
			taskID, node)
		return nil
	}

	return waitArchiveTask(ctx, cli, node, taskID, o.pollInterval, out, "compress")
}

// resolveVolumeSizeMB reconciles the two split-volume flags into a
// single MiB value for the wire `volumeSizeMB` field:
//
//   - --volume-size (unit-aware string) takes precedence and is the
//     documented knob (parsed via archive.ParseVolumeSize so
//     100MB / 1.5GB all work, rounding up, floored at 1).
//   - --volume-size-mb (raw int MiB) stays for back-compat.
//   - Supplying BOTH is rejected — there is no sensible merge and a
//     silent winner would surprise the user.
//   - Neither set → 0 (single volume).
func resolveVolumeSizeMB(volumeSize string, volumeSizeMB int) (int, error) {
	trimmed := strings.TrimSpace(volumeSize)
	if trimmed != "" && volumeSizeMB > 0 {
		return 0, errors.New("pass either --volume-size or --volume-size-mb, not both")
	}
	if trimmed != "" {
		mib, err := archive.ParseVolumeSize(trimmed)
		if err != nil {
			return 0, err
		}
		return mib, nil
	}
	if volumeSizeMB < 0 {
		return 0, fmt.Errorf("--volume-size-mb must not be negative (got %d)", volumeSizeMB)
	}
	return volumeSizeMB, nil
}

// parseCompressSources converts the N-1 leading args of the
// compress argv into per-source archivePath / wire pairs. The
// shapes are loose by design — both file and directory paths are
// legal (a directory is recursively included); the preflight
// catches typos.
//
// Each source must also be in the archive allow-list (Home /
// Data / Cache / External). Reject sync / cloud drives early so
// the user gets an actionable error instead of an opaque backend
// failure halfway through the task.
func parseCompressSources(srcArgs []string) ([]archivePath, []string, error) {
	if len(srcArgs) == 0 {
		return nil, nil, errors.New("compress: at least one <src> is required")
	}
	srcs := make([]archivePath, 0, len(srcArgs))
	wires := make([]string, 0, len(srcArgs))
	for _, a := range srcArgs {
		fp, err := ParseFrontendPath(a)
		if err != nil {
			return nil, nil, err
		}
		if err := validateArchiveNamespace("compress", fp.FileType, fp.Extend); err != nil {
			return nil, nil, err
		}
		if strings.Trim(fp.SubPath, "/") == "" {
			return nil, nil, fmt.Errorf(
				"refusing to use the root of %s/%s as a compress source; "+
					"specify a real file or directory path",
				fp.FileType, fp.Extend)
		}
		p := archivePath{
			FileType: fp.FileType,
			Extend:   fp.Extend,
			SubPath:  fp.SubPath,
		}
		srcs = append(srcs, p)
		wires = append(wires, archive.BuildWirePath(p.FileType, p.Extend, p.SubPath))
	}
	return srcs, wires, nil
}

// parseCompressDestination converts the trailing arg of the
// compress argv into the destination's archivePath + wire path.
// A trailing '/' is refused: the destination MUST be a file
// (an archive file). The format-from-extension heuristic also
// requires a non-trailing-slash basename to work.
//
// Also enforces the archive-namespace allow-list (Home / Data /
// Cache / External). Writing the new archive into sync or a
// cloud drive is not currently supported by the backend.
func parseCompressDestination(raw string) (archivePath, string, error) {
	fp, err := ParseFrontendPath(raw)
	if err != nil {
		return archivePath{}, "", err
	}
	if err := validateArchiveNamespace("compress", fp.FileType, fp.Extend); err != nil {
		return archivePath{}, "", err
	}
	if strings.Trim(fp.SubPath, "/") == "" {
		return archivePath{}, "", fmt.Errorf(
			"refusing to use the root of %s/%s as a compress destination; "+
				"point at a filename, e.g. %s/%s/archive.zip",
			fp.FileType, fp.Extend, fp.FileType, fp.Extend)
	}
	if strings.HasSuffix(fp.SubPath, "/") {
		return archivePath{}, "", fmt.Errorf(
			"refusing to use %s as a compress destination: trailing '/' marks it as a directory, "+
				"but the destination must be a file (the archive itself)",
			fp.String())
	}
	p := archivePath{
		FileType: fp.FileType,
		Extend:   fp.Extend,
		SubPath:  fp.SubPath,
	}
	return p, archive.BuildWirePath(p.FileType, p.Extend, p.SubPath), nil
}

// preflightCompress verifies every <src> exists on the server
// and refuses up front when a source's trailing '/' intent
// disagrees with its server-side file/dir kind (same shape as
// preflightCpMv). The destination ITSELF is not probed —
// conflict policy handles that — but its PARENT directory must
// exist so the new archive file has a home.
//
// statClient is the download.Client that the caller wires from
// the factory-provided HTTP client. Reusing the same transport
// means the preflight inherits the refreshing-transport's
// 401/403 retry, the same way `cp` does.
func preflightCompress(
	ctx context.Context,
	statClient *download.Client,
	format string,
	srcs []archivePath,
	srcWires []string,
	dst archivePath,
	dstWire string,
) error {
	for i, s := range srcs {
		// One canonical string for BOTH the wire path passed to
		// download.Stat AND the human-facing error display (same
		// pattern as preflightCpMv's intentionally-merged
		// plain / display var).
		plain := s.FileType + "/" + s.Extend + s.SubPath
		info, err := statClient.Stat(ctx, plain)
		if err != nil {
			if download.IsNotFound(err) {
				return fmt.Errorf("compress: source %s does not exist on the server", plain)
			}
			return err
		}
		// Single-file compressors (gzip / bzip2 / xz) cannot pack a
		// directory. Now that the Stat told us the real kind, reject
		// a directory source with the format-specific message (it
		// takes priority over the generic trailing-slash hint
		// below). Mirrors LarePass's canCompressItemsWithFormat.
		if info.IsDir {
			if err := archive.ValidateSingleFileCompression(format, len(srcs), true); err != nil {
				return fmt.Errorf("compress: source %s: %w", plain, err)
			}
		}
		// Product policy: refuse "single source file that is already an
		// archive/compressed file" to avoid accidental double-compression.
		// Intentionally scoped to the single-file case only — when the
		// user compresses multiple sources, we skip this format check.
		if detectedFormat, reject := detectCompressedSingleSource(len(srcs), info.IsDir, plain, info.Name); reject {
			return fmt.Errorf(
				"compress: source %s appears to already be a compressed/archive file (detected format %q from filename %q); "+
					"single-file compression of an already compressed file is blocked",
				plain, detectedFormat, info.Name)
		}
		srcHasTrailingSlash := strings.HasSuffix(s.SubPath, "/")
		if srcHasTrailingSlash && !info.IsDir {
			return fmt.Errorf(
				"compress: source %s is a file on the server, not a directory; drop the trailing '/'",
				plain)
		}
		if !srcHasTrailingSlash && info.IsDir {
			return fmt.Errorf(
				"compress: source %s is a directory on the server; add a trailing '/' to confirm directory intent (it will be recursively archived)",
				plain)
		}
		_ = srcWires[i] // surfaced earlier in the caller's plan log
	}

	// Destination parent directory must exist; without it the
	// task either fails server-side or trips the auto-rename
	// quirk on a path that doesn't yet exist. Same reasoning as
	// the "destination directory must exist" leg of preflightCpMv.
	parentPlain := dst.FileType + "/" + dst.Extend + parentSubPath(dst.SubPath)
	info, err := statClient.Stat(ctx, parentPlain)
	if err != nil {
		if download.IsNotFound(err) {
			return fmt.Errorf(
				"compress: destination's parent directory %s does not exist on the server; "+
					"create it first with `olares-cli files mkdir`",
				parentPlain)
		}
		return err
	}
	if !info.IsDir {
		return fmt.Errorf(
			"compress: destination's parent %s is a file on the server, not a directory",
			parentPlain)
	}
	_ = dstWire // already in the caller's plan log
	return nil
}

// detectCompressedSingleSource reports whether the source should be
// rejected by the "single-file already compressed" gate:
//   - applies ONLY when sourceCount == 1
//   - applies ONLY to files (not directories)
//   - detects archive/compressed formats by filename suffix via
//     archive.FormatFromExtension (zip / 7z / tar.* / gzip / bzip2 / xz,
//     including split main-part names like .zip.001 / .7z.001).
//
// Returns the detected format when reject=true.
func detectCompressedSingleSource(sourceCount int, isDir bool, plainPath, statName string) (string, bool) {
	if sourceCount != 1 || isDir {
		return "", false
	}
	if f := archive.FormatFromExtension(statName); f != "" {
		return f, true
	}
	if f := archive.FormatFromExtension(plainPath); f != "" {
		return f, true
	}
	return "", false
}

// waitArchiveTask blocks on the supplied task_id, printing
// periodic progress lines to `out`. Shared between compress and
// extract — the wire shape is identical (the task queue is a
// per-node service, not per-verb).
//
// `verb` is the human label ("compress" / "extract") used in
// the status line so the user can tell which task they're
// waiting on. `pollSeconds` is the cobra-layer's flag value
// in seconds; 0 means "use the archive package default
// (~2 s)".
//
// The progress callback is throttled to one line per poll —
// the server reports `progress` as 0..100, so spamming the
// terminal between two consecutive 5%-jumps would be noisy
// without adding information.
func waitArchiveTask(
	ctx context.Context,
	cli *archive.Client,
	node, taskID string,
	pollSeconds int,
	out io.Writer,
	verb string,
) error {
	interval := time.Duration(pollSeconds) * time.Second
	if interval <= 0 {
		interval = archive.DefaultTaskPollInterval
	}
	var lastProgress float64 = -1
	var lastStatus string
	err := cli.WaitTask(ctx, node, taskID, interval, func(u archive.TaskUpdate) {
		// Suppress duplicate lines: only print when the
		// status flips OR the progress moved by at least 1 %.
		// This keeps the terminal stable on bursty servers
		// without losing significant signal.
		changed := u.Status != lastStatus || (u.Progress-lastProgress) >= 1.0 || u.Progress < lastProgress
		if !changed {
			return
		}
		lastStatus = u.Status
		lastProgress = u.Progress
		phase := ""
		if u.TotalPhase > 0 {
			phase = fmt.Sprintf(" (phase %d/%d)", u.CurrentPhase, u.TotalPhase)
		}
		fmt.Fprintf(out, "  %s task %s%s: %s %.1f%%\n",
			verb, taskID, phase, u.Status, u.Progress)
	})
	if err != nil {
		// A cancelled context means the user hit Ctrl-C (or the
		// parent context was cancelled). The local poll stops here,
		// but the SERVER-SIDE task keeps running — detaching the
		// poll does not cancel the queue entry. Point the user at
		// the explicit cancel verb so the half-built archive doesn't
		// linger. (Mirrors the behavior the SKILL doc warns about.)
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			fmt.Fprintf(out,
				"\nstopped waiting on %s task %s — the task is STILL RUNNING on the server.\n"+
					"  resume waiting : olares-cli files task ... (re-run with --wait)\n"+
					"  cancel it      : olares-cli files task cancel %s --node %s\n",
				verb, taskID, taskID, node)
		}
		return err
	}
	fmt.Fprintf(out, "%s task %s completed\n", verb, taskID)
	return nil
}

