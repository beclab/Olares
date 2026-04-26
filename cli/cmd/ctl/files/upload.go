package files

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/beclab/Olares/cli/internal/files/upload"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

type uploadOptions struct {
	parallel   int
	chunkSize  int64
	maxRetries int
	node       string
}

// NewUploadCommand: `olares-cli files upload <local> <remote>`
//
// Pushes a single file or a whole directory tree from the local
// filesystem into Drive/Home on the per-user files-backend, using the
// same chunked-resumable protocol the LarePass web app speaks
// (Resumable.js + the Drive v2 endpoints under /upload/upload-link,
// /upload/file-uploaded-bytes, /api/resources/drive/Home/...). See
// internal/files/upload/uploader.go for the wire-level details.
//
// Resume is enabled by default and is server-driven: before each file
// the CLI calls /upload/file-uploaded-bytes/<node>/ to ask "how much do
// you already have?", floors that to a chunk boundary, and resumes
// from there. There's no local progress file — re-running the same
// command after a Ctrl-C just re-asks the server, which is robust
// against any state drift between client invocations.
//
// File-level concurrency: --parallel N runs N files concurrently
// through an errgroup. Within a single file, chunks are sent
// sequentially (matching the web app's simultaneousUploads=1 default);
// pipelining chunks per file is not implemented because the
// resume-probe + chunk-sequence assumes a single in-flight chunk per
// file at a time.
//
// Path schema for <remote>: same as `files ls`, but the upload target
// must live under drive/Home (drive/Data is read-only on the wire).
// Trailing '/' on <remote> is significant — it's how we distinguish
// "upload into this directory" from "upload as this exact path
// (rename)" for the single-file case.
func NewUploadCommand(f *cmdutil.Factory) *cobra.Command {
	o := &uploadOptions{}
	cmd := &cobra.Command{
		Use:   "upload <local-path> <remote-path>",
		Short: "upload a file or directory to Drive/Home with resumable chunks",
		Long: `Upload a local file or directory to drive/Home/<...> on the per-user files-backend.

The chunked / resumable protocol mirrors the LarePass web app: each
file is probed against /upload/file-uploaded-bytes/ to figure out the
resume offset, chunks are POSTed (default 8 MiB each) until the file
is complete, and per-chunk failures are retried with backoff. Re-run
the same command after a Ctrl-C and the upload picks up where the
server stopped accepting bytes.

<remote> uses the same 3-segment frontend path as ` + "`olares-cli files ls`" + ` —
the upload target must be under drive/Home (drive/Data is read-only on
the wire). A trailing '/' on <remote> means "upload into this
directory"; without one, <remote> is treated as the full target path
(useful to rename a file on the way in).

The destination directory MUST already exist on the server. The
files-backend's "create directory" call auto-renames on collision
(POST .../Documents/ on an existing Documents creates "Documents (1)"
instead of returning a conflict), so we don't try to pre-create it —
use the web app or a future ` + "`files mkdir`" + ` verb if you need to
materialize a new directory first.

Examples:

    # Upload one file into a directory.
    olares-cli files upload report.pdf drive/Home/Documents/

    # Same, but rename to 2026-Q1.pdf on the server.
    olares-cli files upload report.pdf drive/Home/Documents/2026-Q1.pdf

    # Upload a directory tree (preserves the source folder name).
    olares-cli files upload ./photos drive/Home/Backups/

    # Two files in flight at a time.
    olares-cli files upload ./photos drive/Home/Backups/ --parallel 2
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpload(cmd.Context(), f, cmd.OutOrStdout(), args[0], args[1], o)
		},
	}
	cmd.Flags().IntVar(&o.parallel, "parallel", 2,
		"number of files to upload concurrently (per-file chunks remain sequential)")
	cmd.Flags().Int64Var(&o.chunkSize, "chunk-size", upload.DefaultChunkSize,
		"chunk size in bytes (default 8 MiB; should match the server's expected size)")
	cmd.Flags().IntVar(&o.maxRetries, "max-retries", upload.DefaultMaxRetries,
		"maximum retry attempts per chunk on transient failures")
	cmd.Flags().StringVar(&o.node, "node", "",
		"override the upload node name (defaults to the first node from /api/nodes/)")
	return cmd
}

func runUpload(
	ctx context.Context,
	f *cmdutil.Factory,
	out io.Writer,
	localPath, remotePath string,
	o *uploadOptions,
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if o.parallel < 1 {
		o.parallel = 1
	}

	// Parse remote path with the same parser `files ls` uses, then
	// enforce the upload-only constraint (drive/Home only). The chunk
	// pipeline simply has no path that points at drive/Data on the
	// wire — see apps/packages/app/src/api/files/v2/drive/utils.ts
	// (driveCommonUrl always emits /drive/Home).
	fp, err := ParseFrontendPath(remotePath)
	if err != nil {
		return err
	}
	if fp.FileType != "drive" || fp.Extend != "Home" {
		return fmt.Errorf("upload destination must be under drive/Home (got %q)", fp.String())
	}
	// SubPath always starts with '/' from the parser; strip the leading
	// slash so BuildPlan sees a relative form like "Documents/Backups/".
	remoteSub := strings.TrimPrefix(fp.SubPath, "/")

	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return err
	}

	// Build a dedicated *http.Client for the upload session: same
	// X-Authorization injection convention as the rest of the CLI, but
	// without the factory's 30s overall timeout — an 8 MiB chunk on a
	// slow link can easily exceed that, and we'd rather fail via
	// context cancellation than via http.Client.Timeout (the latter
	// truncates the request body mid-flight, which leaves the server
	// in an inconsistent state).
	httpClient := newUploadHTTPClient(rp.InsecureSkipVerify)
	client := &upload.Client{
		HTTPClient:  httpClient,
		BaseURL:     rp.FilesURL,
		AccessToken: rp.AccessToken,
	}

	node := o.node
	if node == "" {
		nodes, err := client.FetchNodes(ctx)
		if err != nil {
			return fmt.Errorf("fetch upload nodes: %w", err)
		}
		// Mirrors the web app's getUploadNode() — first node wins. A
		// future iteration can pick by master flag or by --node, but
		// the web app's default is good enough for the common case.
		node = nodes[0].Name
		if node == "" {
			return fmt.Errorf("upload node returned by /api/nodes/ has empty name")
		}
	}

	plan, err := upload.BuildPlan(localPath, remoteSub)
	if err != nil {
		return err
	}

	// Why we DON'T pre-mkdir the destination root or any source-tree
	// directory here:
	//
	//   - POST /api/resources/.../<existing-dir>/ does NOT return 409 on
	//     a name collision; the files-backend silently auto-renames to
	//     "<existing-dir> (1)" and creates an empty new directory next
	//     to the original. That's surprising for an idempotent "ensure
	//     this dir exists" operation, and the original report from
	//     hitting this exact bug was a stray "Documents (1)" appearing
	//     on the server even though the file landed in the real
	//     "Documents" via the chunk POST.
	//   - The chunk POST routes by parent_dir + relative_path and the
	//     server transparently creates intermediate directories on the
	//     way (it's how folder upload from the web app works). So the
	//     destination root MUST already exist (matching the web app,
	//     which can only upload to a directory the user already
	//     navigated into via the file picker), and source-tree dirs
	//     are auto-created by the file uploads.
	//
	// Empty subdirectories of a folder upload still surface in
	// plan.EmptyDirs for diagnostic / future-flag use, but we
	// deliberately don't mkdir them by default — same behavior as the
	// browser folder picker (which can't deliver empty directories
	// either). Surface a one-line note so the user isn't surprised
	// that empty dirs disappeared.
	if len(plan.EmptyDirs) > 0 {
		fmt.Fprintf(out, "note: skipping %d empty subdirector%s (matches web app behavior; pass files instead if needed)\n",
			len(plan.EmptyDirs),
			pluralYies(len(plan.EmptyDirs)),
		)
	}

	if len(plan.Files) == 0 {
		fmt.Fprintf(out, "no files to upload (empty source or directories only)\n")
		return nil
	}

	// Plan summary first — gives the user something to look at while
	// the first probe is in flight (which can take a second on a cold
	// connection). The summary is one line so it doesn't crowd the
	// per-file progress lines that follow.
	totalBytes := int64(0)
	for _, ft := range plan.Files {
		totalBytes += ft.Size
	}
	fmt.Fprintf(out, "uploading %d file(s), %s, into %s (parallel=%d, chunk=%s)\n",
		len(plan.Files), formatBytes(totalBytes), plan.ParentDir,
		o.parallel, formatBytes(o.chunkSize))

	return runUploads(ctx, client, plan, node, o, out)
}

// runUploads schedules per-file UploadFile calls through an errgroup
// of `o.parallel` workers. Per-file failures cancel the group's
// context, so an unrecoverable error in one file aborts the rest of
// the batch quickly (otherwise the user would have to wait for every
// remaining file to also fail before getting their shell back).
func runUploads(
	ctx context.Context,
	client *upload.Client,
	plan *upload.Plan,
	node string,
	o *uploadOptions,
	out io.Writer,
) error {
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
			opts := plan.ToUploadOpts(task, node, o.chunkSize, o.maxRetries)

			start := time.Now()
			fmt.Fprintf(out, "  → %s (%s)\n", task.RelativePath, formatBytes(task.Size))

			var lastReported int64
			progress := func(uploaded, total int64) {
				// Throttle progress lines so a 50-chunk file doesn't
				// emit 50 log lines: only print when crossing 25%
				// boundaries. The final 100% line is always emitted
				// because uploaded==total exactly.
				if total <= 0 {
					return
				}
				step := total / 4
				if step <= 0 {
					step = 1
				}
				if uploaded == total || uploaded-lastReported >= step {
					lastReported = uploaded
					fmt.Fprintf(out, "    %s: %d/%d (%s/%s)\n",
						task.RelativePath, uploaded, total,
						formatBytes(uploaded), formatBytes(total))
				}
			}
			if err := client.UploadFile(gctx, opts, progress); err != nil {
				return fmt.Errorf("%s: %w", task.RelativePath, err)
			}

			mu.Lock()
			completed++
			bytesDone += task.Size
			done := completed
			mu.Unlock()
			fmt.Fprintf(out, "  ✓ %s (%s, %s) [%d/%d]\n",
				task.RelativePath, formatBytes(task.Size),
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

// newUploadHTTPClient builds a *http.Client suitable for streaming
// chunks: TLS verification follows the active profile, NO overall
// Timeout (we rely on context cancellation), and explicit keep-alive +
// HTTP/2 from http.DefaultTransport.
//
// The X-Authorization header is injected per-request inside upload.Client
// (rather than via a transport wrapper) so the same Client can talk to
// httptest servers in tests without dragging the access token into the
// fixture surface.
func newUploadHTTPClient(insecureSkipVerify bool) *http.Client {
	base := http.DefaultTransport.(*http.Transport).Clone()
	if insecureSkipVerify {
		base.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec G402 -- explicit profile opt-in
	}
	return &http.Client{
		// Timeout: 0 — no overall request timeout. Big chunks on slow
		// links would otherwise truncate mid-POST. Cancellation flows
		// through context (Ctrl-C / parent ctx).
		Transport: base,
	}
}

// pluralYies turns 1 → "y" and any other number → "ies", so the user-
// facing "1 empty subdirectory" / "2 empty subdirectories" message
// reads naturally without a separate string for the singular case.
func pluralYies(n int) string {
	if n == 1 {
		return "y"
	}
	return "ies"
}
