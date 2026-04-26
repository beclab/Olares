package files

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/internal/files/download"
)

type downloadOptions struct {
	parallel   int
	maxRetries int
	overwrite  bool
	resume     bool
}

// NewDownloadCommand: `olares-cli files download <remote> [<local>]`
//
// Pulls a single file or a whole directory tree from the per-user
// files-backend down to the local filesystem. Single-file downloads
// resume via the server's native `Range: bytes=N-` support
// (raw_service.go's parseRangeHeader); directories walk recursively
// over /api/resources and pull each file with the same code path.
//
// Local destination semantics:
//
//   - omitted        → ./<basename(remote)> in the current directory
//   - existing dir   → write under that directory using the remote
//     basename (mirrors `cp`'s behavior)
//   - any other path → treated as the full local target path
//     (file mode), or the directory to create / use as the root
//     (directory mode).
//
// Concurrency only kicks in for directory mode; --parallel N runs N
// file downloads in flight at once. Per-file resume + retry are
// independent of --parallel.
func NewDownloadCommand(f *cmdutil.Factory) *cobra.Command {
	o := &downloadOptions{}
	cmd := &cobra.Command{
		Use:   "download <remote-path> [<local-path>]",
		Short: "download a file or directory from the per-user files-backend",
		Long: `Download a file or directory tree from the per-user files-backend.

Single-file resume is server-driven: pass --resume and the CLI sends
Range: bytes=<localSize>- so the server only ships the bytes you don't
already have. The local file is opened with O_APPEND so a Ctrl-C +
re-run keeps making forward progress without sidecar progress files.

Without --resume or --overwrite, the command refuses to clobber an
existing local file. Use --overwrite to replace it (writes to
<dst>.tmp + rename, so the previous version stays intact until the
new one lands), or --resume to continue a previously-interrupted
download.

Directory downloads recursively walk /api/resources, recreate the
remote directory tree under the local destination (the remote root's
own basename becomes the top-level directory there, matching the
LarePass folder-download UX), and run --parallel N file fetches
concurrently. Empty subdirectories are mirrored locally so the on-disk
tree matches even when a directory has no files.

<remote> uses the same 3-segment frontend path as ` + "`olares-cli files ls`" + `.
A trailing '/' on <remote> means "treat as directory" (validated
against the server's actual type via /api/resources stat); without
one, the path is treated as a file.

Examples:

    # Download one file into the current directory.
    olares-cli files download drive/Home/Documents/report.pdf

    # Same, but pick a different local name.
    olares-cli files download drive/Home/Documents/report.pdf ./Q1.pdf

    # Resume an interrupted download.
    olares-cli files download drive/Home/Backups/big.tar ./big.tar --resume

    # Recursively pull a folder, 4 files at a time.
    olares-cli files download drive/Home/Documents/ ./out/ --parallel 4
`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			localArg := ""
			if len(args) == 2 {
				localArg = args[1]
			}
			return runDownload(cmd.Context(), f, cmd.OutOrStdout(), args[0], localArg, o)
		},
	}
	cmd.Flags().IntVar(&o.parallel, "parallel", 4,
		"number of files to download concurrently in directory mode")
	cmd.Flags().IntVar(&o.maxRetries, "max-retries", download.DefaultMaxRetries,
		"maximum retry attempts per file on transient failures")
	cmd.Flags().BoolVar(&o.overwrite, "overwrite", false,
		"replace existing local files (writes to <dst>.tmp + rename)")
	cmd.Flags().BoolVar(&o.resume, "resume", false,
		"resume an interrupted download via the server's Range support")
	return cmd
}

