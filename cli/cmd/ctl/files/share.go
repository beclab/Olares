package files

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/internal/files/download"
	"github.com/beclab/Olares/cli/internal/files/share"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/credential"
)

// NewShareCommand returns the `olares-cli files share` parent
// command, which groups together the three folder-share creation
// flavors (internal / public / smb) plus the management verbs the
// resulting share IDs need to be useful (list / get / rm).
//
// All three creation flavors converge on the same wire endpoint —
//
//	POST /api/share/share_path/<fileType>/<extend><subPath>/
//
// — with the share-type discriminator in the JSON body. The split
// into three subcommands lives at the CLI surface only, because each
// flavor has a meaningfully different flag set: internal takes a
// member list, public takes a password + expiration + upload limits,
// SMB takes a public toggle + SMB-account list. Trying to share-by-
// flag would force the user to think about the wire shape; share-by-
// subcommand keeps each flag set focused on one workflow.
//
// Management verbs (`list` / `get` / `rm`) are kept at this level
// rather than under the per-type subcommand because a share id is
// share-type-agnostic on the wire (the `/api/share/share_path/`
// surface treats every type the same way once the share exists).
func NewShareCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "share",
		Short: "create and manage internal / public / SMB shares for files-backend directories",
		Long: `Create and manage shares for directories on the per-user files-backend.

All three create flavors are DIRECTORY-ONLY: each ` + "`share <flavor>" + "`" + `
verb Stats the target before posting the create record and refuses
up front when the path is a file (or doesn't exist on the server).
This matches the LarePass GUI's per-driver share-menu gating on
` + "`event.isDir`" + ` — sharing a single file is rejected by the web app
and by the CLI. To "share a single file", place it in a dedicated
directory and share that directory instead.

Three share flavors:

    internal   Cross-user share inside Olares. Recipients are picked
               from the Olares user roster; permissions are per-member
               (view / edit / admin).

    public     External shareable link with a password and either an
               expiration window (days from now) or an explicit
               expiration time. Optionally restricted to upload-only,
               with a per-upload size cap.

    smb        Samba network share — exposes the folder over the SMB
               protocol so a desktop / Finder / Explorer can mount it.
               Recipients are picked from the SMB-account roster
               (` + "`smb-users`" + ` subcommand to list / create those), or
               toggled to "anyone on the local network" with --public.

Wire shape (all three converge on the same endpoint):

    POST /api/share/share_path/<fileType>/<extend><subPath>/
    body: {name, share_type, permission, password, ...}

The response carries the new share's id, plus per-flavor extras
(smb_link / smb_user / smb_password for SMB shares; the public-link
URL is constructed from the share id by the LarePass app's
shareBaseUrl + /sharable-link/<id>/ pattern).

Management verbs (` + "`list`" + ` / ` + "`get`" + ` / ` + "`rm`" + `) target the share id and are
share-type-agnostic.

Per-flavor update verbs (each rejects mismatched share types up
front so a typo doesn't reach the server):

    set-password   roll the access password of a Public-link share
                   (PUT /api/share/share_password/)
    set-members    REPLACE the member list of an Internal share
                   (PUT /api/share/share_path/share_members/)
    set-smb        REPLACE the SMB account list, or flip to public-
                   SMB mode (POST /api/share/smb_share_member/)

Examples:

    # Internal share with two members.
    olares-cli files share internal drive/Home/Backups/ \
        --users alice:edit,bob:view

    # Public link valid for 7 days, password auto-generated.
    olares-cli files share public drive/Home/Photos/ --expire-days 7

    # SMB share for two SMB users with read+write.
    olares-cli files share smb drive/Home/Movies/ \
        --users smb-uid-1:edit,smb-uid-2:edit

    # Roll a Public link's password.
    olares-cli files share set-password <share-id>

    # Promote bob from view to admin on an Internal share.
    olares-cli files share set-members <share-id> \
        --users alice:edit,bob:admin

    # Switch an SMB share to public-SMB.
    olares-cli files share set-smb <share-id> --public

    # List, inspect, remove.
    olares-cli files share list --shared-by-me
    olares-cli files share get <share-id>
    olares-cli files share rm <share-id> [<share-id>...]
`,
	}
	cmd.AddCommand(
		newShareInternalCommand(f),
		newSharePublicCommand(f),
		newShareSMBCommand(f),
		newShareSetPasswordCommand(f),
		newShareSetMembersCommand(f),
		newShareSetSMBCommand(f),
		newShareListCommand(f),
		newShareGetCommand(f),
		newShareRmCommand(f),
		newShareSMBUsersCommand(f),
	)
	for _, sub := range cmd.Commands() {
		// Same rationale as the top-level files command: bad-creds /
		// network / not-found errors are already actionable, don't
		// bury them under a usage dump.
		sub.SilenceUsage = true
	}
	return cmd
}

