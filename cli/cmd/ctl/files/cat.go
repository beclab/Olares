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
// Streams the raw bytes of a single remote file to stdout. The wire
// call is GET /api/raw/<encPath>?inline=true (same path the LarePass
// web app uses for text-content previews — `inline=true` only
// affects Content-Disposition, the body is identical).
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

Examples:

    olares-cli files cat drive/Home/Documents/notes.md
    olares-cli files cat drive/Home/Logs/today.log | tail -n 50
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

	httpClient := newUploadHTTPClient(rp.InsecureSkipVerify)
	client := &download.Client{
		HTTPClient:  httpClient,
		BaseURL:     rp.FilesURL,
		AccessToken: rp.AccessToken,
	}

	plain := strings.TrimSuffix(fp.String(), "/")

	// Probe before streaming. Two cheap wins:
	//   - friendly "is a directory" message for `cat drive/Home/`
	//     instead of the server's terse 400;
	//   - 401/403/404 reformatted with the standard CTA before we
	//     start writing partial data to stdout.
	st, err := client.Stat(ctx, plain)
	if err != nil {
		return reformatHTTPErr(err, rp.OlaresID, "stat", plain)
	}
	if st.IsDir {
		return fmt.Errorf("%s is a directory: cat only works on files (use `olares-cli files ls %s` to list it)",
			fp.String(), fp.String())
	}

	if _, err := client.StreamRaw(ctx, plain, out); err != nil {
		return reformatHTTPErr(err, rp.OlaresID, "cat", plain)
	}
	return nil
}
