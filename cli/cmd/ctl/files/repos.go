package files

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"syscall"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/beclab/Olares/cli/internal/files/repos"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/credential"
	"github.com/beclab/Olares/cli/pkg/utils"
)

// NewReposCommand returns the `olares-cli files repos` parent
// command, which surfaces the per-user files-backend's catalog of
// Sync (Seafile) libraries.
//
// Why a top-level verb rather than a subcommand under, say, `files
// share` or `files ls`:
//
//   - The `<repo_id>` is what the rest of the CLI types into the
//     `<extend>` segment of `sync/<repo_id>/<sub>` — without a way
//     to enumerate IDs from the CLI, the user has to copy them out
//     of the LarePass web app every time. Surfacing repos as a
//     first-class verb keeps the discovery loop in-CLI.
//   - The Sync repo catalog is its own concept (it's the only
//     fileType whose `<extend>` is a server-assigned UUID rather
//     than a user-typed name). The other fileTypes don't need a
//     symmetric `cache repos` / `external repos` because their
//     `<extend>` values come from `/api/nodes/` (which `files
//     upload --node` already exposes) or from URLs the user
//     knows.
//
// Verbs:
//
//	list     enumerate the user's repos (mine / share-to-me /
//	         shared / all), with --json for scripting
//	get      fetch a single repo's metadata by id
//	create   provision a new (unencrypted) Sync library
//	rename   change the display name of a repo (id stays stable)
//	rm       delete a repo (irreversible from the CLI)
//
// Encryption / unlock is intentionally NOT exposed here: the
// per-user files-backend's createLibrary endpoint has no
// password / encryption flag, and the LarePass UI doesn't expose
// one either. Encrypted libraries must be created from the LarePass
// app or directly via Seahub.
func NewReposCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repos",
		Short: "list the user's Sync (Seafile) repositories",
		Long: `List the per-user files-backend's Sync (Seafile) repositories.

A "repo" (in Seafile parlance, a "library") is the unit of storage
that backs the ` + "`sync/<repo_id>/<sub>`" + ` paths every other ` + "`files`" + ` verb
accepts: ` + "`ls`" + `, ` + "`download`" + `, ` + "`upload`" + `, ` + "`cp`" + `, ` + "`mv`" + `, ` + "`rename`" + `, ` + "`rm`" + `,
` + "`share`" + `. Each repo has a stable UUID (` + "`repo_id`" + `) that becomes the
` + "`<extend>`" + ` segment, and a mutable display name the LarePass UI
shows.

Three flavors are addressable:

    mine          libraries you own (default; matches the LarePass
                  "My Libraries" group).
    share-to-me   libraries other users have shared with you.
    shared        libraries you have shared with other users.

Wire shape:

    GET /api/repos/                       → mine
    GET /api/repos/?type=share_to_me      → share-to-me
    GET /api/repos/?type=shared           → shared

Examples:

    # List your own libraries.
    olares-cli files repos list

    # All three flavors at once.
    olares-cli files repos list --type all

    # JSON for scripts.
    olares-cli files repos list --json

    # Create a new library, then list to confirm.
    olares-cli files repos create "Project Alpha"
    olares-cli files repos list

    # Rename the freshly created repo (the repo id stays stable).
    olares-cli files repos rename <repo-id> "Project Alpha (archived)"

    # Tear it down — destructive, no client-side undo.
    olares-cli files repos rm <repo-id>
`,
	}
	cmd.AddCommand(newReposListCommand(f))
	cmd.AddCommand(newReposGetCommand(f))
	cmd.AddCommand(newReposCreateCommand(f))
	cmd.AddCommand(newReposRenameCommand(f))
	cmd.AddCommand(newReposRmCommand(f))
	for _, sub := range cmd.Commands() {
		// Same rationale as the rest of the files surface: keep the
		// usage dump out of the way when the error is already
		// actionable (auth failure, network drop, repo not found).
		sub.SilenceUsage = true
	}
	return cmd
}