// newShareListCommand: `olares-cli files share list [--shared-by-me]
// [--shared-to-me] [--type internal,smb] [--owner alice,bob]`.
//
// Mirrors the web app's getShareList query-shape (see
// share.ts L12-25): filters are comma-joined strings on the wire so
// the user can fan out across multiple types / owners with a single
// flag.
func newShareListCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		sharedByMe bool
		sharedToMe bool
		shareType  string
		owner      string
		asJSON     bool
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list shares (filterable by direction, type, owner)",
		Long: `List shares the current user has created or has access to.

By default, both directions are listed. Pass --shared-by-me /
--shared-to-me to scope; --type / --owner filter the result set.
Filters are comma-joined: --type internal,smb returns both flavors.

Wire shape: GET /api/share/share_path/?<filters>.

Examples:

    olares-cli files share list
    olares-cli files share list --shared-by-me
    olares-cli files share list --type smb,external
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runShareList(cmd.Context(), f, cmd.OutOrStdout(),
				cmd.Flags().Changed("shared-by-me"), sharedByMe,
				cmd.Flags().Changed("shared-to-me"), sharedToMe,
				shareType, owner, asJSON)
		},
	}
	cmd.Flags().BoolVar(&sharedByMe, "shared-by-me", true, "include shares created by you (default true; pass --shared-by-me=false to exclude)")
	cmd.Flags().BoolVar(&sharedToMe, "shared-to-me", true, "include shares created by other users that you can access (default true)")
	cmd.Flags().StringVar(&shareType, "type", "", "comma-joined share types: internal,external,smb")
	cmd.Flags().StringVar(&owner, "owner", "", "comma-joined owner names")
	addOutputFormatFlag(cmd, &asJSON, "print the raw share records instead of a table")
	return cmd
}

// newShareGetCommand: `olares-cli files share get <share-id>`. Single
// lookup against /api/share/share_path/?path_id=<id>. Returns nothing
// (with a hint) if the id is unknown — handy for scripting.
func newShareGetCommand(f *cmdutil.Factory) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "get <share-id>",
		Short: "fetch one share by id",
		Long: `Fetch a single share by id (UUID assigned at creation time).

Wire shape: GET /api/share/share_path/?path_id=<id>.

If the id doesn't exist the command exits with a non-zero status and
"share not found" — useful in scripts that want to branch on absence
without parsing list output.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShareGet(cmd.Context(), f, cmd.OutOrStdout(), args[0], asJSON)
		},
	}
	addOutputFormatFlag(cmd, &asJSON, "print the raw share record instead of a key:value layout")
	return cmd
}

// newShareRmCommand: `olares-cli files share rm <share-id>...`. Batch
// delete via a single DELETE call (the wire endpoint takes a comma-
// joined list, mirroring the web app).
func newShareRmCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rm <share-id> [<share-id>...]",
		Aliases: []string{"delete", "remove"},
		Short:   "remove one or more shares by id",
		Long: `Remove one or more shares by id.

Wire shape: DELETE /api/share/share_path/?path_ids=<comma-joined-ids>.

The IDs are joined into a single DELETE call so the operation is
atomic from the caller's perspective. Removing a share doesn't
delete the shared resource itself — only the share record / link.
`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShareRm(cmd.Context(), f, cmd.OutOrStdout(), args)
		},
	}
	return cmd
}