func runDownload(
	ctx context.Context,
	f *cmdutil.Factory,
	out io.Writer,
	remoteArg, localArg string,
	o *downloadOptions,
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if o.parallel < 1 {
		o.parallel = 1
	}
	if o.overwrite && o.resume {
		return errors.New("--overwrite and --resume are mutually exclusive")
	}

	fp, err := ParseFrontendPath(remoteArg)
	if err != nil {
		return err
	}

	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return err
	}

	httpClient := newUploadHTTPClient(rp.InsecureSkipVerify)
	client := &download.Client{
		HTTPClient:  httpClient,
		BaseURL:     rp.FilesURL,
		AccessToken: rp.AccessToken,
	}

	// Stat first so we (a) reject auth errors / 404s with a clean
	// message before we touch the local filesystem, and (b) know
	// whether to take the single-file or recursive directory branch.
	plain := strings.TrimSuffix(fp.String(), "/")
	st, err := client.Stat(ctx, plain)
	if err != nil {
		return reformatHTTPErr(err, rp.OlaresID, "stat", plain)
	}

	if st.IsDir {
		return runDownloadDir(ctx, client, fp, plain, localArg, o, out)
	}
	return runDownloadFile(ctx, client, fp, plain, st.Size, localArg, o, out)
}

// runDownloadFile handles the single-file branch. The local destination
// is resolved here so the helper is easy to test in isolation; it's
// the only place we synthesise the implicit "<basename> in cwd" /
// "into existing dir" rules.
func runDownloadFile(
	ctx context.Context,
	client *download.Client,
	fp FrontendPath,
	plain string,
	remoteSize int64,
	localArg string,
	o *downloadOptions,
	out io.Writer,
) error {
	remoteBase := lastSegmentOfFrontendPath(fp)
	if remoteBase == "" {
		return fmt.Errorf("cannot derive a local filename from remote %q (no trailing path component)", fp.String())
	}
	dst, err := resolveLocalFile(localArg, remoteBase)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "downloading %s (%s) → %s\n", fp.String(), formatBytes(remoteSize), dst)

	start := time.Now()
	written, err := client.DownloadFile(ctx, plain, dst, download.Options{
		Overwrite:  o.overwrite,
		Resume:     o.resume,
		MaxRetries: o.maxRetries,
	}, nil)
	if err != nil {
		return reformatHTTPErr(err, "", "download", plain)
	}
	fmt.Fprintf(out, "done: wrote %s in %s (file size %s)\n",
		formatBytes(written),
		time.Since(start).Truncate(time.Millisecond),
		formatBytes(remoteSize),
	)
	return nil
}

