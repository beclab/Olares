package files

import (
	"context"
	"fmt"
	"io"
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
// filesystem into one of the supported per-user files-backend
// namespaces (drive/Home, drive/Data, sync/<repo_id>, cache/<node>,
// external/<node>/<volume>, awss3/<account>, google/<account>, or
// dropbox/<account>) using the same chunked-resumable protocol the
// LarePass web app speaks (Resumable.js + the Drive v2 endpoints
// under /upload/upload-link, /upload/file-uploaded-bytes,
// /api/resources/...). See internal/files/upload/uploader.go for the
// wire-level details.
//
// awss3 / google / dropbox share the chunk pipeline + resume probe
// with Drive (the web app's Awss3DataAPI / GoogleDataAPI /
// DropboxDataAPI all extend DriveDataAPI **without** overriding
// getFileServerUploadLink or getFileUploadedBytes — see
// apps/packages/app/src/api/files/v2/{awss3,google,dropbox}/data.ts),
// so stage 1 is byte-identical to Drive. They're a TWO-STAGE upload
// though: stage 1 only delivers bytes to the Olares files-backend's
// staging area; stage 2 is a server-side "Olares-staging → cloud
// bucket" transfer task that the backend queues and the client must
// wait on. The taskId for stage 2 is returned in the FINAL chunk's
// response body, and we drive the polling via
// upload.Client.WaitCloudTask (see runUploads). This mirrors
// resumejs.ts onFileUploadSuccess L591-606, where the web app's
// Taskmanager.addTask consumes the same taskId and runs the same
// poll loop. Tencent is the lone holdout (see
// TencentDataAPI.getFileServerUploadLink, which posts to
// /drive/create_direct_upload_task and uploads via the octet
// /drive/direct_upload_file flow); we don't speak that protocol yet,
// so this verb rejects it explicitly with a self-describing error.
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
// must live under one of:
//
//   - drive/Home/<sub>
//   - drive/Data/<sub>
//   - sync/<repo_id>/<sub>
//   - cache/<node>/<sub>
//   - external/<node>/<volume>/<sub>
//   - awss3/<account>/<bucket>/<sub>
//   - google/<account>/<sub>
//   - dropbox/<account>/<sub>
//
// Tencent (`tencent/<account>/<sub>`) uses a different upload
// protocol (POST /drive/create_direct_upload_task → octet upload via
// /drive/direct_upload_file/<task_id>) that this verb does not yet
// implement; it's rejected with a self-describing error. Trailing
// '/' on <remote> is significant — it's how we distinguish "upload
// into this directory" from "upload as this exact path (rename)" for
// the single-file case.
//
// Node selection cascade (mirrors `files cp`):
//
//   - --node <name>: explicit override, wins over everything else.
//   - cache/<node>/... or external/<node>/...: <node> from the path.
//   - everything else (drive/sync/awss3/google/dropbox): first node
//     returned by /api/nodes/.
func NewUploadCommand(f *cmdutil.Factory) *cobra.Command {
	o := &uploadOptions{}
	cmd := &cobra.Command{
		Use:   "upload <local-path> <remote-path>",
		Short: "upload a file or directory to drive, sync, cache, external, or a cloud drive (awss3/google/dropbox) with resumable chunks",
		Long: `Upload a local file or directory into one of the supported per-user
files-backend namespaces:

    drive/Home/<sub>                 (your Olares Home volume)
    drive/Data/<sub>                 (your Olares Data volume)
    sync/<repo_id>/<sub>             (a Seafile sync library)
    cache/<node>/<sub>               (node-local cache)
    external/<node>/<volume>/<sub>   (an attached external volume)
    awss3/<account>/<bucket>/<sub>   (S3-compatible cloud drive)
    google/<account>/<sub>           (Google Drive)
    dropbox/<account>/<sub>          (Dropbox)

Tencent COS (` + "`tencent/<account>/...`" + `) uses a different upload
protocol (octet upload via /drive/direct_upload_file/<task_id>) that
this verb does not yet implement; it is rejected up-front. The
remaining cloud-drive namespaces share the same chunked-resumable
multipart-POST protocol as Drive, so they're handled by the regular
upload pipeline below.

The chunked / resumable protocol mirrors the LarePass web app: each
file is probed against /upload/file-uploaded-bytes/ to figure out the
resume offset, chunks are POSTed (default 8 MiB each) until the file
is complete, and per-chunk failures are retried with backoff. Re-run
the same command after a Ctrl-C and the upload picks up where the
server stopped accepting bytes.

<remote> uses the same 3-segment frontend path as ` + "`olares-cli files ls`" + `.
A trailing '/' on <remote> means "upload into this directory";
without one, <remote> is treated as the full target path (useful to
rename a file on the way in).

The destination directory MUST already exist on the server. The
files-backend's "create directory" call auto-renames on collision
(POST .../Documents/ on an existing Documents creates "Documents (1)"
instead of returning a conflict), so we don't try to pre-create it —
use ` + "`olares-cli files mkdir [-p] <remote-path>`" + ` ahead of the
upload (or the web app) if you need to materialize a new directory
first.

Examples:

    # Upload one file into a directory.
    olares-cli files upload report.pdf drive/Home/Documents/

    # Same, but rename to 2026-Q1.pdf on the server.
    olares-cli files upload report.pdf drive/Home/Documents/2026-Q1.pdf

    # Upload a directory tree (preserves the source folder name).
    olares-cli files upload ./photos drive/Home/Backups/

    # Two files in flight at a time.
    olares-cli files upload ./photos drive/Home/Backups/ --parallel 2

    # Upload into a Sync (Seafile) library.
    olares-cli files upload notes.md sync/<repo_id>/Notes/

    # Upload into the Data volume or node-local cache / external storage.
    olares-cli files upload bigtar drive/Data/Backups/
    olares-cli files upload report.csv cache/<node>/<app>/
    olares-cli files upload movie.mp4 external/<node>/hdd1/Movies/

    # Upload into a connected cloud drive (S3 / Google Drive / Dropbox).
    olares-cli files upload backup.tar awss3/<account>/<bucket>/Backups/
    olares-cli files upload doc.pdf google/<account>/Documents/
    olares-cli files upload notes.md dropbox/<account>/Notes/
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
		"override the upload node name (defaults to <extend> for cache/external paths, "+
			"otherwise the first node from /api/nodes/)")
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
	// dispatch on FileType + Extend to compute the per-namespace
	// upload-protocol parameters (API parent_dir prefix, chunk-form
	// parent_dir prefix, web-app driveType form value, and whether
	// the path itself supplies the upload <node>).
	fp, err := ParseFrontendPath(remotePath)
	if err != nil {
		return err
	}
	apiRoot, chunkRoot, driveType, pathNode, err := uploadRootAndDriveType(fp)
	if err != nil {
		return err
	}
	// SubPath always starts with '/' from the parser; strip the leading
	// slash so BuildPlan sees a relative form like "Documents/Backups/".
	remoteSub := strings.TrimPrefix(fp.SubPath, "/")

	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return err
	}

	// Use the factory's no-timeout client for the upload session: an
	// 8 MiB chunk on a slow link can easily exceed the standard 30s
	// timeout, and we'd rather fail via context cancellation than
	// via http.Client.Timeout (the latter truncates the request body
	// mid-flight, which leaves the server in an inconsistent state).
	// X-Authorization injection + refresh-on-401 are handled by the
	// factory's refreshingTransport.
	httpClient, err := f.HTTPClientWithoutTimeout(ctx)
	if err != nil {
		return err
	}
	client := &upload.Client{
		HTTPClient: httpClient,
		BaseURL:    rp.FilesURL,
	}

	// Node-selection cascade — mirrors the web app's getUploadNode()
	// per-namespace plus our `files cp` / `files mv` `--node` override
	// behavior:
	//
	//   1. --node <name> wins outright.
	//   2. For cache/<node>/... and external/<node>/... the path's
	//      <extend> IS the upload node (web app's CacheDataAPI /
	//      ExternalDataAPI override getUploadNode() to return
	//      currentNode.name, which the user picked by navigating
	//      into /Cache/<node>/ or /Files/External/<node>/). We avoid
	//      the /api/nodes/ round-trip entirely in that case — it's a
	//      visible cost on slow networks and there's no need for it.
	//   3. drive/* and sync/*: fall back to the first node returned
	//      by /api/nodes/ (matches the web app's masterNode default).
	node := o.node
	if node == "" && pathNode != "" {
		node = pathNode
	}
	if node == "" {
		nodes, err := client.FetchNodes(ctx)
		if err != nil {
			return fmt.Errorf("fetch upload nodes: %w", err)
		}
		// Defense in depth: client.FetchNodes already errors on empty
		// data, but guard the index so a future regression in the
		// lower layer surfaces as a clean error here instead of an
		// "index out of range" panic at the call site.
		if len(nodes) == 0 {
			return fmt.Errorf("upload node list returned by /api/nodes/ is empty")
		}
		// Mirrors the web app's getUploadNode() — first node wins. A
		// future iteration can pick by master flag or by --node, but
		// the web app's default is good enough for the common case.
		node = nodes[0].Name
		if node == "" {
			return fmt.Errorf("upload node returned by /api/nodes/ has empty name")
		}
	}

	plan, err := upload.BuildPlan(localPath, remoteSub, apiRoot, chunkRoot)
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

	return runUploads(ctx, client, plan, node, driveType, o, out)
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
	driveType string,
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
			opts := plan.ToUploadOpts(task, node, driveType, o.chunkSize, o.maxRetries)

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
			res, err := client.UploadFile(gctx, opts, progress)
			if err != nil {
				return fmt.Errorf("%s: %w", task.RelativePath, err)
			}

			// Stage 2 of the cloud-drive (awss3 / google / dropbox)
			// upload protocol: the LAST chunk's response carried a
			// `taskId`, meaning the server queued a follow-up
			// "Olares-staging → cloud-bucket" transfer task. The
			// chunked upload merely delivered bytes to the Olares
			// side; the file is NOT visible in the user's actual
			// cloud bucket until WaitCloudTask sees the task hit
			// `completed`. Mirrors apps/packages/app/src/utils/
			// resumejs.ts onFileUploadSuccess L591-606 where the
			// web app registers the same taskId with Taskmanager
			// and lets it drive the post-upload polling.
			//
			// We deliberately do this inline (not in a separate
			// goroutine) so the per-file slot in the parallelism
			// errgroup stays held until stage 2 is fully accounted
			// for. That keeps `--parallel` honest: 2 cloud uploads
			// in flight means at most 2 stage-2 polls in flight,
			// and a failure surfaces against the file that caused
			// it instead of bleeding into the next file's progress
			// lines.
			if res.CloudTaskID != "" {
				fmt.Fprintf(out, "    %s: cloud transfer queued (task=%s)\n",
					task.RelativePath, res.CloudTaskID)

				var lastReportedStatus string
				var lastReportedProgress float64
				if err := client.WaitCloudTask(
					gctx, node, res.CloudTaskID, 0,
					func(u upload.CloudTaskUpdate) {
						// Throttle update lines: emit only when the
						// status string changes OR progress crosses
						// a 25-point boundary. Otherwise a long-
						// running transfer would print one line per
						// poll (every 2s by default), which buries
						// the per-file lines in noise without adding
						// useful info.
						crossed := false
						if u.Progress > 0 && u.Progress-lastReportedProgress >= 25 {
							crossed = true
						}
						if u.Status != lastReportedStatus || crossed {
							lastReportedStatus = u.Status
							lastReportedProgress = u.Progress
							if u.Status == "" {
								return
							}
							fmt.Fprintf(out, "    %s: cloud transfer %s (%.0f%%)\n",
								task.RelativePath, u.Status, u.Progress)
						}
					},
				); err != nil {
					return fmt.Errorf("%s: %w", task.RelativePath, err)
				}
				fmt.Fprintf(out, "    %s: cloud transfer completed\n",
					task.RelativePath)
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

// supportedUploadNamespaces is the human-readable list of
// destinations `files upload` accepts, used in error messages so the
// failure mode is self-describing without the caller having to read
// the long help text. Keep in sync with uploadRootAndDriveType
// switch arms below.
const supportedUploadNamespaces = "drive/Home, drive/Data, sync/<repo_id>, cache/<node>, external/<node>/<volume>, awss3/<account>, google/<account>, or dropbox/<account>"

// uploadRootAndDriveType inspects a parsed FrontendPath and returns
// the four values runUpload + BuildPlan + the chunk pipeline need:
//
//   - apiRoot:   the prefix used in API queries
//     (`/upload/upload-link/<node>/?file_path=...`,
//     `/upload/file-uploaded-bytes/<node>/?parent_dir=...`).
//     Always `/<fileType>/<extend>` (e.g. `/drive/Home`,
//     `/drive/Data`, `/sync/<repo_id>`, `/cache/<node>`,
//     `/external/<node>`, `/awss3/<account>`, ...).
//   - chunkRoot: the prefix used in the chunk POST's `parent_dir`
//     multipart form field. For all the files-backend-managed
//     namespaces (Drive/Data/Cache/External) AND for the cloud
//     drives we currently support (awss3/google/dropbox — which in
//     v2 inherit the regular chunk pipeline from DriveDataAPI
//     without overrides) the chunk endpoint is the files-backend's
//     `/upload/...` route which expects the same API-form
//     `parent_dir`, so chunkRoot == apiRoot. For Sync the chunk
//     endpoint is Seafile's `/seafhttp/upload-aj/<token>`, which
//     expects `parent_dir` to be the path INSIDE the repo (the
//     token already pins repo + permission); so we pass an empty
//     chunkRoot and `parentDirFor("", sub)` produces `/sub/`.
//   - driveType: the literal that goes into the chunk POST's
//     `driveType` form field (`Drive` / `Data` / `Sync` / `Cache` /
//     `External` / `Awss3` / `Google` / `Dropbox`), matching the
//     web app's resumejs.ts setQuery output. The server is
//     permissive on case here (the existing CLI ships capitalized
//     literals while the web app sends lowercase enum values, both
//     work) — we keep capitalized because that's what the existing
//     wire tests assert and what shipped in the Drive/Sync arms.
//   - pathNode: the upload-`{node}` URL segment when the path itself
//     supplies it (cache/external — `<extend>` IS the node, mirroring
//     the web app's CacheDataAPI/ExternalDataAPI getUploadNode()
//     overrides). Empty for namespaces where the cobra command should
//     fall back to /api/nodes/.
func uploadRootAndDriveType(fp FrontendPath) (apiRoot, chunkRoot, driveType, pathNode string, err error) {
	switch fp.FileType {
	case "drive":
		switch fp.Extend {
		case "Home":
			return "/drive/Home", "/drive/Home", "Drive", "", nil
		case "Data":
			// drive/Data uses the same upload protocol as drive/Home
			// — the web app's DataDataAPI inherits getFileServerUploadLink
			// / getFileUploadedBytes / formatUploaderPath from
			// DriveDataAPI without overrides; only `driveType` differs.
			return "/drive/Data", "/drive/Data", "Data", "", nil
		default:
			return "", "", "", "", fmt.Errorf("upload destination must be under %s (drive extend must be Home or Data, got %q)",
				supportedUploadNamespaces, fp.String())
		}
	case "sync":
		if fp.Extend == "" {
			return "", "", "", "", fmt.Errorf("upload destination sync extend(repo_id) must be non-empty")
		}
		// chunkRoot is intentionally empty: Seafile's seafhttp/upload-aj
		// reads `parent_dir` as a path inside the repo, so the chunk
		// form field should be `/sub/` rather than `/sync/<repo_id>/sub/`.
		return "/sync/" + fp.Extend, "", "Sync", "", nil
	case "cache":
		if fp.Extend == "" {
			return "", "", "", "", fmt.Errorf("upload destination cache extend(node) must be non-empty (e.g. cache/<node>/<sub>/)")
		}
		// Cache is node-local storage; the path's <extend> IS the
		// upload node, so we surface it as pathNode for the cobra
		// layer (which then skips the /api/nodes/ round-trip).
		root := "/cache/" + fp.Extend
		return root, root, "Cache", fp.Extend, nil
	case "external":
		if fp.Extend == "" {
			return "", "", "", "", fmt.Errorf("upload destination external extend(node) must be non-empty (e.g. external/<node>/<volume>/<sub>/)")
		}
		root := "/external/" + fp.Extend
		return root, root, "External", fp.Extend, nil

	// --- Cloud drives that share the Drive multipart-POST upload
	//     pipeline. The web app's Awss3DataAPI / GoogleDataAPI /
	//     DropboxDataAPI all extend DriveDataAPI in v2 WITHOUT
	//     overriding getFileServerUploadLink or
	//     getFileUploadedBytes (see
	//     apps/packages/app/src/api/files/v2/{awss3,google,dropbox}/data.ts),
	//     so the wire flow is identical to Drive — only the
	//     parent_dir prefix and driveType form-field literal change.
	//     Per-account upload still routes through /api/nodes/ since
	//     these namespaces share the masterNode default
	//     (DriveDataAPI.getUploadNode()).
	case "awss3":
		if fp.Extend == "" {
			return "", "", "", "", fmt.Errorf("upload destination awss3 extend(account) must be non-empty (e.g. awss3/<account>/<bucket>/<sub>/)")
		}
		root := "/awss3/" + fp.Extend
		return root, root, "Awss3", "", nil
	case "google":
		if fp.Extend == "" {
			return "", "", "", "", fmt.Errorf("upload destination google extend(account) must be non-empty (e.g. google/<account>/<sub>/)")
		}
		root := "/google/" + fp.Extend
		return root, root, "Google", "", nil
	case "dropbox":
		if fp.Extend == "" {
			return "", "", "", "", fmt.Errorf("upload destination dropbox extend(account) must be non-empty (e.g. dropbox/<account>/<sub>/)")
		}
		root := "/dropbox/" + fp.Extend
		return root, root, "Dropbox", "", nil

	// Tencent COS (v2 TencentDataAPI) is the lone cloud drive that
	// overrides getFileServerUploadLink + getFileUploadedBytes to
	// post to /drive/create_direct_upload_task and stream chunks
	// via /drive/direct_upload_file/<task_id> as octet payloads
	// (NOT multipart). That's a different protocol the CLI's chunk
	// pipeline can't speak; reject explicitly with a self-describing
	// error rather than letting it fall into the generic default
	// arm below — the latter would just say "must be under <list>"
	// and leave the user wondering why the parser accepted the
	// path but the verb refuses it.
	case "tencent":
		return "", "", "", "", fmt.Errorf(
			"upload to tencent COS is not supported by this verb: it uses the octet "+
				"/drive/direct_upload_file/<task_id> protocol (see web app's TencentDataAPI), "+
				"which the CLI chunk pipeline does not implement; got %q",
			fp.String())

	default:
		return "", "", "", "", fmt.Errorf("upload destination must be under %s (got %q)",
			supportedUploadNamespaces, fp.String())
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
