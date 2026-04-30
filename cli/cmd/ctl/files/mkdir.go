package files

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/internal/files/mkdir"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/credential"
)

// mkdirOptions holds the per-invocation flags. Kept tiny on purpose —
// `mkdir` is a small verb and a `mkdir -p` sibling is the only
// behavior switch worth threading through a flag.
type mkdirOptions struct {
	// parents toggles `-p` semantics: create missing intermediate
	// directories left-to-right and silently skip ones that already
	// exist. Without this flag, only the leaf path is POSTed; the
	// server's auto-rename quirk surfaces as-is (see the comment
	// inside runMkdir for why we don't pre-validate).
	parents bool
}

// NewMkdirCommand: `olares-cli files mkdir [-p] <remote-path>`.
//
// Creates a directory on the per-user files-backend. Wire shape:
//
//	POST /api/resources/<fileType>/<extend><subPath>/
//
// matching the LarePass web app's per-driver `createDir` helpers
// (apps/.../api/files/v2/{drive,sync,cache,external,awss3,dropbox,
// google,tencent}/utils.ts) — the trailing '/' on the URL is what
// distinguishes "create directory" from "create empty file" on this
// endpoint. Body is empty.
//
// Two important behavioral notes:
//
//  1. **Auto-rename on collision (NOT 409).** The current files-
//     backend silently auto-renames on collision: POST
//     `/.../Documents/` when `Documents` already exists creates
//     `Documents (1)` next to the original instead of returning 409.
//     That's surprising for an "ensure dir exists" call — so the CLI
//     surfaces this in the post-run summary and recommends running
//     `files ls` to confirm what landed.
//
//  2. **`-p` mode does parent-listing existence checks.** Because of
//     the auto-rename quirk, naively POSTing every prefix in
//     `-p drive/Home/A/B/C/` would produce `A (1)/B (1)/C` if any
//     prefix already existed. The CLI works around this by listing
//     each prefix's parent and skipping the segment when the
//     basename is already there as a directory. The cost is one
//     extra GET per existing prefix, which is well worth the
//     correctness win.
//
// Supported namespaces: ALL of the 3-segment frontend types (drive,
// sync, cache, external, awss3, google, dropbox, tencent — and also
// share/internal if the user's role permits, the wire endpoint is
// the same; the server will 403 on read-only views). Tencent is
// supported here even though `files upload` rejects it — mkdir
// shares the generic /api/resources POST path with all other
// namespaces, so there's no upload-pipeline divergence to worry
// about.
func NewMkdirCommand(f *cmdutil.Factory) *cobra.Command {
	o := &mkdirOptions{}
	cmd := &cobra.Command{
		Use:     "mkdir [-p] <remote-path>",
		Aliases: []string{"md"},
		Short:   "create a remote directory (optionally with -p for missing intermediates)",
		Long: `Create a directory on the per-user files-backend.

Wire shape (uniform across drive / sync / cache / external / cloud-drive
namespaces):

    POST /api/resources/<fileType>/<extend><subPath>/

The trailing '/' is what tells the backend "this is a directory" (the
same convention the LarePass web app's createDir helpers emit).

Argument shape:

    <remote-path>   3-segment frontend path — same as ` + "`olares-cli files ls`" + `:
                    drive/Home/Documents/Backups
                    sync/<repo_id>/Notes/2026
                    awss3/<account>/Backups
                    google/<account>/Documents/Photos
                    cache/<node>/scratch
                    external/<node>/usb1/Backups
                    tencent/<account>/Backups

A trailing '/' on <remote-path> is allowed but not required; mkdir is
always a "this is a directory" operation, so the URL ends with '/'
either way.

Flags:

    -p, --parents    Create missing intermediate directories. Skips
                     prefixes that already exist (one extra parent-
                     listing GET per existing prefix). Without this
                     flag, only the leaf is created and the parent
                     MUST already exist on the server.

` + "`-p`" + ` examples:

    olares-cli files mkdir drive/Home/Documents/Backups
    olares-cli files mkdir -p drive/Home/A/B/C/
    olares-cli files mkdir -p sync/<repo_id>/notes/2026/Q2
    olares-cli files mkdir -p awss3/<account>/Backups/2026
    olares-cli files mkdir -p google/<account>/Drafts

Caveats:

  - Auto-rename quirk: if the leaf (or any parent) already exists,
    the files-backend may silently create "Foo (1)" instead of
    returning 409. ` + "`-p`" + ` mode side-steps this for parents (it skips
    existing prefixes via a parent listing). For the leaf, the CLI
    prints a hint after the call so you can ` + "`files ls`" + ` and confirm.
  - Refuses to mkdir a volume root (e.g. ` + "`drive/Home/`" + ` or
    ` + "`sync/<repo_id>/`" + `) — those always exist, so the call would
    just be a no-op (or trigger the auto-rename quirk on the
    extend folder, which is never what you want).
  - Bare ` + "`.`" + ` / ` + "`..`" + ` segments are rejected as obvious typos.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMkdir(cmd.Context(), f, cmd.OutOrStdout(), args[0], o)
		},
	}
	cmd.Flags().BoolVarP(&o.parents, "parents", "p", false,
		"create missing intermediate directories (skip prefixes that already exist)")
	return cmd
}

// runMkdir is the cobra-side glue: parse the path, build the mkdir
// Op(s), set up the HTTP client, fire the POSTs, and reformat any
// HTTPError into a friendly CTA. Mirrors the runRename / runCpMv
// shape so all per-verb runners read the same way.
func runMkdir(
	ctx context.Context,
	f *cmdutil.Factory,
	out io.Writer,
	pathArg string,
	o *mkdirOptions,
) error {
	if ctx == nil {
		ctx = context.Background()
	}

	tgt, err := frontendPathToMkdirTarget(pathArg)
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
	client := &mkdir.Client{
		HTTPClient: httpClient,
		BaseURL:    rp.FilesURL,
	}

	if o.parents {
		return runMkdirP(ctx, client, tgt, rp.OlaresID, out)
	}
	return runMkdirSingle(ctx, client, tgt, rp.OlaresID, out)
}

// runMkdirSingle handles the no-`-p` case: one POST against the leaf
// path. The auto-rename quirk on the server side means a collision
// here does NOT surface as 409 — the call appears to succeed and a
// second directory ("Foo (1)") gets created. We can't detect that
// from the response alone, so we emit a one-line hint after success
// recommending `files ls` to confirm. (Detecting it would require an
// extra parent-listing GET, which `-p` mode already pays — non-`-p`
// mode prefers speed.)
func runMkdirSingle(
	ctx context.Context,
	client *mkdir.Client,
	tgt mkdir.Target,
	olaresID string,
	out io.Writer,
) error {
	op, err := mkdir.Plan(tgt)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "mkdir: %s\n", op.DisplayPath)
	if err := client.Mkdir(ctx, op); err != nil {
		return reformatMkdirHTTPErr(err, olaresID, op.DisplayPath)
	}
	fmt.Fprintf(out, "  ✓ %s\n", op.DisplayPath)
	fmt.Fprintf(out, "  hint: if %q already existed, the server may have auto-renamed the new entry to \"%s (1)\"; "+
		"run `olares-cli files ls %s` on the parent to confirm.\n",
		lastSegmentForHint(tgt.SubPath),
		lastSegmentForHint(tgt.SubPath),
		parentDisplayPath(tgt))
	return nil
}

// runMkdirP handles `-p`: walk the segments left-to-right, list each
// prefix's parent, skip prefixes that already exist as directories,
// and POST the rest. A prefix that exists as a NON-directory is a
// hard error — falling through would auto-rename us into a
// `(1)`-suffixed sibling.
//
// One-extra-GET-per-prefix is the cost we pay for correctness on the
// auto-rename-on-collision backend. For typical `-p` invocations
// (2-4 segments) this is negligible; for pathological 20-segment
// trees the user can always batch with multiple non-`-p` commands.
func runMkdirP(
	ctx context.Context,
	client *mkdir.Client,
	tgt mkdir.Target,
	olaresID string,
	out io.Writer,
) error {
	ops, err := mkdir.PlanRecursive(tgt)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "mkdir -p: %s (%d level%s)\n",
		ops[len(ops)-1].DisplayPath, len(ops), pluralS(len(ops)))

	for _, op := range ops {
		// `plain` is the un-encoded form Exists takes. Strip the
		// `/api/resources/` prefix and trailing '/' from
		// op.Endpoint and inverse-decode? Actually Plan exposes
		// DisplayPath which is exactly that form (with a trailing
		// '/'). Easier to derive once.
		plain := strings.TrimSuffix(op.DisplayPath, "/")
		found, isDir, err := client.Exists(ctx, plain)
		if err != nil {
			return reformatMkdirHTTPErr(err, olaresID, op.DisplayPath)
		}
		if found {
			if !isDir {
				return fmt.Errorf(
					"mkdir -p: %s already exists but is NOT a directory; "+
						"refusing to proceed (the auto-rename quirk would otherwise create a sibling)",
					op.DisplayPath)
			}
			fmt.Fprintf(out, "  · %s (already exists, skipped)\n", op.DisplayPath)
			continue
		}
		if err := client.Mkdir(ctx, op); err != nil {
			return reformatMkdirHTTPErr(err, olaresID, op.DisplayPath)
		}
		fmt.Fprintf(out, "  ✓ %s\n", op.DisplayPath)
	}
	return nil
}

// frontendPathToMkdirTarget converts the user-supplied path into the
// mkdir package's Target shape. Refusing the volume root is
// enforced both here AND in mkdir.Plan — defense in depth: this
// layer can give a friendlier error message naming the actual
// `<fileType>/<extend>`, while Plan stays robust against a future
// caller that bypasses this helper.
func frontendPathToMkdirTarget(raw string) (mkdir.Target, error) {
	fp, err := ParseFrontendPath(raw)
	if err != nil {
		return mkdir.Target{}, err
	}
	if strings.Trim(fp.SubPath, "/") == "" {
		return mkdir.Target{}, fmt.Errorf(
			"refusing to mkdir the root of %s/%s; pick a subdirectory name (e.g. %s/%s/NewFolder)",
			fp.FileType, fp.Extend, fp.FileType, fp.Extend)
	}
	return mkdir.Target{
		FileType: fp.FileType,
		Extend:   fp.Extend,
		SubPath:  fp.SubPath,
	}, nil
}

// reformatMkdirHTTPErr maps mkdir.HTTPError onto user-friendly
// messages, mirroring the cp / rm / rename / download reformatters.
// Status branches:
//
//   - 401/403: token rejected → suggest `profile login`. Same wording
//     as the other verbs so the user gets one consistent CTA.
//   - 404: parent directory not found → suggest `-p` to create the
//     missing intermediates.
//   - 409: target name already exists (rare on this backend, see
//     Mkdir's comment for context) → suggest a different name or
//     remove the existing entry.
//
// Typed credential errors from the refreshing transport are surfaced
// verbatim — see reformatHTTPErr in download.go for why those bypass
// the HTTP-status branches entirely.
func reformatMkdirHTTPErr(err error, olaresID, displayPath string) error {
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
	var hErr *mkdir.HTTPError
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
			return fmt.Errorf(
				"mkdir %s: parent directory does not exist (HTTP 404); pass -p to create missing intermediates",
				displayPath)
		case 409:
			return fmt.Errorf(
				"mkdir %s: target already exists (HTTP 409); remove it first or pick a different name",
				displayPath)
		}
	}
	return err
}

// lastSegmentForHint returns the basename of a SubPath for use in the
// post-success "auto-rename" hint. SubPath always starts with '/';
// trailing '/' is tolerated. Returns "" for the root case (which
// Plan should already have rejected, so this is a defensive
// fallback).
func lastSegmentForHint(sub string) string {
	s := strings.Trim(sub, "/")
	if s == "" {
		return ""
	}
	if i := strings.LastIndex(s, "/"); i >= 0 {
		return s[i+1:]
	}
	return s
}

// parentDisplayPath returns the human-readable form of the requested
// path's parent, suitable for the post-success `files ls` CTA. The
// volume root falls back to `<fileType>/<extend>/` so the suggestion
// always points at a listable path.
func parentDisplayPath(t mkdir.Target) string {
	clean := strings.Trim(t.SubPath, "/")
	if clean == "" {
		return t.FileType + "/" + t.Extend + "/"
	}
	if i := strings.LastIndex(clean, "/"); i >= 0 {
		return t.FileType + "/" + t.Extend + "/" + clean[:i] + "/"
	}
	return t.FileType + "/" + t.Extend + "/"
}