// newShareSMBUsersCommand groups SMB-account roster verbs (list /
// create). The IDs returned by `list` are what `share smb --users`
// expects — kept under the same parent command so the workflow is
// discoverable: list users → reference their IDs in a `share smb`
// invocation.
func newShareSMBUsersCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "smb-users",
		Short: "list / create SMB accounts (referenced by `share smb --users`)",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "list SMB accounts",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSMBUsersList(cmd.Context(), f, cmd.OutOrStdout())
		},
	}
	createCmd := &cobra.Command{
		Use:   "create <name> <password>",
		Short: "create a new SMB account",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSMBUsersCreate(cmd.Context(), f, cmd.OutOrStdout(), args[0], args[1])
		},
	}
	for _, sub := range []*cobra.Command{listCmd, createCmd} {
		sub.SilenceUsage = true
		cmd.AddCommand(sub)
	}
	return cmd
}

// runShareList is the cobra-side glue for `share list`: build a
// ListParams from the flags, call the share client, render a tab-
// aligned table.
func runShareList(
	ctx context.Context,
	f *cmdutil.Factory,
	out io.Writer,
	byMeChanged bool, byMe bool,
	toMeChanged bool, toMe bool,
	shareType, owner string,
	asJSON bool,
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	client, rp, err := setupShareClient(ctx, f)
	if err != nil {
		return err
	}
	params := share.ListParams{ShareType: shareType, Owner: owner}
	// Pass the filter only when the user explicitly set it; the
	// server's default for both is "true" (return all). Sending
	// shared_by_me=true unconditionally would be redundant but
	// harmless — sending it as=false is what we want to support
	// for "only show shares I received" / "only show shares I
	// created" workflows.
	if byMeChanged {
		params.SharedByMe = &byMe
	}
	if toMeChanged {
		params.SharedToMe = &toMe
	}

	rows, err := client.List(ctx, params)
	if err != nil {
		return reformatShareHTTPErr(err, rp.OlaresID, "list shares")
	}
	if len(rows) == 0 && !asJSON {
		fmt.Fprintln(out, "no shares found")
		return nil
	}
	// Stable order so repeated invocations produce identical output —
	// the server doesn't guarantee any ordering.
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].ShareType != rows[j].ShareType {
			return rows[i].ShareType < rows[j].ShareType
		}
		return rows[i].ID < rows[j].ID
	})

	// An empty list is `[]`, not `null`: a caller piping this into jq
	// should be iterating zero rows, not handling a null.
	if asJSON {
		if rows == nil {
			rows = []share.Result{}
		}
		return writeShareJSON(out, rows)
	}

	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTYPE\tNAME\tOWNER\tPATH\tPERMISSION\tEXPIRE")
	for i := range rows {
		r := &rows[i]
		expire := r.ExpireTime
		if expire == "" {
			expire = "-"
		}
		// list trusts r.SyncRepoName (the server already echoes
		// it on every shared sync library) — calling
		// resolveShareDisplayPath here would N+1 a `repos.Get`
		// per row, which is unacceptable for `share list`.
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			r.ID, r.ShareType, r.Name, r.Owner,
			formatSharePathLine(r, ""),
			r.Permission.String(), expire)
	}
	return w.Flush()
}