// reposListOptions bundles the flags `repos list` understands. Pulled
// into a struct so RunE stays one line and the flag wiring is easy to
// audit.
type reposListOptions struct {
	// kind is the filter from --type: "mine" / "share-to-me" /
	// "shared" / "all". The empty string defaults to "mine" (matches
	// the web app's behavior when no type is selected).
	kind string
	// asJSON renders the raw repo records as a pretty-printed JSON
	// array, suitable for scripts (jq, etc.). Otherwise the output
	// is a tab-aligned table.
	asJSON bool
}

// newReposListCommand: `olares-cli files repos list [--type ...]
// [--json]`. Defaults to "mine", which matches the web app's left-
// nav "My Libraries" view.
func newReposListCommand(f *cmdutil.Factory) *cobra.Command {
	o := &reposListOptions{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list Sync repositories (default: repos you own)",
		Long: `List Sync repositories from the per-user files-backend.

By default lists the repos you own (the "My Libraries" group in the
LarePass UI). Pass --type to fan out:

    --type mine           (default) repos you own
    --type share-to-me    repos others have shared with you
    --type shared         repos you have shared out
    --type all            all three flavors, concatenated

The output table includes ` + "`REPO_ID`" + `, which is the value to put in
the ` + "`<extend>`" + ` segment of any other ` + "`files`" + ` command (e.g.
` + "`files ls sync/<REPO_ID>/`" + `).

Examples:

    olares-cli files repos list
    olares-cli files repos list --type all
    olares-cli files repos list --type share-to-me --json
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runReposList(cmd.Context(), f, cmd.OutOrStdout(), o)
		},
	}
	cmd.Flags().StringVar(&o.kind, "type", "mine",
		"filter: mine | share-to-me | shared | all")
	cmd.Flags().BoolVar(&o.asJSON, "json", false,
		"print raw JSON instead of a table")
	return cmd
}

// newReposGetCommand: `olares-cli files repos get <repo_id>`. Uses
// the repos package's fan-out filter to find a single repo across all
// three flavors. Useful for scripts that resolve a name → id pair
// after listing.
func newReposGetCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <repo_id>",
		Short: "fetch one repo by id (across all three flavors)",
		Long: `Fetch a single Sync repo by id.

Searches the three flavors (mine → share-to-me → shared) and returns
the first match. Exits non-zero with "repo not found" if the id
doesn't appear in any list — useful for scripts that want to branch
on absence without parsing list output.

Wire shape: same as ` + "`repos list`" + ` — the lookup is a client-side
filter over the same /api/repos/ responses (the per-user files-
backend doesn't expose a single-repo GET).
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReposGet(cmd.Context(), f, cmd.OutOrStdout(), args[0])
		},
	}
	return cmd
}

// runReposList is the cobra-side glue for `repos list`.
//
// Special-cases --type=all to a fan-out across the three flavors so
// the user sees mine + share-to-me + shared in a single table. The
// repos package surfaces a List + ListAll pair so the CLI can pick
// the right one without re-implementing the fan-out.
func runReposList(ctx context.Context, f *cmdutil.Factory, out io.Writer, o *reposListOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	client, rp, err := setupReposClient(ctx, f)
	if err != nil {
		return err
	}

	var (
		rows []repos.Repo
		all  bool
	)
	switch o.kind {
	case "all", "*":
		all = true
		rows, err = client.ListAll(ctx)
	default:
		kind, perr := repos.ParseType(o.kind)
		if perr != nil {
			return perr
		}
		rows, err = client.List(ctx, kind)
	}
	if err != nil {
		return reformatReposHTTPErr(err, rp.OlaresID, "list repos")
	}

	if o.asJSON {
		return writeJSON(out, rows)
	}
	if len(rows) == 0 {
		fmt.Fprintln(out, "no repos found")
		return nil
	}
	// Stable sort: type group first (mine before shared variants
	// when --type=all), then repo name. Without this the server
	// occasionally re-orders rows between calls, which is annoying
	// for diff-friendly scripting.
	sort.SliceStable(rows, func(i, j int) bool {
		if all && rows[i].Type != rows[j].Type {
			return rows[i].Type < rows[j].Type
		}
		if rows[i].RepoName != rows[j].RepoName {
			return rows[i].RepoName < rows[j].RepoName
		}
		return rows[i].RepoID < rows[j].RepoID
	})
	return renderReposTable(out, rows, all)
}

