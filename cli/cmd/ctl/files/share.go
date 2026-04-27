package files

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

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
		Short: "create and manage internal / public / SMB shares for files-backend resources",
		Long: `Create and manage shares for files / folders on the per-user files-backend.

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

Examples:

    # Internal share with two members.
    olares-cli files share internal drive/Home/Backups/ \
        --users alice:edit,bob:view

    # Public link valid for 7 days, password auto-generated.
    olares-cli files share public drive/Home/Photos/ --expire-days 7

    # SMB share for two SMB users with read+write.
    olares-cli files share smb drive/Home/Movies/ \
        --users smb-uid-1:edit,smb-uid-2:edit

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
				shareType, owner)
		},
	}
	cmd.Flags().BoolVar(&sharedByMe, "shared-by-me", true, "include shares created by you (default true; pass --shared-by-me=false to exclude)")
	cmd.Flags().BoolVar(&sharedToMe, "shared-to-me", true, "include shares created by other users that you can access (default true)")
	cmd.Flags().StringVar(&shareType, "type", "", "comma-joined share types: internal,external,smb")
	cmd.Flags().StringVar(&owner, "owner", "", "comma-joined owner names")
	return cmd
}

// newShareGetCommand: `olares-cli files share get <share-id>`. Single
// lookup against /api/share/share_path/?path_id=<id>. Returns nothing
// (with a hint) if the id is unknown — handy for scripting.
func newShareGetCommand(f *cmdutil.Factory) *cobra.Command {
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
			return runShareGet(cmd.Context(), f, cmd.OutOrStdout(), args[0])
		},
	}
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
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	client, _, err := setupShareClient(ctx, f)
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
		return reformatShareHTTPErr(err, "", "list shares")
	}
	if len(rows) == 0 {
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

	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTYPE\tNAME\tOWNER\tPATH\tPERMISSION\tEXPIRE")
	for _, r := range rows {
		expire := r.ExpireTime
		if expire == "" {
			expire = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s/%s%s\t%s\t%s\n",
			r.ID, r.ShareType, r.Name, r.Owner,
			r.FileType, r.Extend, r.Path,
			r.Permission.String(), expire)
	}
	return w.Flush()
}

// runShareGet renders one share record in a key:value layout. We keep
// the formatting hand-rolled (rather than leaning on encoding/json) so
// SMB shares' link / user / password fields stand out — that's the
// information the user usually needs from this verb.
func runShareGet(ctx context.Context, f *cmdutil.Factory, out io.Writer, shareID string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	client, _, err := setupShareClient(ctx, f)
	if err != nil {
		return err
	}
	r, err := client.Query(ctx, shareID)
	if err != nil {
		return reformatShareHTTPErr(err, "", "get share "+shareID)
	}
	if r == nil {
		return fmt.Errorf("share %s: not found on the server", shareID)
	}

	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "ID:\t%s\n", r.ID)
	fmt.Fprintf(w, "Name:\t%s\n", r.Name)
	fmt.Fprintf(w, "Type:\t%s\n", r.ShareType)
	fmt.Fprintf(w, "Owner:\t%s\n", r.Owner)
	fmt.Fprintf(w, "Path:\t%s/%s%s\n", r.FileType, r.Extend, r.Path)
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
	client, _, err := setupShareClient(ctx, f)
	if err != nil {
		return err
	}
	if err := client.Remove(ctx, ids); err != nil {
		return reformatShareHTTPErr(err, "", "remove shares")
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
	client, _, err := setupShareClient(ctx, f)
	if err != nil {
		return err
	}
	accounts, err := client.ListSMBAccounts(ctx)
	if err != nil {
		return reformatShareHTTPErr(err, "", "list smb accounts")
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
	client, _, err := setupShareClient(ctx, f)
	if err != nil {
		return err
	}
	if err := client.CreateSMBAccount(ctx, user, password); err != nil {
		return reformatShareHTTPErr(err, "", "create smb account")
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

// frontendPathToShareTarget converts a user-supplied path into the
// share package's Target. Same shape as cp / rename's converters; we
// don't reject the volume root here because (a) sharing the root of
// drive/Home is an unusual but legitimate use case, and (b) the
// server is the authoritative gate for "can this be shared".
func frontendPathToShareTarget(raw string) (share.Target, error) {
	fp, err := ParseFrontendPath(raw)
	if err != nil {
		return share.Target{}, err
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
// Empty subPath (extend root) falls back to the extend itself, so
// e.g. sharing `drive/Home/` yields name="Home", which is the
// least-surprising default.
func shareNameFromPath(t share.Target) string {
	sub := strings.Trim(t.SubPath, "/")
	if sub == "" {
		return t.Extend
	}
	if i := strings.LastIndex(sub, "/"); i >= 0 {
		return sub[i+1:]
	}
	return sub
}

// reformatShareHTTPErr maps share.HTTPError onto user-friendly
// messages — same pattern as cp / rm / download. The op string
// describes which verb hit the error so multiple verbs in one
// session can be told apart in error logs.
//
// Typed credential errors from the refreshing transport are surfaced
// verbatim; see reformatHTTPErr in download.go for the rationale.
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
	var hErr *share.HTTPError
	if errors.As(err, &hErr) {
		switch hErr.Status {
		case 401, 403:
			if olaresID != "" {
				return fmt.Errorf("server rejected the access token (HTTP %d) during %s; please run: olares-cli profile login --olares-id %s",
					hErr.Status, op, olaresID)
			}
			return fmt.Errorf("server rejected the access token (HTTP %d) during %s; please re-run `olares-cli profile login`",
				hErr.Status, op)
		case 404:
			return fmt.Errorf("%s: not found on the server (HTTP 404)", op)
		case 409:
			return fmt.Errorf("%s: server reported a conflict (HTTP 409); the resource may already be shared, or the share id is in use", op)
		}
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