// runShareGet renders one share record in a key:value layout. We keep
// the formatting hand-rolled (rather than leaning on encoding/json) so
// SMB shares' link / user / password fields stand out — that's the
// information the user usually needs from this verb.
func runShareGet(ctx context.Context, f *cmdutil.Factory, out io.Writer, shareID string, asJSON bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	client, rp, err := setupShareClient(ctx, f)
	if err != nil {
		return err
	}
	r, err := client.Query(ctx, shareID)
	if err != nil {
		return reformatShareHTTPErr(err, rp.OlaresID, "get share "+shareID)
	}
	if r == nil {
		return fmt.Errorf("share %s: not found on the server", shareID)
	}
	if asJSON {
		return writeShareJSON(out, r)
	}

	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "ID:\t%s\n", r.ID)
	fmt.Fprintf(w, "Name:\t%s\n", r.Name)
	fmt.Fprintf(w, "Type:\t%s\n", r.ShareType)
	fmt.Fprintf(w, "Owner:\t%s\n", r.Owner)
	// Single-record read: it's worth one extra /api/repos/ list
	// call to swap the bare repo_id for a human library name in
	// the Path line. resolveShareDisplayPath is a no-op on
	// non-sync namespaces and on sync records the server already
	// labelled.
	fmt.Fprintf(w, "Path:\t%s\n", resolveShareDisplayPath(ctx, f, r))
	fmt.Fprintf(w, "Permission:\t%s\n", r.Permission)
	if r.ExpireTime != "" {
		fmt.Fprintf(w, "Expires at:\t%s\n", r.ExpireTime)
	}
	if r.UploadSizeLimit > 0 {
		fmt.Fprintf(w, "Upload limit:\t%s\n", formatBytes(r.UploadSizeLimit))
	}
	if r.SMBLink != "" {
		fmt.Fprintf(w, "SMB link:\t%s\n", r.SMBLink)
	}
	if r.SMBUser != "" {
		fmt.Fprintf(w, "SMB user:\t%s\n", r.SMBUser)
	}
	if r.SMBPassword != "" {
		fmt.Fprintf(w, "SMB password:\t%s\n", r.SMBPassword)
	}
	if r.CreateTime != "" {
		fmt.Fprintf(w, "Created at:\t%s\n", r.CreateTime)
	}
	if r.UpdateTime != "" {
		fmt.Fprintf(w, "Updated at:\t%s\n", r.UpdateTime)
	}
	return w.Flush()
}

// runShareRm deletes one or more shares in a single DELETE call.
// Pre-validates the IDs aren't empty (the share package does this
// too, but rejecting here lets us point at the bad arg position).
func runShareRm(ctx context.Context, f *cmdutil.Factory, out io.Writer, ids []string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	for i, id := range ids {
		if strings.TrimSpace(id) == "" {
			return fmt.Errorf("share rm: arg %d is empty; expected a non-empty share id", i+1)
		}
	}
	client, rp, err := setupShareClient(ctx, f)
	if err != nil {
		return err
	}
	if err := client.Remove(ctx, ids); err != nil {
		return reformatShareHTTPErr(err, rp.OlaresID, "remove shares")
	}
	fmt.Fprintf(out, "removed %d share%s: %s\n", len(ids), pluralS(len(ids)), strings.Join(ids, ", "))
	return nil
}

// runSMBUsersList prints the SMB-account roster as a 2-column table.
// IDs are exactly what `share smb --users` accepts.
func runSMBUsersList(ctx context.Context, f *cmdutil.Factory, out io.Writer) error {
	if ctx == nil {
		ctx = context.Background()
	}
	client, rp, err := setupShareClient(ctx, f)
	if err != nil {
		return err
	}
	accounts, err := client.ListSMBAccounts(ctx)
	if err != nil {
		return reformatShareHTTPErr(err, rp.OlaresID, "list smb accounts")
	}
	if len(accounts) == 0 {
		fmt.Fprintln(out, "no SMB accounts; create one with `olares-cli files share smb-users create <name> <password>`")
		return nil
	}
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME")
	for _, a := range accounts {
		fmt.Fprintf(w, "%s\t%s\n", a.ID, a.Name)
	}
	return w.Flush()
}

// runSMBUsersCreate creates a new SMB account. Same wire path as the
// roster GET — server distinguishes by HTTP method.
func runSMBUsersCreate(ctx context.Context, f *cmdutil.Factory, out io.Writer, user, password string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if user == "" || password == "" {
		return errors.New("share smb-users create: both <name> and <password> are required and non-empty")
	}
	client, rp, err := setupShareClient(ctx, f)
	if err != nil {
		return err
	}
	if err := client.CreateSMBAccount(ctx, user, password); err != nil {
		return reformatShareHTTPErr(err, rp.OlaresID, "create smb account")
	}
	fmt.Fprintf(out, "created SMB account: %s\n", user)
	return nil
}

// setupShareClient bundles the boilerplate every share verb needs:
// resolve the profile, build the HTTP client, return both. Returns
// the resolved profile too so the caller can pass OlaresID into
// reformatShareHTTPErr for the friendly login CTA.
//
// Kept package-private — share is the only file that uses it.
func setupShareClient(ctx context.Context, f *cmdutil.Factory) (*share.Client, *credential.ResolvedProfile, error) {
	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return nil, nil, err
	}
	httpClient, err := f.HTTPClient(ctx)
	if err != nil {
		return nil, nil, err
	}
	return &share.Client{HTTPClient: httpClient, BaseURL: rp.FilesURL}, rp, nil
}

