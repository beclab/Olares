package files

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/internal/files/rename"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/credential"
)

// NewRenameCommand: `olares-cli files rename <remote-path> <new-name>`
// (alias: `rn`).
//
// In-place rename — the entry stays in the same parent; only its
// basename changes. Wire shape mirrors the LarePass web app's
// renameFileItem helper (apps/.../api/files/v2/common/utils.ts):
//
//	PATCH /api/resources/<fileType>/<extend><subPath>[/]?destination=<encName>
//
// This is intentionally NOT routed through `cp` / `mv`'s
// PATCH /api/paste/<node>/ surface, because:
//
//   - Rename is synchronous on the backend (no task_id), so the user
//     gets a "done" response, not a "queued" one.
//   - It doesn't take a {node} URL segment — /api/resources is the
//     uniform per-resource path that works against drive / sync /
//     cloud / external without any node hint.
//   - The only payload is the new bare basename in the `destination`
//     query value; there's no JSON body. That matches the frontend's
//     "Rename" modal exactly.
//
// `mv` can also rename (single-source, non-trailing-slash <dst>) but
// goes through the async paste queue. Prefer `rename` for the simple
// "I just want to change the name" case — fewer round-trips, no node
// resolution, and immediate feedback.
//
// CLI semantics:
//
//   - <remote-path>: full 3-segment frontend path (e.g.
//     `drive/Home/Documents/foo.pdf`); same parser as `ls`/`cp`. A
//     trailing '/' marks the source as a directory and is preserved
//     on the wire so the backend routes through its directory handler.
//   - <new-name>: BARE basename only. No '/' or '\\' allowed —
//     cross-directory moves are `mv`'s job. Empty / "."/".." are
//     rejected as obvious typos.
//   - Refuses to rename the volume root (e.g. `drive/Home/`) — same
//     safety policy as `rm` and `cp`.
//   - Refuses a no-op rename where <new-name> equals the source's
//     current basename — almost always a typo.
//
// Collision handling is server-decided: the backend may auto-rename,
// overwrite, or 409 depending on the storage class. We surface its
// answer verbatim. If a future need for explicit `--force` / `--no-clobber`
// flags shows up we'll thread `override`/`rename` query params through
// the rename package.
func NewRenameCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rename <remote-path> <new-name>",
		Aliases: []string{"rn"},
		Short:   "rename a remote file or directory in place (synchronous)",
		Long: `Rename a remote entry in place — same parent, new basename.

Wire shape:

    PATCH /api/resources/<fileType>/<extend><subPath>[/]?destination=<encName>

This hits the per-resource PATCH endpoint (synchronous; not the
` + "`/api/paste/<node>/`" + ` task-queue surface that ` + "`cp`" + ` / ` + "`mv`" + ` use), so the
response is the final state — no ` + "`task_id`" + ` polling needed.

Differences from ` + "`files mv`" + `:

    files rename            — same dir, new basename;   synchronous, no node
    files mv <src> <dst>    — same OR different dir;    async via paste queue

Use ` + "`rename`" + ` for the simple in-place case; reach for ` + "`mv`" + ` when you
need cross-directory or cross-volume moves.

Argument shape:

    <remote-path>   3-segment frontend path (same as ` + "`files ls`" + `):
                    drive/Home/Documents/foo.pdf
                    sync/<repo_id>/notes/old-name/
    <new-name>      BARE basename — no '/' or '\\'.
                    Empty, '.', '..' are rejected.

A trailing '/' on <remote-path> signals the source is a directory and
is preserved on the wire (so the backend routes through its directory
handler). Volume roots (e.g. ` + "`drive/Home/`" + `) are refused.

Examples:

    # Rename a file.
    olares-cli files rename drive/Home/Documents/foo.pdf foo-2026.pdf

    # Rename a directory (note the trailing '/').
    olares-cli files rename drive/Home/Photos/old/ archive

    # Sync repo subdirectory.
    olares-cli files rename sync/<repo_id>/notes/draft/ final

    # Alias.
    olares-cli files rn drive/Home/foo.txt bar.txt

Collision behavior is server-determined (the backend may auto-rename,
overwrite, or return 409 depending on the storage class). If a target
with the new name exists, run ` + "`olares-cli files ls`" + ` afterwards to
confirm what landed.
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRename(cmd.Context(), f, cmd.OutOrStdout(), args[0], args[1])
		},
	}
	return cmd
}

// runRename is the cobra-side glue: parse the path, build the rename
// Op (which validates the new basename), set up the HTTP client, fire
// one PATCH, and reformat any HTTPError into a friendly CTA.
//
// Kept as a free function (rather than a method on a struct) so it
// matches the runCpMv / runRm shape in this package — easier to unit-
// test by passing in a custom io.Writer if the need arises.
func runRename(
	ctx context.Context,
	f *cmdutil.Factory,
	out io.Writer,
	pathArg string,
	newName string,
) error {
	if ctx == nil {
		ctx = context.Background()
	}

	tgt, err := frontendPathToRenameTarget(pathArg)
	if err != nil {
		return err
	}

	op, err := rename.Plan(tgt, newName)
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
	client := &rename.Client{
		HTTPClient: httpClient,
		BaseURL:    rp.FilesURL,
	}

	// Plan summary first — gives the user a chance to ^C before any
	// state-changing wire call goes out. Same UX shape as cp/mv/rm.
	fmt.Fprintf(out, "rename: %s → %s\n", op.DisplaySrc, op.DisplayDst)

	if err := client.Rename(ctx, op); err != nil {
		return reformatRenameHTTPErr(err, rp.OlaresID, op.DisplaySrc, op.DisplayDst)
	}

	fmt.Fprintf(out, "  ✓ %s → %s\n", op.DisplaySrc, op.DisplayDst)
	return nil
}

// frontendPathToRenameTarget converts a user-supplied path into the
// rename package's Target shape. The trailing '/' on the input is
// preserved as the IsDirIntent signal so the wire URL keeps the
// directory marker (the backend routes file vs dir handlers off it,
// same as `rm` / `cp`).
//
// Refusing to rename the volume root is enforced both here AND in
// rename.Plan — defense in depth: this layer can give a friendlier
// error message naming the actual <fileType>/<extend>, while Plan
// stays robust against a future caller that bypasses this helper.
func frontendPathToRenameTarget(raw string) (rename.Target, error) {
	fp, err := ParseFrontendPath(raw)
	if err != nil {
		return rename.Target{}, err
	}
	if strings.Trim(fp.SubPath, "/") == "" {
		return rename.Target{}, fmt.Errorf(
			"refusing to rename the root of %s/%s; rename needs an entry name",
			fp.FileType, fp.Extend)
	}
	return rename.Target{
		FileType:    fp.FileType,
		Extend:      fp.Extend,
		SubPath:     fp.SubPath,
		IsDirIntent: strings.HasSuffix(fp.SubPath, "/"),
	}, nil
}

// reformatRenameHTTPErr maps rename.HTTPError onto user-friendly
// messages, mirroring the cp/rm/download reformatters. Status branches:
//
//   - 401/403: token rejected → suggest `profile login`. Same wording
//     as the other verbs so the user gets one consistent CTA.
//   - 404: source not found → echo the path so the user can re-try
//     against a corrected one.
//   - 409: target name already exists → tell the user to either
//     remove the conflict or pick a different name. The backend
//     handles collision policy, so we don't try to overwrite.
//
// Typed credential errors from the refreshing transport are surfaced
// verbatim — see reformatHTTPErr in download.go for why those bypass
// the HTTP-status branches entirely.
func reformatRenameHTTPErr(err error, olaresID, src, dst string) error {
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
	var hErr *rename.HTTPError
	if errors.As(err, &hErr) {
		switch hErr.Status {
		case 401, 403:
			if olaresID != "" {
				return fmt.Errorf(
					"server rejected the access token (HTTP %d); please run: olares-cli profile login --olares-id %s",
					hErr.Status, olaresID)
			}
			return fmt.Errorf(
				"server rejected the access token (HTTP %d); please re-run `olares-cli profile login`",
				hErr.Status)
		case 404:
			return fmt.Errorf("rename %s: not found on the server (HTTP 404)", src)
		case 409:
			return fmt.Errorf(
				"rename %s → %s: target already exists (HTTP 409); remove it first or pick a different name",
				src, dst)
		}
	}
	return err
}