// runDownloadDir handles the recursive directory branch: walk the
// remote tree, recreate it locally, run errgroup-bounded parallel file
// downloads. Uses the same per-file Options as the single-file branch
// so --resume / --overwrite have consistent semantics regardless of
// mode.
func runDownloadDir(
	ctx context.Context,
	client *download.Client,
	fp FrontendPath,
	plain string,
	localArg string,
	o *downloadOptions,
	out io.Writer,
) error {
	if localArg == "" {
		// "Into the current directory" — the recreated remote root
		// becomes ./<basename(remote)>/.
		localArg = "."
	}
	plan, err := download.BuildPlan(ctx, client, plain, localArg)
	if err != nil {
		return reformatHTTPErr(err, "", "list", plain)
	}

	// Pre-create the local root + every empty subdirectory so the
	// on-disk tree mirrors the remote one. Doing this before any
	// downloads start means concurrent file writes never race on
	// MkdirAll for shared parents.
	if err := os.MkdirAll(plan.LocalRoot, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", plan.LocalRoot, err)
	}
	for _, ed := range plan.EmptyDirs {
		full := filepath.Join(plan.LocalRoot, filepath.FromSlash(ed))
		if err := os.MkdirAll(full, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", full, err)
		}
	}

	if len(plan.Files) == 0 {
		fmt.Fprintf(out, "no files to download (remote tree has no regular files)\n")
		return nil
	}

	totalBytes := int64(0)
	for _, t := range plan.Files {
		totalBytes += t.Size
	}
	fmt.Fprintf(out, "downloading %d file(s), %s, into %s (parallel=%d)\n",
		len(plan.Files), formatBytes(totalBytes), plan.LocalRoot, o.parallel)

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(o.parallel)
	var (
		mu        sync.Mutex
		completed int
		bytesDone int64
	)
	totalFiles := len(plan.Files)

	for _, task := range plan.Files {
		task := task
		g.Go(func() error {
			start := time.Now()
			fmt.Fprintf(out, "  → %s (%s)\n", task.RelativePath, formatBytes(task.Size))
			written, err := client.DownloadFile(gctx, task.RemotePlainPath, task.LocalPath, download.Options{
				Overwrite:  o.overwrite,
				Resume:     o.resume,
				MaxRetries: o.maxRetries,
			}, nil)
			if err != nil {
				return fmt.Errorf("%s: %w", task.RelativePath, err)
			}
			mu.Lock()
			completed++
			bytesDone += written
			done := completed
			mu.Unlock()
			fmt.Fprintf(out, "  ✓ %s (%s, %s) [%d/%d]\n",
				task.RelativePath, formatBytes(written),
				time.Since(start).Truncate(time.Millisecond),
				done, totalFiles)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}
	fmt.Fprintf(out, "done: %d file(s), %s\n", completed, formatBytes(bytesDone))
	return nil
}

// lastSegmentOfFrontendPath returns the basename of the resource the
// path points at — the last non-empty '/'-split segment of SubPath, or
// the Extend value when SubPath is just "/" (which means "the root of
// this volume", whose effective name on the local side is the volume
// name like "Home").
func lastSegmentOfFrontendPath(fp FrontendPath) string {
	sub := strings.Trim(fp.SubPath, "/")
	if sub == "" {
		return fp.Extend
	}
	if idx := strings.LastIndex(sub, "/"); idx >= 0 {
		return sub[idx+1:]
	}
	return sub
}

// resolveLocalFile applies the implicit local-destination rules for
// the single-file download path:
//
//   - empty localArg     → ./<remoteBase>
//   - localArg is a dir  → <localArg>/<remoteBase>
//   - any other localArg → use as the full target path
//
// Returned path may not yet exist; the downloader's
// planLocalWrite handles the existence + overwrite/resume policy.
func resolveLocalFile(localArg, remoteBase string) (string, error) {
	if localArg == "" {
		return remoteBase, nil
	}
	st, err := os.Stat(localArg)
	switch {
	case err == nil && st.IsDir():
		return filepath.Join(localArg, remoteBase), nil
	case err == nil:
		return localArg, nil
	case errors.Is(err, os.ErrNotExist):
		// Trailing slash means "treat as directory even if it doesn't
		// exist yet" — same convention as `cp` / `rsync`.
		if strings.HasSuffix(localArg, string(os.PathSeparator)) || strings.HasSuffix(localArg, "/") {
			return filepath.Join(localArg, remoteBase), nil
		}
		return localArg, nil
	default:
		return "", fmt.Errorf("stat %s: %w", localArg, err)
	}
}

// reformatHTTPErr maps download.HTTPError codes onto user-friendly
// messages, mirroring the formatHTTPError helper in ls.go. We don't
// share the helper directly because the download package's HTTPError
// type isn't compatible with the upload package's, and untyped
// duck-typing here would be more confusing than the small duplication.
func reformatHTTPErr(err error, olaresID, op, target string) error {
	if err == nil {
		return nil
	}
	var hErr *download.HTTPError
	if errors.As(err, &hErr) {
		switch hErr.Status {
		case 401, 403:
			if olaresID != "" {
				return fmt.Errorf("server rejected the access token (HTTP %d); please run: olares-cli profile login --olares-id %s",
					hErr.Status, olaresID)
			}
			return fmt.Errorf("server rejected the access token (HTTP %d); please re-run `olares-cli profile login`", hErr.Status)
		case 404:
			return fmt.Errorf("%s %s: not found on the server (HTTP 404)", op, target)
		}
	}
	return err
}