// shareFlavorsFor lists the share flavors a directory in
// fileType/extend can have. It is the same table as common.ShareTypesOf
// in the files backend and the shareModes of the LarePass adapters:
//
//	storage                          internal  smb  public
//	drive/Home                       yes       yes  yes
//	sync                             yes       no   no
//	external                         no        yes  no
//	drive/Data, drive/Common, cache  no        no   no
//	cloud drives                     no        no   no
//
// The backend refuses every other combination; checking here gives a
// clearer message before any request is sent.
func shareFlavorsFor(fileType, extend string) []share.Type {
	switch fileType {
	case "drive":
		if extend == "Home" {
			return []share.Type{share.TypeInternal, share.TypeSMB, share.TypePublic}
		}
	case "external":
		return []share.Type{share.TypeSMB}
	case "sync":
		return []share.Type{share.TypeInternal}
	}
	return nil
}

// shareStorageName names fileType/extend in messages: drive/Home,
// drive/Common, cache, sync ...
func shareStorageName(fileType, extend string) string {
	if fileType == "drive" {
		return "drive/" + extend
	}
	return fileType
}

// cloudFileTypes lists the cloud-drive namespaces (awss3 / google /
// dropbox / tencent) that are uniformly NOT supported by any
// share-create flavor.
//
// Why a dedicated set rather than just "any fileType not in the
// per-flavor allow-list":
//
//   - The error message for cloud rejections cites a different
//     root cause than the per-flavor mismatch (cross-account
//     credentials, not "this namespace doesn't fit this flavor"),
//     so we want a distinct branch in validateShareNamespace.
//
//   - Listing them explicitly catches the case where a future
//     flavor's allow-list accidentally permits a cloud namespace —
//     the cloud rejection then short-circuits regardless. Defense
//     in depth against a regression.
var cloudFileTypes = map[string]struct{}{
	"awss3":   {},
	"google":  {},
	"dropbox": {},
	"tencent": {},
}

// shareFlavorFriendlyName maps a wire-level [share.Type] (whose
// values are the JSON-body discriminators "internal" / "external"
// / "smb") to the user-facing CLI verb name ("internal" / "public"
// / "smb"). The wire value for Public is the historically confusing
// `"external"` string; using that verbatim in error messages would
// be misleading, hence this small translation step.
func shareFlavorFriendlyName(t share.Type) string {
	switch t {
	case share.TypeInternal:
		return "internal"
	case share.TypePublic:
		return "public"
	case share.TypeSMB:
		return "smb"
	}
	return string(t)
}

// validateShareNamespace returns nil when a directory in
// fileType/extend can have a share of the given flavor (see
// [shareFlavorsFor]), or an error that says which flavors it can have.
// Cloud drives get their own message: they take no share at all.
func validateShareNamespace(flavor share.Type, fileType, extend, displayPath string) error {
	switch flavor {
	case share.TypeInternal, share.TypeSMB, share.TypePublic:
	default:
		return fmt.Errorf("validateShareNamespace: unknown share flavor %q", string(flavor))
	}
	allowed := shareFlavorsFor(fileType, extend)
	for _, f := range allowed {
		if f == flavor {
			return nil
		}
	}
	flavorName := shareFlavorFriendlyName(flavor)

	if _, isCloud := cloudFileTypes[fileType]; isCloud {
		return fmt.Errorf(
			"refusing to create a %s share for %s: cloud namespaces "+
				"(awss3 / google / dropbox / tencent) do not support sharing through "+
				"`files share` — the share endpoints don't grant cross-cloud-account access, "+
				"and the resulting share record would point at a path that no other "+
				"Olares user has the credential to read. "+
				"If you need to share cloud-backed data, download it first "+
				"(`files download`) and re-upload it to drive/Home, then share that.",
			flavorName, displayPath)
	}

	storage := shareStorageName(fileType, extend)
	if len(allowed) == 0 {
		return fmt.Errorf("refusing to create a %s share for %s: %s cannot be shared. "+
			"Copy the data into drive/Home first.", flavorName, displayPath, storage)
	}
	var names, cmds []string
	for _, f := range allowed {
		names = append(names, shareFlavorFriendlyName(f))
		cmds = append(cmds, "`files share "+shareFlavorFriendlyName(f)+"`")
	}
	return fmt.Errorf(
		"refusing to create a %s share for %s: %s only supports %s shares "+
			"(matches the LarePass GUI). Use %s instead, or copy the data into drive/Home first.",
		flavorName, displayPath, storage, strings.Join(names, " / "), strings.Join(cmds, " or "))
}