// runReposGet renders a single repo's fields in a key:value layout.
// Keeping this hand-rolled (vs. shoving everything through a generic
// pretty-printer) makes the SHARE_PERMISSION / OWNER fields stand out
// for shared repos, which is the typical reason someone calls `get`.
func runReposGet(ctx context.Context, f *cmdutil.Factory, out io.Writer, repoID string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	client, rp, err := setupReposClient(ctx, f)
	if err != nil {
		return err
	}
	r, err := client.Get(ctx, repoID)
	if err != nil {
		return reformatReposHTTPErr(err, rp.OlaresID, "get repo "+repoID)
	}
	if r == nil {
		return fmt.Errorf("repo %s: not found in any of mine / share-to-me / shared", repoID)
	}
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Repo ID:\t%s\n", r.RepoID)
	fmt.Fprintf(w, "Name:\t%s\n", r.RepoName)
	if r.OwnerName != "" || r.OwnerEmail != "" {
		fmt.Fprintf(w, "Owner:\t%s\n", joinNonEmpty(r.OwnerName, r.OwnerEmail))
	}
	if r.Permission != "" {
		fmt.Fprintf(w, "Permission:\t%s\n", r.Permission)
	}
	if r.SharePermission != "" {
		fmt.Fprintf(w, "Share permission:\t%s\n", r.SharePermission)
	}
	if r.ShareType != "" {
		fmt.Fprintf(w, "Share type:\t%s\n", r.ShareType)
	}
	if r.UserName != "" || r.UserEmail != "" {
		fmt.Fprintf(w, "Counterparty:\t%s\n", joinNonEmpty(r.UserName, r.UserEmail))
	}
	fmt.Fprintf(w, "Encrypted:\t%t\n", bool(r.Encrypted))
	if r.Size > 0 {
		fmt.Fprintf(w, "Size:\t%s\n", utils.FormatBytes(int64(r.Size)))
	}
	if r.LastModified != "" {
		fmt.Fprintf(w, "Last modified:\t%s\n", r.LastModified)
	}
	if r.Status != "" {
		fmt.Fprintf(w, "Status:\t%s\n", r.Status)
	}
	if err := w.Flush(); err != nil {
		return err
	}
	// Trailing usage hint — most users hit `get` to find the path
	// they should pass to other verbs, so spell it out.
	fmt.Fprintf(out, "\nuse with: olares-cli files ls sync/%s/\n", r.RepoID)
	return nil
}

// renderReposTable prints the rows as a tab-aligned table. The
// columns are deliberately compact so wide terminals don't waste
// real estate but every cell still fits on one line for typical
// repo names. When `all` is true (i.e. --type=all) we add a TYPE
// column so the user can tell mine / shared apart at a glance.
func renderReposTable(out io.Writer, rows []repos.Repo, all bool) error {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	if all {
		fmt.Fprintln(w, "REPO_ID\tNAME\tTYPE\tPERMISSION\tOWNER\tSIZE\tMODIFIED\tENC")
	} else {
		fmt.Fprintln(w, "REPO_ID\tNAME\tPERMISSION\tOWNER\tSIZE\tMODIFIED\tENC")
	}
	for _, r := range rows {
		perm := r.Permission
		if perm == "" {
			perm = r.SharePermission
		}
		if perm == "" {
			perm = "-"
		}
		owner := joinNonEmpty(r.OwnerName, r.OwnerEmail)
		if owner == "" {
			owner = joinNonEmpty(r.UserName, r.UserEmail)
		}
		if owner == "" {
			owner = "-"
		}
		size := "-"
		if r.Size > 0 {
			size = utils.FormatBytes(int64(r.Size))
		}
		modified := r.LastModified
		if modified == "" {
			modified = "-"
		}
		enc := "no"
		if bool(r.Encrypted) {
			enc = "yes"
		}
		if all {
			typ := r.Type
			if typ == "" {
				typ = "-"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				r.RepoID, r.RepoName, typ, perm, owner, size, modified, enc)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				r.RepoID, r.RepoName, perm, owner, size, modified, enc)
		}
	}
	return w.Flush()
}

