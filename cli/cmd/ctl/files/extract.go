package files

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/internal/files/archive"
	"github.com/beclab/Olares/cli/internal/files/download"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// extractOptions holds the flags the extract verb accepts. The
// surface is intentionally narrower than compress: extract has
// no level / volume-size / per-format dial — it's "given this
// archive, undo the compress into this directory".
type extractOptions struct {
	format           string
	preserveSymlinks bool
	conflict         string
	passwordStdin    bool
	node             string
	wait             bool
	pollInterval     int
}

// NewExtractCommand: `olares-cli files extract <archive> <dst-dir>/`.
//
// Decompresses one remote archive into one remote directory via
// the per-user files-backend's `/api/archive/<node>/extract`
// endpoint. The wire call returns one task_id; pass --wait to
// block until the extraction finishes.
//
// Destination semantics:
//
//   - <dst-dir>/ MUST end with '/'. Mirrors the cp/mv "drop-into-
//     directory" rule, makes intent unambiguous in script files,
//     and avoids the "extract drive/Home/x.zip drive/Home/x"
//     foot-shot of accidentally creating a file with the same
//     name as the archive's stem.
//   - The destination directory itself can either already exist
//     (the conflict policy decides what happens to colliding
//     entries inside it) or be created on-the-fly by the
//     backend's writer. The preflight Stats the parent so the
//     mkdir is rooted in real soil, same shape as the compress
//     preflight.
//
// Format detection mirrors compress: derive from the source's
// extension via archive.FormatFromExtension; --format wins
// outright when set.
func NewExtractCommand(f *cmdutil.Factory) *cobra.Command {
	o := &extractOptions{}
	cmd := &cobra.Command{
		Use:   "extract <archive> <dst-dir>/",
		Short: "extract a remote archive into a remote directory",
		Long: `Unpack a remote archive on the per-user files-backend into a remote
directory. The actual writing happens asynchronously on the server's
task queue.

Wire shape:

    POST /api/archive/<node>/extract
        body: {source, destination, format, preserveSymlinks, conflict}
        headers: X-Archive-Password (zip / 7z only)

Returns one task_id; pass --wait to block until completion.

Source / destination paths use the same 3-segment frontend path as
` + "`olares-cli files ls`" + ` (e.g. ` + "`drive/Home/Backups/2026-Q1.zip`" + `,
` + "`sync/<repo_id>/unpacked/`" + `). The destination MUST end with '/'
(drop-into-directory mode).

Supported formats:

    zip, 7z, tar, tar.gz, tgz, tar.bz2, tar.xz, gzip, bzip2, xz

The format is derived from <archive>'s filename suffix when
--format is omitted (.zip / .7z / .zip.001 / .7z.001 / .tar.gz / .tgz / ...). Pass
--format when the archive has no canonical suffix.

Knobs:

    --preserve-symlinks   Land symlinks inside the archive as
                          symlinks on disk (default: dereference).
    --conflict POLICY     On-collision policy at the destination
                          for entries that already exist:
                          rename (default) / overwrite / skip.
    --password-stdin      Read the password from STDIN (zip / 7z).
    --wait                Block until the task reaches a terminal
                          status.
    --node                Override the {node} URL segment.

Examples:

    # Unpack a zip into a sibling directory.
    olares-cli files extract drive/Home/Backups/2026-Q1.zip \
        drive/Home/Backups/2026-Q1/

    # Encrypted 7z with --wait.
    echo "s3cret" | olares-cli files extract \
        drive/Home/Vault/data.7z drive/Home/Vault/unpacked/ \
        --password-stdin --wait

    # Overwrite on collision.
    olares-cli files extract drive/Home/Backups/2026-Q1.zip \
        drive/Home/Backups/2026-Q1/ --conflict overwrite
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExtract(cmd.Context(), f, cmd.OutOrStdout(), args[0], args[1], o)
		},
	}
	cmd.Flags().StringVar(&o.format, "format", "",
		"archive format (one of: "+archive.JoinFormats()+"); derived from <archive>'s extension when omitted")
	cmd.Flags().BoolVar(&o.preserveSymlinks, "preserve-symlinks", false,
		"land symlinks inside the archive as symlinks on disk (default: dereference)")
	cmd.Flags().StringVar(&o.conflict, "conflict", string(archive.ConflictDefault),
		"on-collision policy at the destination: rename (default) / overwrite / skip")
	cmd.Flags().BoolVar(&o.passwordStdin, "password-stdin", false,
		"read the archive password from STDIN (zip / 7z only)")
	cmd.Flags().StringVar(&o.node, "node", "",
		"override the {node} URL segment for /api/archive/<node>/ (defaults to the first /api/nodes/ entry)")
	cmd.Flags().BoolVar(&o.wait, "wait", false,
		"block until the extract task finishes, printing periodic progress updates")
	cmd.Flags().IntVar(&o.pollInterval, "poll-interval", 0,
		"task-status poll interval in seconds when --wait is set (0 = client default ~2s)")
	return cmd
}

func runExtract(
	ctx context.Context,
	f *cmdutil.Factory,
	out io.Writer,
	archiveArg, dstArg string,
	o *extractOptions,
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := requireArchiveBackendVersion(ctx, f); err != nil {
		return err
	}

	src, srcWire, err := parseArchiveSource(archiveArg, "extract")
	if err != nil {
		return err
	}
	dst, dstWire, err := parseExtractDestination(dstArg)
	if err != nil {
		return err
	}
	if err := requireCommonBackendVersion(ctx, f,
		isCommonFrontendPath(src.FileType, src.Extend) || isCommonFrontendPath(dst.FileType, dst.Extend)); err != nil {
		return err
	}

	conflict, err := archive.ParseConflict(o.conflict)
	if err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	format := o.format
	if format == "" {
		format = archive.FormatFromExtension(archiveArg)
		if format == "" {
			return fmt.Errorf(
				"extract: cannot derive --format from %q; pass --format (one of: %s)",
				archiveArg, archive.JoinFormats())
		}
	}
	if err := archive.ValidateFormat(format, "extract"); err != nil {
		return err
	}

	password, err := readArchivePasswordStdin(o.passwordStdin)
	if err != nil {
		return err
	}
	if password != "" && !archive.SupportsPassword(format) {
		return fmt.Errorf(
			"extract: --password-stdin is only supported on passwordable formats (zip, 7z); got format %q",
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

	node, err := resolveArchiveNode(ctx, f, rp, []frontendPathLike{dst, src}, o.node)
	if err != nil {
		return err
	}

	// Preflight: the archive must exist, and the destination
	// directory's "parent of the leaf" must exist. We don't
	// require the destination directory itself to exist — the
	// extract writer will mkdir it — but the parent must be real.
	statClient := &download.Client{
		HTTPClient: httpClient,
		BaseURL:    rp.FilesURL,
	}
	if err := preflightExtract(ctx, statClient, srcWire, dst); err != nil {
		return reformatArchiveHTTPErr(err, rp.OlaresID, "extract preflight", "")
	}

	fmt.Fprintf(out, "extract %s → %s (format=%s, node=%s)\n",
		srcWire, dstWire, format, node)

	// Wrap the queue call in the password-retry loop: an encrypted
	// archive surfaces as HTTP 400 code 30001 (required) / 30002
	// (incorrect) BEFORE the task is queued, so we can prompt for a
	// password on a TTY and retry — mirroring TermiPass's
	// isArchivePasswordError flow.
	var taskID string
	err = withArchivePasswordRetry(password, func(pw string) error {
		var e error
		taskID, e = cli.Extract(ctx, archive.ExtractOptions{
			Source:           srcWire,
			Destination:      dstWire,
			Format:           format,
			PreserveSymlinks: o.preserveSymlinks,
			Conflict:         conflict,
			Node:             node,
		}, pw)
		return e
	})
	if err != nil {
		return reformatArchiveHTTPErr(err, rp.OlaresID, "extract", srcWire)
	}

	fmt.Fprintf(out, "queued extract task: %s\n", taskID)
	if !o.wait {
		fmt.Fprintf(out, "(pass --wait to block until completion; "+
			"manage it with `olares-cli files task {cancel,pause,resume} %s --node %s`)\n",
			taskID, node)
		return nil
	}

	return waitArchiveTask(ctx, cli, node, taskID, o.pollInterval, out, "extract")
}

// parseExtractDestination converts the destination arg into the
// canonical archivePath / wire pair. The destination MUST end
// with '/' — drop-into-directory mode, mirroring cp / mv. This
// makes intent unambiguous in script files and avoids the easy
// foot-shot of `extract X.zip Y` accidentally creating a file
// `Y` containing the first archive entry.
//
// Also enforces the archive-namespace allow-list (Home / Data /
// Cache / External only); extracting INTO sync or a cloud drive
// is not currently supported by the backend.
func parseExtractDestination(raw string) (archivePath, string, error) {
	fp, err := ParseFrontendPath(raw)
	if err != nil {
		return archivePath{}, "", err
	}
	if err := validateArchiveNamespace("extract", fp.FileType, fp.Extend); err != nil {
		return archivePath{}, "", err
	}
	if !strings.HasSuffix(fp.SubPath, "/") {
		return archivePath{}, "", fmt.Errorf(
			"refusing to use %s as an extract destination: must end with '/' to declare directory intent "+
				"(e.g. %s/)",
			fp.String(), strings.TrimSuffix(fp.String(), "/"))
	}
	p := archivePath{
		FileType: fp.FileType,
		Extend:   fp.Extend,
		SubPath:  fp.SubPath,
	}
	return p, archive.BuildWirePath(p.FileType, p.Extend, p.SubPath), nil
}

// preflightExtract probes the source archive and the destination
// directory's parent before the wire call goes out. The source
// MUST exist and MUST be a file; the destination's parent MUST
// exist as a directory (the extract writer mkdirs the leaf).
//
// Splitting the parent-exists guard from the destination-itself
// guard reflects the backend's contract: the writer auto-creates
// the leaf directory when it doesn't exist, so requiring it
// upfront would refuse the legitimate "extract into a new
// sibling" workflow. Same split as preflightCpMv's exact-target
// branch.
func preflightExtract(
	ctx context.Context,
	statClient *download.Client,
	srcWire string,
	dst archivePath,
) error {
	plain := strings.TrimPrefix(srcWire, "/")
	info, err := statClient.Stat(ctx, plain)
	if err != nil {
		if download.IsNotFound(err) {
			return fmt.Errorf("extract: archive %s does not exist on the server", plain)
		}
		return err
	}
	if info.IsDir {
		return fmt.Errorf(
			"extract: %s is a directory on the server, not a file; extract requires an archive file",
			plain)
	}

	// Destination parent. We stat the dir-of-the-leaf because
	// the leaf itself may not exist yet (the writer creates it).
	parentPlain := dst.FileType + "/" + dst.Extend + parentSubPath(dst.SubPath)
	info, err = statClient.Stat(ctx, parentPlain)
	if err != nil {
		if download.IsNotFound(err) {
			return fmt.Errorf(
				"extract: destination's parent directory %s does not exist on the server; "+
					"create it first with `olares-cli files mkdir`",
				parentPlain)
		}
		return err
	}
	if !info.IsDir {
		return fmt.Errorf(
			"extract: destination's parent %s is a file on the server, not a directory",
			parentPlain)
	}
	return nil
}