// frontendPathToShareTarget converts a user-supplied path into the
// share package's Target, applying every share-create policy gate
// in order. Used by every share-create cobra command (internal /
// public / smb); `list` / `get` / `rm` and `smb-users` don't go
// through this helper, so the rejections below scope cleanly to
// share creation only.
//
// The flavor parameter selects which share flavor to check the
// storage against — see shareFlavorsFor. Caller passes
// share.TypeInternal / share.TypePublic / share.TypeSMB; passing
// any other value lands on the defense-in-depth branch in
// validateShareNamespace.
//
// Order of checks:
//
//  1. Path parse via [ParseFrontendPath] (catches bad fileType,
//     missing extend, drive-extend-not-Home/Data, etc.).
//
//  2. Storage / flavor check via [validateShareNamespace]. Runs
//     before the volume-listing / node-picker checks below so an
//     internal or public share against `external/<node>/` surfaces
//     the broader "external only supports smb" message rather than
//     the narrower "external/<node>/ is the volume listing layer"
//     one — same final answer, but a more actionable error.
//
//  3. `external/<node>/` (volume-listing layer) rejection. Only SMB
//     reaches it: step (2) already rejected the other flavors.
//
// Volume-root targets (e.g. `drive/Home/`, `sync/<repo>/`) are NOT
// rejected here: sharing the entire Home volume is unusual but
// legitimate, and the server is the authoritative gate.
func frontendPathToShareTarget(raw string, flavor share.Type) (share.Target, error) {
	fp, err := ParseFrontendPath(raw)
	if err != nil {
		return share.Target{}, err
	}
	if err := validateShareNamespace(flavor, fp.FileType, fp.Extend, raw); err != nil {
		return share.Target{}, err
	}
	if fp.IsExternalNodeRoot() {
		return share.Target{}, fmt.Errorf(
			"refusing to share external/%s/: this is the volume listing layer (read-only); "+
				"point at a real volume, e.g. external/%s/<volume>/<sub>/. "+
				"Use `files ls external/%s/` first to discover the attached volumes.",
			fp.Extend, fp.Extend, fp.Extend)
	}
	return share.Target{
		FileType:    fp.FileType,
		Extend:      fp.Extend,
		SubPath:     fp.SubPath,
		IsDirIntent: strings.HasSuffix(fp.SubPath, "/"),
	}, nil
}

// shareNameFromPath derives the human-readable share label from the
// last segment of a 3-segment frontend path. The web app passes
// `decodeURI(file.name)` for this — for the CLI we fall back to the
// last subPath segment, which matches what `files ls` would show.
//
// syncRepoName is the (optional) resolved repo display name for
// `sync/<repo_id>/...` paths; the cobra layer feeds it in via
// [lookupSyncRepoName]. When the path is the sync repo root
// (empty subPath), we prefer the repo name over the bare repo_id —
// otherwise the share record's `name` would be a UUID, which is
// unfriendly in `share list` and on recipients' screens. Pass ""
// for non-sync paths or when the lookup didn't resolve; in those
// cases the function falls through to the legacy "extend as name"
// behavior.
//
// Subpath cases (non-empty) are unaffected: a directory inside the
// repo gets named after the directory, regardless of namespace.
func shareNameFromPath(t share.Target, syncRepoName string) string {
	sub := strings.Trim(t.SubPath, "/")
	if sub != "" {
		if i := strings.LastIndex(sub, "/"); i >= 0 {
			return sub[i+1:]
		}
		return sub
	}
	if t.FileType == "sync" && syncRepoName != "" {
		return syncRepoName
	}
	return t.Extend
}

