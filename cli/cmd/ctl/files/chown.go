package files

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/internal/files/permission"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/credential"
)

// chownOptions captures the parsed flag state for `files chown`. We
// keep `uidStr` (the raw flag string) separate from the parsed int so
// "the user passed --uid" can be distinguished from "the user typed
// --uid 0" — the former triggers a GET, the latter triggers a PUT
// even when the value is zero.
type chownOptions struct {
	uidStr    string
	uidSet    bool
	recursive bool
}

// commonChownUIDPresets enumerates the two values the LarePass web
// app's properties dialog exposes (apps/.../components/files/prompts/
// InfoDialog.vue: `permissionOption` is `[{label: 'Root', value: 0},
// {label: 'User', value: 1000}]`). They're surfaced in the long help
// + error messages so a user who's unsure what UID to pass has
// somewhere to start.
//
// The CLI does NOT restrict --uid to these two values — the wire
// accepts any int and a well-set-up Olares node can host POSIX users
// outside the LarePass-default 0 / 1000 split. This is purely
// documentation.
var commonChownUIDPresets = []struct {
	UID   int
	Label string
}{
	{0, "Root"},
	{1000, "User"},
}

// NewChownCommand: `olares-cli files chown <remote-path> [--uid N] [-r]`
//
// Get or set the POSIX owner UID of a file or directory on the
// per-user files-backend. This is the CLI counterpart of LarePass's
// "Permission" tab in the file properties dialog
// (apps/packages/app/src/components/files/prompts/InfoDialog.vue) —
// it's the only writable field that dialog exposes, so a single CLI
// verb covers the same surface.
//
// Wire shape (mirrors `operationStore.getPermission` /
// `setPermission` in apps/.../stores/operation.ts):
//
//	GET  /api/permission/<fileType>/<extend><subPath>          → {uid: <int>}
//	PUT  /api/permission/<fileType>/<extend><subPath>?uid=<int>[&recursive=1]
//
// Modes:
//
//   - Without --uid: GET the current uid and print it. Useful for
//     scripts that want to inspect ownership without changing it.
//   - With --uid: PUT the new uid. Add -r/--recursive to apply the
//     change to every child as well, same as the GUI's recursive
//     toggle.
//
// Namespace gate (mirrors LarePass's `permissionInDriveType` =
// [Drive, Data, Cache] in InfoDialog.vue): only `drive/Home`,
// `drive/Data`, and `cache/<node>/...` are accepted on the CLI.
// Other namespaces (sync, external, every cloud account drive) hide
// the Permission tab in the web app and the server-side behavior at
// those paths is not part of this surface — we fail fast client-side
// rather than emit a request the GUI itself wouldn't.
//
// Refused targets:
//
//   - The volume root (e.g. `drive/Home/`, `cache/<node>/`): chowning
//     a whole namespace root is almost never what the user means and
//     the failure mode if they did mean it is severe enough to be
//     worth refusing. Use a one-level-deeper path with -r if you want
//     to fan out across an entire volume.
func NewChownCommand(f *cmdutil.Factory) *cobra.Command {
	o := &chownOptions{}
	cmd := &cobra.Command{
		Use:   "chown <remote-path> [--uid <uid>] [--recursive]",
		Short: "get / set the POSIX owner UID of a remote file or directory",
		Long: `Get or set the POSIX owner UID of a file or directory on the
per-user files-backend.

This is the CLI counterpart of the LarePass web app's "Permission"
tab in the file properties dialog. The dialog exposes exactly two
choices ("Root"=uid 0, "User"=uid 1000) plus a "Recursive" toggle;
the CLI accepts any integer uid for flexibility but those two are
the values the LarePass GUI surfaces.

Wire shape:

    GET  /api/permission/<fileType>/<extend><subPath>          → {uid: <int>}
    PUT  /api/permission/<fileType>/<extend><subPath>?uid=<int>[&recursive=1]

Modes:

    files chown <path>                  — GET; print the current uid
    files chown <path> --uid <int>      — PUT; replace the uid
    files chown <path> --uid <int> -r   — PUT; recurse into children

Supported namespaces (mirrors LarePass's ` + "`permissionInDriveType`" + `):

    drive/Home/<sub>           — the Home volume on the user's PVC
    drive/Data/<sub>           — the Data volume on the user's PVC
    cache/<node>/<sub>         — the per-node Cache volume

Other namespaces are intentionally rejected client-side:

    sync/<repo_id>/...         — Seafile permissions are managed via
                                 the libraries' own ACL surface
                                 (` + "`files repos`" + `), not POSIX uid.
    external/<node>/<volume>/  — the LarePass GUI hides the
                                 Permission tab for external mounts;
                                 the wire surface there is not part
                                 of this contract.
    awss3/dropbox/google/tencent — cloud accounts are object stores,
                                 not POSIX filesystems; ownership is
                                 not a meaningful concept.

Other refusals:

  - The volume root (` + "`drive/Home/`" + `, ` + "`drive/Data/`" + `, ` + "`cache/<node>/`" + `):
    chowning an entire namespace root is almost never the user's
    intent and the blast radius is severe. Pick a one-level-deeper
    path with -r if you want to fan out.

UID conventions in LarePass:

    0       — Root (system; only set this if you know why)
    1000    — User (the default LarePass user; matches the GUI's
              "User" preset)

Examples:

    # Inspect the current owner.
    olares-cli files chown drive/Home/Documents/foo.pdf

    # Hand a file to root.
    olares-cli files chown drive/Home/Documents/foo.pdf --uid 0

    # Hand an entire directory tree to the default user.
    olares-cli files chown drive/Home/Pictures/Trip2024/ --uid 1000 -r

    # Cache namespace.
    olares-cli files chown cache/<node>/scratch/build/ --uid 1000 -r
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			o.uidSet = cmd.Flags().Changed("uid")
			return runChown(cmd.Context(), f, cmd.OutOrStdout(), args[0], o)
		},
	}
	cmd.Flags().StringVar(&o.uidStr, "uid", "",
		"new POSIX owner UID; LarePass presets: 0=Root, 1000=User. Omit to GET the current uid.")
	cmd.Flags().BoolVarP(&o.recursive, "recursive", "r", false,
		"apply the change to every child entry as well (only meaningful with --uid)")
	cmd.Flags().BoolP("recursive-bsd", "R", false, "alias for -r")
	cmd.Flags().Lookup("recursive-bsd").Hidden = true
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		bsd, err := cmd.Flags().GetBool("recursive-bsd")
		if err == nil && bsd {
			o.recursive = true
		}
		return nil
	}
	return cmd
}

// runChown is the cobra-side glue: validate flags, resolve the path,
// build the wire client, fire either a GET or a PUT depending on
// flag presence, and reformat any HTTPError into a friendly CTA
// (same spirit as runRename / runRm / runCpMv).
func runChown(
	ctx context.Context,
	f *cmdutil.Factory,
	out io.Writer,
	pathArg string,
	o *chownOptions,
) error {
	if ctx == nil {
		ctx = context.Background()
	}

	// Without --uid we ignore -r entirely — recursion only makes
	// sense for a write. The GUI's toggle is co-located with the
	// "Submit" button on the Permission tab; surfacing the
	// inconsistency early prevents a script that meant to PUT but
	// forgot --uid from silently no-op'ing as a GET.
	if !o.uidSet && o.recursive {
		return errors.New("--recursive only applies when setting a uid; pass --uid <int> to use it, or drop --recursive to GET the current uid")
	}

	tgt, err := frontendPathToChownTarget(pathArg)
	if err != nil {
		return err
	}
	if err := requireCommonBackendVersion(ctx, f, isCommonFrontendPath(tgt.FileType, tgt.Extend)); err != nil {
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
	client := &permission.Client{
		HTTPClient: httpClient,
		BaseURL:    rp.FilesURL,
	}

	display := tgt.String()

	if !o.uidSet {
		uid, err := client.Get(ctx, tgt)
		if err != nil {
			return reformatChownHTTPErr(err, rp.OlaresID, display, "get")
		}
		fmt.Fprintf(out, "%s  uid=%d%s\n", display, uid, prettifyUID(uid))
		return nil
	}

	uid, err := parseChownUID(o.uidStr)
	if err != nil {
		return err
	}

	verb := "set uid"
	if o.recursive {
		verb = "set uid (recursive)"
	}
	fmt.Fprintf(out, "%s: %s → uid=%d%s\n", verb, display, uid, prettifyUID(uid))

	if err := client.Set(ctx, tgt, uid, o.recursive); err != nil {
		return reformatChownHTTPErr(err, rp.OlaresID, display, "set")
	}

	suffix := ""
	if o.recursive {
		suffix = " (recursive)"
	}
	fmt.Fprintf(out, "  ✓ %s  uid=%d%s%s\n", display, uid, prettifyUID(uid), suffix)
	return nil
}

// parseChownUID validates the --uid argument. The wire accepts any
// int but we reject negatives client-side: a negative POSIX UID is
// neither expressible on the wire (the server casts to uint) nor a
// meaningful LarePass concept, so a "-1" typo should fail loudly
// rather than be silently rounded.
//
// We do NOT clamp to {0, 1000} — those are the GUI's presets, not
// the protocol's bounds.
func parseChownUID(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, fmt.Errorf("--uid is empty; pass an integer (LarePass GUI uses 0 for Root and 1000 for User)")
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("--uid %q is not an integer (LarePass GUI uses 0 for Root and 1000 for User): %w", raw, err)
	}
	if n < 0 {
		return 0, fmt.Errorf("--uid must be non-negative (got %d)", n)
	}
	return n, nil
}

// prettifyUID adds a "(Root)"/"(User)" annotation for the LarePass
// preset UIDs. Returned string is empty for non-preset values so the
// regular case stays terse.
func prettifyUID(uid int) string {
	for _, p := range commonChownUIDPresets {
		if p.UID == uid {
			return " (" + p.Label + ")"
		}
	}
	return ""
}

// frontendPathToChownTarget converts a user-supplied path into the
// permission package's Target shape, applying CLI-side validation:
//
//   - The fileType must be in permission.SupportedFileTypes
//     (drive / cache). Rejecting sync / external / cloud client-side
//     keeps the error message specific (which LarePass affordance
//     would have done what the user actually wants?) instead of
//     opaque (a 404 / 500 from the server).
//   - The volume root is refused — chowning a whole `drive/Home/`,
//     `drive/Data/`, or `cache/<node>/` is almost certainly a typo.
//
// IsDirIntent is preserved from the input slash so the wire URL
// keeps the trailing '/' a directory target was typed with — the
// permission endpoint doesn't seem to care today, but matching the
// GUI byte-for-byte avoids future divergence.
func frontendPathToChownTarget(raw string) (permission.Target, error) {
	fp, err := ParseFrontendPath(raw)
	if err != nil {
		return permission.Target{}, err
	}
	if !permission.IsSupported(fp.FileType) {
		return permission.Target{}, chownNamespaceError(fp.FileType)
	}
	if strings.Trim(fp.SubPath, "/") == "" {
		return permission.Target{}, fmt.Errorf(
			"refusing to chown the root of %s/%s; pick a child path (use -r to fan out across the volume)",
			fp.FileType, fp.Extend)
	}
	return permission.Target{
		FileType:    fp.FileType,
		Extend:      fp.Extend,
		SubPath:     fp.SubPath,
		IsDirIntent: strings.HasSuffix(fp.SubPath, "/"),
	}, nil
}

// chownNamespaceError constructs the per-namespace rejection
// message. We split out reasoning per-namespace (sync vs. external
// vs. cloud) because the recovery path differs:
//
//   - sync     → suggest `files repos` for ACLs; POSIX uid is not
//                the right concept on Seafile.
//   - external → the GUI hides the Permission tab; the wire is not
//                part of this contract.
//   - cloud    → object stores have no POSIX uid; the operation is
//                conceptually meaningless.
func chownNamespaceError(fileType string) error {
	switch fileType {
	case "sync":
		return fmt.Errorf(
			"namespace %q is not supported by `files chown`; Seafile permissions live on the library itself — use `olares-cli files repos` for sync ACLs",
			fileType,
		)
	case "external":
		return fmt.Errorf(
			"namespace %q is not supported by `files chown`; the LarePass GUI hides the Permission tab for external mounts. Allowed: %s",
			fileType, permission.SupportedFileTypesList(),
		)
	case "awss3", "dropbox", "google", "tencent":
		return fmt.Errorf(
			"namespace %q is a cloud account; object stores have no POSIX uid concept. Allowed: %s",
			fileType, permission.SupportedFileTypesList(),
		)
	}
	return fmt.Errorf(
		"namespace %q is not supported by `files chown`; allowed: %s",
		fileType, permission.SupportedFileTypesList(),
	)
}

// reformatChownHTTPErr maps permission.HTTPError onto user-friendly
// messages, mirroring the rename / rm / cp counterparts. Status
// branches:
//
//   - 401/403: token rejected → suggest `profile login`. Same
//     wording as the other verbs so the user gets one consistent CTA.
//   - 404: target not found → echo the path so the user can re-try
//     against a corrected one.
//
// Other 4xx / 5xx flow through verbatim: the server's body in the
// HTTPError already includes the message and the URL the call hit,
// which is the most useful diagnostic we can produce without
// guessing.
//
// Typed credential errors from the refreshing transport are surfaced
// verbatim (same rationale as reformatRenameHTTPErr / reformatRmHTTPErr).
func reformatChownHTTPErr(err error, olaresID, display, verb string) error {
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
	var hErr *permission.HTTPError
	if errors.As(err, &hErr) {
		switch hErr.Status {
		case 401, 403:
			if olaresID != "" {
				return fmt.Errorf(
					"server rejected the access token (HTTP %d); please run: olares-cli profile login --olares-id %s",
					hErr.Status, olaresID,
				)
			}
			return fmt.Errorf(
				"server rejected the access token (HTTP %d); please re-run `olares-cli profile login`",
				hErr.Status,
			)
		case 404:
			return fmt.Errorf("chown %s %s: not found on the server (HTTP 404)", verb, display)
		}
	}
	return err
}