// writeJSON renders rows as a pretty-printed JSON array, suitable
// for `--json` consumers (typically `jq` pipelines). We use the same
// indentation the rest of the CLI uses (2 spaces) so output stays
// uniform across verbs.
func writeJSON(out io.Writer, rows []repos.Repo) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(rows)
}

// setupReposClient bundles the boilerplate every repos verb needs:
// resolve the profile, build the HTTP client, return both. Returns
// the resolved profile too so the caller can pass OlaresID into
// reformatReposHTTPErr for the friendly login CTA.
//
// Same shape as setupShareClient — kept package-private and
// duplicated rather than factored into a generic helper because the
// return type differs per package (each internal/files/* package
// owns its own Client / HTTPError types to keep its public surface
// leak-free).
func setupReposClient(ctx context.Context, f *cmdutil.Factory) (*repos.Client, *credential.ResolvedProfile, error) {
	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return nil, nil, err
	}
	httpClient, err := f.HTTPClient(ctx)
	if err != nil {
		return nil, nil, err
	}
	return &repos.Client{HTTPClient: httpClient, BaseURL: rp.FilesURL}, rp, nil
}

// reformatReposHTTPErr maps repos.HTTPError onto user-friendly
// messages — same pattern as cp / share / download. The op string
// describes which verb hit the error so multiple verbs in one
// session can be told apart in error logs.
//
// Typed credential errors from the refreshing transport are surfaced
// verbatim; see reformatHTTPErr in download.go for the rationale.
func reformatReposHTTPErr(err error, olaresID, op string) error {
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
	var hErr *repos.HTTPError
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
		}
	}
	return err
}

// joinNonEmpty renders "<name> (<email>)" / "<email>" / "<name>"
// depending on which of the two are populated. Used for the OWNER
// and COUNTERPARTY columns where either field may be missing for a
// given repo flavor.
func joinNonEmpty(name, email string) string {
	switch {
	case name != "" && email != "":
		return name + " (" + email + ")"
	case name != "":
		return name
	default:
		return email
	}
}

// newReposCreateCommand: `olares-cli files repos create <name>`.
//
// Creates a new (unencrypted) Sync library. The library `<name>` is
// the human-readable display name only — the server picks the
// `<repo_id>` UUID. We print both at the end so the user can
// immediately pipe the id into other verbs.
//
// Wire shape: POST /api/repos/?repoName=<name>. See the
// internal/files/repos.Client.Create doc for the full rationale and
// the Seahub-side caveats (encryption is not supported here).
func newReposCreateCommand(f *cmdutil.Factory) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "create a new (unencrypted) Sync repository",
		Long: `Create a new Sync (Seafile) library.

` + "`<name>`" + ` is the human-readable display label; the server picks
the stable ` + "`<repo_id>`" + ` UUID and returns it in the response. The
new repo is unencrypted — encrypted libraries must be provisioned
from the LarePass app (the per-user files-backend's createLibrary
endpoint accepts no password / encryption parameters).

Wire shape:

    POST /api/repos/?repoName=<name>

Output is the new repo id + name (or the full record with --json),
ready to be piped into ` + "`files ls sync/<repo_id>/`" + ` etc.

Examples:

    olares-cli files repos create "Project Alpha"
    REPO_ID=$(olares-cli files repos create "Project Alpha" --json | jq -r .repo_id)
    olares-cli files ls sync/$REPO_ID/
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReposCreate(cmd.Context(), f, cmd.OutOrStdout(), args[0], asJSON)
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false,
		"print the new repo as a JSON object instead of two human-readable lines")
	return cmd
}

// newReposRenameCommand: `olares-cli files repos rename <repo_id>
// <newName>`.
//
// Renames a repo's display label. The repo's UUID is stable across
// renames, so any cached `sync/<repo_id>/...` paths keep working —
// only the LarePass UI label changes.
//
// We deliberately accept the new name positionally rather than
// behind a `--name` flag because (a) the verb is binary by nature
// (id + new label) and (b) two positional args reads naturally:
// `repos rename old-id "new label"`.
func newReposRenameCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rename <repo_id> <new_name>",
		Short: "rename a Sync repository (repo_id stays stable)",
		Long: `Rename a Sync (Seafile) library.

The repo's UUID stays the same — only the display name changes — so
already-cached ` + "`sync/<repo_id>/...`" + ` front-end paths keep working
across the rename.

Wire shape:

    PATCH /api/repos/?destination=<new-name>&repoId=<repo-id>

Examples:

    olares-cli files repos rename abc-123 "Project Alpha (archived)"
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReposRename(cmd.Context(), f, cmd.OutOrStdout(), args[0], args[1])
		},
	}
	return cmd
}

// newReposRmCommand: `olares-cli files repos rm <repo_id>... [-y]`.
//
// Same confirmation model as `files rm` (see cmd/ctl/files/rm.go): on a
// TTY, list the target repo ids (with best-effort display names) and
// ask y/N. In a non-TTY context (CI, heredoc, pipe) the user must pass
// --yes / -y (or -f / --force, the same `files rm -f` opt-in) so
// scripts don't delete blindly.
func newReposRmCommand(f *cmdutil.Factory) *cobra.Command {
	var assumeYes bool
	cmd := &cobra.Command{
		Use:     "rm <repo_id>...",
		Aliases: []string{"delete", "remove"},
		Short:   "delete one or more Sync repositories (irreversible)",
		Long: `Delete one or more Sync (Seafile) libraries.