// lookupSyncRepoName tries to resolve a Sync (Seafile) library's
// human-readable name from its UUID via repos.Client.Get. Returns
// the empty string on any error — every caller falls back to
// displaying the repo_id when the lookup doesn't resolve, so an
// unreachable /api/repos/ endpoint or a missing repo never blocks
// a share-create / share-display call.
//
// This is intentionally fire-and-forget on the error path: the
// /api/share/ surface is independent from /api/repos/, and we'd
// rather surface a UUID-shaped name than fail the whole verb just
// because the repo lookup couldn't complete.
//
// Cost: one /api/repos/ list call (Get internally lists `mine`,
// then `share_to_me`, then `shared` until it finds the id). That's
// fine for the create-flavor commands and for single-record reads
// (`share get`, the `set-*` update verbs); list-style commands
// must NOT call this per row to avoid an N+1 explosion — they
// trust the server's r.SyncRepoName field instead.
func lookupSyncRepoName(ctx context.Context, f *cmdutil.Factory, repoID string) string {
	if repoID == "" {
		return ""
	}
	client, _, err := setupReposClient(ctx, f)
	if err != nil {
		return ""
	}
	repo, err := client.Get(ctx, repoID)
	if err != nil || repo == nil {
		return ""
	}
	return repo.RepoName
}

// formatSharePathLine renders the user-facing path of a share record
// for `share list` / `share get` / the `set-*` update verbs.
//
// For sync namespace the wire-level <extend> is the repo's UUID,
// which is too noisy for human consumption — we prefer the
// resolved name (server-supplied or looked up) and append the
// repo_id in parentheses so the user can still cross-reference the
// underlying repo (e.g. for `repos rename` / `repos rm`).
//
// Falls back to the wire shape (FileType + Extend + Path) for
// every other namespace, AND for sync records where neither the
// server-supplied SyncRepoName nor the optional override resolved.
//
// override takes precedence over r.SyncRepoName: the get / update
// callers do a [lookupSyncRepoName] when r.SyncRepoName is empty
// and pass the resolved name in via override; list does NOT
// override (per-row repos lookups would be an N+1).
func formatSharePathLine(r *share.Result, override string) string {
	if r == nil {
		return ""
	}
	base := r.FileType + "/" + r.Extend + r.Path
	if r.FileType != "sync" {
		return base
	}
	name := override
	if name == "" {
		name = r.SyncRepoName
	}
	if name == "" {
		return base
	}
	return r.FileType + "/" + name + r.Path + "  (repo " + r.Extend + ")"
}

// shareTargetDisplay renders the same friendly path form as
// [formatSharePathLine] but from a share.Target (the shape the
// cobra layer holds BEFORE the share record exists). Used by the
// create-flavor commands so the "created share" output reads like
// a list / get row rather than echoing the raw UUID the user
// typed in.
func shareTargetDisplay(t share.Target, syncRepoName string) string {
	base := t.FileType + "/" + t.Extend + t.SubPath
	if t.FileType != "sync" || syncRepoName == "" {
		return base
	}
	return t.FileType + "/" + syncRepoName + t.SubPath + "  (repo " + t.Extend + ")"
}

// resolveShareDisplayPath is the single-record convenience for
// get / update verbs: it tries r.SyncRepoName first (the server's
// echo, free) and falls back to a one-shot [lookupSyncRepoName]
// when the field is empty. List-style commands skip the fallback
// to avoid N+1 queries; they call formatSharePathLine directly
// with override="".
func resolveShareDisplayPath(ctx context.Context, f *cmdutil.Factory, r *share.Result) string {
	if r == nil {
		return ""
	}
	if r.FileType != "sync" {
		return formatSharePathLine(r, "")
	}
	name := r.SyncRepoName
	if name == "" {
		name = lookupSyncRepoName(ctx, f, r.Extend)
	}
	return formatSharePathLine(r, name)
}

