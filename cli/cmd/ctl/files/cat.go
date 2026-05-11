package files

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/internal/files/download"
)

// NewCatCommand: `olares-cli files cat <remote-path>`
//
// Streams the raw bytes of a single remote file to stdout. Two wire
// flows are dispatched based on the FrontendPath's namespace:
//
//   - drive / sync / cache / external (and share): GET
//     /api/raw/<encPath>?inline=true — same path the LarePass web
//     app uses for text-content previews; `inline=true` only affects
//     Content-Disposition, the body is identical.
//   - awss3 / google / dropbox / tencent: GET
//     /drive/download_sync_stream?drive=<type>&cloud_file_path=<path>&name=<extend>,
//     mirroring the web app's `generateDownloadUrl` helper
//     (apps/packages/app/src/api/files/v2/{awss3,google,dropbox}/utils.ts).
//     The /api/raw/ endpoint on these namespaces returns metadata /
//     preview JSON rather than raw bytes, so it's not a substitute
//     for cat — see cat_test.go for the divergence in detail.
//
// Like `cat` itself, this is binary-safe: we don't sniff or
// interpret the body, we just copy it through. That means cat-ing a
// huge image will dump the bytes — the user is expected to pipe to
// `less`, `head`, or a similar tool when they care about safety.
//
// We Stat the path before fetching so a directory target produces a
// clear "is a directory" error rather than the server's terser
// "not a file, path: ..." 400.
func NewCatCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cat <remote-path>",
		Short: "stream a remote file's contents to stdout",
		Long: `Stream the raw bytes of a single file on the per-user files-backend to stdout.

Equivalent to ` + "`olares-cli files download <remote> -`" + ` if a future
` + "`-`" + ` -means-stdout convention is added — for now ` + "`cat`" + ` is the explicit
verb. The transfer is binary-safe (no buffering, no transformation),
so piping into ` + "`less`" + ` / ` + "`hexdump`" + ` / ` + "`head -c`" + ` works as expected.

Directories produce an error rather than a recursive concatenation
(use ` + "`files download <remote>/`" + ` if you want the contents on disk
first).

Cloud drives (awss3 / google / dropbox / tencent) use a different
wire endpoint (` + "`/drive/download_sync_stream`" + `) than the
files-backend-managed namespaces — the client picks the right one
automatically based on the remote path's first segment.

Examples:

    olares-cli files cat drive/Home/Documents/notes.md
    olares-cli files cat drive/Home/Logs/today.log | tail -n 50
    olares-cli files cat awss3/<account>/notes.md
    olares-cli files cat google/<account>/Documents/today.log
    olares-cli files cat dropbox/<account>/Notes/idea.md
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCat(cmd.Context(), f, cmd.OutOrStdout(), args[0])
		},
	}
	return cmd
}

func runCat(ctx context.Context, f *cmdutil.Factory, out io.Writer, remoteArg string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	fp, err := ParseFrontendPath(remoteArg)
	if err != nil {
		return err
	}

	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return err
	}

	httpClient, err := f.HTTPClientWithoutTimeout(ctx)
	if err != nil {
		return err
	}
	client := &download.Client{
		HTTPClient: httpClient,
		BaseURL:    rp.FilesURL,
	}

	plain := strings.TrimSuffix(fp.String(), "/")

	// Probe before streaming. Two cheap wins:
	//   - friendly "is a directory" message for `cat drive/Home/`
	//     instead of the server's terse 400;
	//   - 401/403/404 reformatted with the standard CTA before we
	//     start writing partial data to stdout.
	//
	// Stat works uniformly across all namespaces — the parent-listing
	// strategy (download/stat.go) decodes both the Drive `items`
	// envelope and the cloud-drive `data` envelope.
	st, err := client.Stat(ctx, plain)
	if err != nil {
		return reformatHTTPErr(err, rp.OlaresID, "stat", plain)
	}
	if st.IsDir {
		return fmt.Errorf("%s is a directory: cat only works on files (use `olares-cli files ls %s` to list it)",
			fp.String(), fp.String())
	}

	// Cloud drives (awss3 / google / dropbox / tencent) don't serve
	// raw bytes from /api/raw/<path> — that endpoint returns
	// JSON/preview content on those namespaces, not the file itself.
	// The /drive/download_sync_stream endpoint is the right one for
	// streaming, and it's what the LarePass web app's
	// `generateDownloadUrl` helpers emit (utils.ts in v2/awss3,
	// v2/google, v2/dropbox). cloud_file_path is the SubPath the
	// frontend parser already extracted (it preserves the leading '/'
	// and any unicode/spaces); the URL builder percent-encodes it for
	// us in StreamCloudFile.
	if isCloudDriveType(fp.FileType) {
		if _, err := client.StreamCloudFile(ctx, fp.FileType, fp.SubPath, fp.Extend, out); err != nil {
			return reformatHTTPErr(err, rp.OlaresID, "cat", plain)
		}
		return nil
	}

	if _, err := client.StreamRaw(ctx, plain, out); err != nil {
		return reformatHTTPErr(err, rp.OlaresID, "cat", plain)
	}
	return nil
}

// isCloudDriveType reports whether `fileType` is one of the
// cloud-bridge-backed namespaces that need /drive/download_sync_stream
// for raw-byte access (rather than /api/raw/<path>). The list mirrors
// the `DriveType` enum from
// apps/packages/app/src/utils/interface/files.ts plus what
// awss3/dropbox/google/tencent v2 utils emit for `download()`.
//
// Kept as a free function (rather than a method on FrontendPath) so
// it stays close to the cat-specific dispatch logic — other verbs
// (cp/rm/mv) need their own per-namespace decisions, not necessarily
// keyed off the same set, so a shared predicate would over-couple.
func isCloudDriveType(fileType string) bool {
	switch fileType {
	case "awss3", "google", "dropbox", "tencent":
		return true
	}
	return false
}