Destructive: removes the repo and all of its contents. The Seafile
deployment may keep the data in a server-side trash window, but the
CLI does not expose a restore verb — recovery requires the LarePass
app or direct Seahub access.

Multiple ids may be passed; each one is deleted in turn and the
command continues on per-id failure (other ids still get a chance).
The exit code is non-zero if ANY deletion failed.

Confirmation (same spirit as files rm):

  - In an interactive shell (TTY on stdin), a list of target repos
    and "proceed with repo deletion? [y/N]:" are shown first.
  - In a non-TTY context (automation, CI), you must pass --yes / -y
    or -f / --force so the deletion is explicit in the command line.

Wire shape:

    DELETE /api/repos/?repoId=<repo-id>

Examples:

    olares-cli files repos rm abc-123              # y/N in a terminal
    olares-cli files repos rm abc-123 -y
    olares-cli files repos rm abc-123 def-456 -f
`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReposRm(cmd.Context(), f, cmd.OutOrStdout(), os.Stdin, args, assumeYes)
		},
	}
	// -y/--yes and -f/--force are aliases, mirroring the files rm -f
	// idiom: both set the same bool; either flag alone opts out of
	// the y/N prompt.
	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false,
		"skip the confirmation prompt (required when stdin is not a TTY)")
	cmd.Flags().BoolVarP(&assumeYes, "force", "f", false,
		"alias for --yes, same as `files rm -f`")
	return cmd
}

// runReposCreate is the cobra-side glue for `repos create`. The
// happy path prints two human-readable lines (id + name) so a
// follow-up verb can be typed immediately, or the full repo record
// as JSON when --json is set.
func runReposCreate(ctx context.Context, f *cmdutil.Factory, out io.Writer, name string, asJSON bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	client, rp, err := setupReposClient(ctx, f)
	if err != nil {
		return err
	}
	repo, err := client.Create(ctx, name)
	if err != nil {
		return reformatReposHTTPErr(err, rp.OlaresID, "create repo "+name)
	}
	if asJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(repo)
	}
	fmt.Fprintf(out, "created repo: %s (id: %s)\n", repo.RepoName, repo.RepoID)
	fmt.Fprintf(out, "use with: olares-cli files ls sync/%s/\n", repo.RepoID)
	return nil
}

// runReposRename is the cobra-side glue for `repos rename`. We
// echo the rename ("<id>: <old?> -> <new>") so the user has a
// single-line audit trail in shell history. The "old name" is
// best-effort: we try a `Get` first to fetch it, but skip if that
// lookup fails so a rename never fails because of a flaky list call.
func runReposRename(ctx context.Context, f *cmdutil.Factory, out io.Writer, repoID, newName string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	client, rp, err := setupReposClient(ctx, f)
	if err != nil {
		return err
	}
	var oldName string
	if existing, getErr := client.Get(ctx, repoID); getErr == nil && existing != nil {
		oldName = existing.RepoName
	}
	if err := client.Rename(ctx, repoID, newName); err != nil {
		return reformatReposHTTPErr(err, rp.OlaresID, "rename repo "+repoID)
	}
	if oldName != "" {
		fmt.Fprintf(out, "renamed repo %s: %q -> %q\n", repoID, oldName, newName)
	} else {
		fmt.Fprintf(out, "renamed repo %s -> %q\n", repoID, newName)
	}
	return nil
}

// runReposRm is the cobra-side glue for `repos rm`. We loop instead
// of bailing on first failure so a batch ("rm a b c") can make
// partial progress — same behavior as `files rm`. The exit code is
// driven by the joined error returned at the end.
func runReposRm(ctx context.Context, f *cmdutil.Factory, out io.Writer, in io.Reader, repoIDs []string, assumeYes bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	// Defer the HTTP client until we know we are allowed to run:
	// in non-interactive mode without -y, fail before touching the
	// network (mirrors `files rm` and its --force check).
	if !assumeYes {
		if !term.IsTerminal(int(syscall.Stdin)) {
			return errors.New("repos rm: refusing to delete without -y / --yes (or -f / --force) in a non-interactive context (no TTY)")
		}
	}
	// Filter empty IDs up front so the prompt loop and the delete
	// loop agree on what's about to happen. Previously the prompt
	// silently skipped empties while the delete loop produced an
	// error per empty entry, so a user who confirmed deletion of
	// the IDs they SAW could still get a non-zero exit driven by
	// entries that never appeared in the prompt.
	var errs []error
	validIDs := make([]string, 0, len(repoIDs))
	for _, id := range repoIDs {
		if id == "" {
			errs = append(errs, errors.New("empty repo id in argument list"))
			continue
		}
		validIDs = append(validIDs, id)
	}
	if len(validIDs) == 0 {
		return fmt.Errorf("repos rm: no valid repo ids supplied: %w", errors.Join(errs...))
	}

	client, rp, err := setupReposClient(ctx, f)
	if err != nil {
		return err
	}
	if !assumeYes {
		// Resolve display names with a single ListAll instead of one
		// Client.Get per ID. Get fans out across mine / share_to_me /
		// shared, so per-ID lookup balloons to 3*N /api/repos/ calls
		// just to render the prompt; ListAll runs the same fan-out
		// exactly once and we look each id up locally. Failure is
		// non-fatal: name resolution is best-effort and we'd rather
		// fall through to "?" than block the rm on a flaky list.
		nameByID := map[string]string{}
		if rows, listErr := client.ListAll(ctx); listErr == nil {
			for _, r := range rows {
				if _, seen := nameByID[r.RepoID]; !seen {
					nameByID[r.RepoID] = r.RepoName
				}
			}
		}
		fmt.Fprintln(out, "The following Sync library/libraries will be PERMANENTLY deleted (all contents):")
		for _, id := range validIDs {
			name := "?"
			if n, ok := nameByID[id]; ok && n != "" {
				name = n
			}
			fmt.Fprintf(out, "  %s  (%s)\n", id, name)
		}
		fmt.Fprint(out, "proceed with repo deletion? [y/N]: ")
		ok, err := readYesNo(in)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(out, "aborted")
			return nil
		}
	}

	deleted := 0
	for _, id := range validIDs {
		if err := client.Delete(ctx, id); err != nil {
			// Reformat once and use the same message in both the
			// printed "failed:" line and the joined return error,
			// so credential-aware CTAs (e.g. "please run profile
			// login") aren't silently dropped from the on-screen
			// output for auth failures.
			rerr := reformatReposHTTPErr(err, rp.OlaresID, "delete repo "+id)
			errs = append(errs, rerr)
			fmt.Fprintf(out, "failed: %s (%v)\n", id, rerr)
			continue
		}
		deleted++
		fmt.Fprintf(out, "deleted repo %s\n", id)
	}
	if deleted > 0 {
		fmt.Fprintf(out, "removed %d repo%s\n", deleted, pluralS(deleted))
	}
	if len(errs) > 0 {
		// errors.Join keeps every per-id failure addressable; same
		// pattern as `files rm` so the parent's exit-code mapping
		// works the same way.
		return fmt.Errorf("repos rm: %d of %d failed: %w",
			len(errs), len(repoIDs), errors.Join(errs...))
	}
	return nil
}