// reformatShareHTTPErr maps share.HTTPError / download.HTTPError
// onto user-friendly messages — same pattern as cp / rm / download.
// The op string describes which verb hit the error so multiple
// verbs in one session can be told apart in error logs.
//
// Typed credential errors from the refreshing transport are surfaced
// verbatim; see reformatHTTPErr in download.go for the rationale.
//
// Two error types because the share-create cobra flow now goes
// through TWO packages:
//   - share.Client.{Create,...} for the share record CRUD
//     (share.HTTPError);
//   - download.Client.Stat for the preflight existence / dir-only
//     check that runs before Create (download.HTTPError).
//
// Status code handling is identical for both — we want the user to
// see the same `profile login` CTA on 401/403 regardless of which
// leg of the operation failed — so we collapse the two into one
// status integer and run a single switch.
//
// The 459 status is a back-compat artifact of the LarePass auth
// proxy (legacy "token invalidated" indicator that pre-dates the
// refreshing transport's typed errors). It only ever comes from
// share.HTTPError; download.HTTPError doesn't model it explicitly,
// but the switch arm handles either source uniformly because
// preflight 401/403 already gets the same treatment.
//
// The 409 arm is share.HTTPError-only by design — it describes a
// share-record state conflict (existing share id, double-create),
// which can only originate from Create/Add/Update calls. The
// preflight's Stat path never produces 409, so the arm is
// effectively `share.HTTPError`-scoped without needing an extra
// type guard.
func reformatShareHTTPErr(err error, olaresID, op string) error {
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
	var status int
	var shareErr *share.HTTPError
	if errors.As(err, &shareErr) {
		status = shareErr.Status
	}
	var dlErr *download.HTTPError
	if status == 0 && errors.As(err, &dlErr) {
		status = dlErr.Status
	}
	switch status {
	case 401, 403, 459:
		return fmt.Errorf("%s: %w", op, credential.FormatHTTPAuthError(status, nil, olaresID))
	case 404:
		return fmt.Errorf("%s: not found on the server (HTTP 404)", op)
	case 409:
		return fmt.Errorf("%s: server reported a conflict (HTTP 409); the resource may already be shared, or the share id is in use", op)
	}
	return err
}

// pluralS handles "share" / "shares" — same micro-helper pattern as
// pluralYies / pluralEs in the upload / rm files.
func pluralS(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// requireShareType is the share-type check every share-update verb
// runs after fetching the share record by id. It returns nil when
// the record's wire-level share_type matches `want`, and a self-
// describing error otherwise.
//
// The `verb` argument is interpolated verbatim into the rejection
// message (e.g. "set the password of"). Pass an action phrase that
// reads naturally after "refusing to ", since that's how the
// composed error sentence starts.
//
// Why pull this out of each cobra layer:
//
//   - Keeps the per-verb run-funcs trivial — they do a Query, hand
//     the result to requireShareType, and bail on error.
//
//   - The mismatch error is the same shape across all three update
//     verbs (set-password / set-members / set-smb), and centralizing
//     the wording makes it easy to keep them in sync if the recovery
//     hint ever needs to change.
//
//   - It's a pure function (string in → error out), so unit tests
//     don't need to stand up a share.Client / fake server.
//
// The friendly-name translation through shareFlavorFriendlyName is
// load-bearing: the wire value for Public is `"external"`, but the
// CLI verb is `share public` / `share set-password`, so the error
// must say "public" or the user will hunt for a non-existent
// `share external` command.
func requireShareType(actual *share.Result, want share.Type, verb, shareID string) error {
	if actual == nil {
		return fmt.Errorf("share %s: not found on the server", shareID)
	}
	if actual.ShareType == want {
		return nil
	}
	gotFriendly := shareFlavorFriendlyName(actual.ShareType)
	wantFriendly := shareFlavorFriendlyName(want)
	return fmt.Errorf(
		"refusing to %s share %s: the share is %s (wire type %q), not %s; "+
			"use the matching update verb instead — `share set-password` for public shares, "+
			"`share set-members` for internal shares, `share set-smb` for SMB shares",
		verb, shareID, gotFriendly, string(actual.ShareType), wantFriendly)
}

// writeShareJSON prints a share record the way the server describes it.
// `permission` is the wire integer rather than the "view" / "edit" label
// the tables print: the label is a rendering, and a caller reading JSON
// is going to feed the value back to a flag that takes the number.
func writeShareJSON(out io.Writer, v any) error {
	encoded, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("render share record as JSON: %w", err)
	}
	_, err = fmt.Fprintln(out, string(encoded))
	return err
}
