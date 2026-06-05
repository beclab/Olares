package files

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/internal/files/smbmount"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewNFSCommand returns the `olares-cli files nfs` parent — the NFS
// half of the LarePass web app's "Connect to Server" flow
// (the SMB half is `olares-cli files smb`). Both mount an external
// server into the per-user files-backend's `external/<node>/...`
// namespace, after which every other `files` verb works against
// `external/<node>/<entry>/...` like any other namespace.
//
// NFS differs from SMB in three ways the CLI surfaces:
//
//   - No credentials. NFS exports are mounted by address alone; there
//     is no username / password step.
//   - Address shape. An NFS target is either a bare host/IP
//     (`192.168.1.10`) — which triggers server-side export discovery —
//     or a full `host:/export` path (`192.168.1.10:/data`) that mounts
//     directly. (SMB uses `//host/share`.)
//   - Discovery code. The server returns the export list under the
//     same HTTP/envelope code it uses for a successful mount; the CLI
//     disambiguates by whether it asked for a list (bare host) or a
//     mount (full path).
//
// All three sub-surfaces share the wire endpoints with `files smb`
// (`/api/mount`, `/api/unmount`, `/api/smb_history`) — see
// internal/files/smbmount for the shared client. `--node` resolution
// also mirrors `files smb` / `files cp`: explicit flag wins, else the
// first `/api/nodes/` entry.
func NewNFSCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nfs",
		Short: "mount / unmount external NFS exports and manage the per-node history",
		Long: `Mount external NFS exports into the per-user files-backend's
` + "`external/<node>/...`" + ` namespace, and manage the per-node "Favorite
Servers" history shared with ` + "`files smb`" + `.

This is the NFS counterpart of ` + "`olares-cli files smb`" + ` — both implement
the LarePass app's "Connect to Server" modal. NFS needs no credentials;
a target is either a bare host (export discovery) or a full host:/export
path (direct mount). Once mounted, every other ` + "`files`" + ` verb works
against the resulting ` + "`external/<node>/<entry>/...`" + ` path.

Wire shape (shared with files smb; dispatched by ?external_type=nfs):

    POST   /api/mount/[<node>/]?external_type=nfs
        body: {url}                      → mount host:/export
        body: {url, operate:"list"}      → discover a host's exports
        reply: {code, message, data}
            mount  → code 200; visible at external/<node>/<entry>/
            list   → code 200; data is the array of exports — re-run
                     mount with one of the host:/export paths.
    POST   /api/unmount/external/<node>/<name>/?external_type=nfs
        body: {}
    GET    /api/smb_history/<node>/                 → array of entries
    PUT    /api/smb_history/<node>/                 body: array (upsert)
    DELETE /api/smb_history/<node>/                 body: array of {url}

Sub-commands:

    mount   <host | host:/export>  [--node N]
    unmount <name>                 [--node N]
    history list                   [--node N] [--json] [--all]
    history add <host:/export>     [--node N]
    history rm  <url>...           [--node N]

Examples:

    # Discover a host's exports (bare host → list, then re-run).
    olares-cli files nfs mount 192.168.1.10
    # → server returned 2 export(s): 192.168.1.10:/data, 192.168.1.10:/backups

    # Mount a specific export.
    olares-cli files nfs mount 192.168.1.10:/data

    # Inspect the mounted entries.
    olares-cli files ls external/<node>/

    # Unmount when done.
    olares-cli files nfs unmount nfs-192-168-1-10-data

    # Stash / list / drop a favorite (no credentials for NFS).
    olares-cli files nfs history add 192.168.1.10:/data
    olares-cli files nfs history list
    olares-cli files nfs history rm 192.168.1.10:/data
`,
	}
	cmd.AddCommand(
		newNFSMountCommand(f),
		newNFSUnmountCommand(f),
		newNFSHistoryCommand(f),
	)
	for _, sub := range cmd.Commands() {
		sub.SilenceUsage = true
	}
	return cmd
}

// nfsFullPathRe matches a full NFS target `host:/export` (single
// slash after the colon, NOT a `scheme://` URL). The optional
// trailing group keeps `host:/` (export root) valid while rejecting
// `host://...`. Mirrors LarePass's NFS_URL_REG, rewritten without
// the lookahead RE2 doesn't support.
var nfsFullPathRe = regexp.MustCompile(`^[^\s/]+:/([^/].*)?$`)

// nfsTargetKind classifies a user-supplied NFS target into:
//
//	"full" → host:/export — mount directly.
//	"host" → bare host/IP — ask the server to list exports.
//
// Errors on a scheme URL (`nfs://...`), an SMB-style `//host/share`,
// or anything else that is neither shape. Mirrors LarePass's
// isValidMountUrlForType(url, NFS) (NFS_URL_REG || NFS_HOST_REG).
func nfsTargetKind(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", errors.New("nfs target is empty")
	}
	if strings.Contains(s, "://") {
		return "", fmt.Errorf("nfs target %q looks like a URL scheme; use a bare host (192.168.1.10) or host:/export (192.168.1.10:/data)", s)
	}
	if strings.HasPrefix(s, "//") {
		return "", fmt.Errorf("nfs target %q is an SMB-style path; use `olares-cli files smb` for // shares, or pass an NFS host:/export here", s)
	}
	if nfsFullPathRe.MatchString(s) {
		return "full", nil
	}
	if !strings.Contains(s, "/") {
		return "host", nil
	}
	return "", fmt.Errorf("nfs target %q is malformed; expected a bare host (192.168.1.10) or host:/export (192.168.1.10:/data)", s)
}

// nfsMountOptions captures flag state for `files nfs mount`.
type nfsMountOptions struct {
	node    string
	jsonOut bool
}

func newNFSMountCommand(f *cmdutil.Factory) *cobra.Command {
	o := &nfsMountOptions{}
	cmd := &cobra.Command{
		Use:   "mount <host | host:/export> [--node <node>]",
		Short: "mount an external NFS export into external/<node>/ (or discover a host's exports)",
		Long: `Mount an external NFS export into ` + "`external/<node>/...`" + `.

The target is one of:

    192.168.1.10            host-only → the server lists the host's
                            exports; the CLI prints them and exits
                            non-zero so a script can re-target.
    192.168.1.10:/data      a full export path → mounted directly.

NFS needs no credentials, so there is no username / password step
(unlike ` + "`files smb mount`" + `).

After a successful mount the entry appears under
` + "`external/<node>/<entry>/`" + ` — confirm with ` + "`olares-cli files ls external/<node>/`" + `.

Examples:

    # Discover, then mount.
    olares-cli files nfs mount 192.168.1.10
    olares-cli files nfs mount 192.168.1.10:/data

    # Discovery as JSON (for scripts).
    olares-cli files nfs mount 192.168.1.10 --json
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNFSMount(cmd.Context(), f, cmd.OutOrStdout(), args[0], o)
		},
	}
	cmd.Flags().StringVar(&o.node, "node", "", "target node (defaults to the first /api/nodes/ entry)")
	cmd.Flags().BoolVar(&o.jsonOut, "json", false, "print the discovered export list as JSON instead of a table")
	return cmd
}

func runNFSMount(
	ctx context.Context,
	f *cmdutil.Factory,
	out io.Writer,
	target string,
	o *nfsMountOptions,
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	target = strings.TrimSpace(target)
	kind, err := nfsTargetKind(target)
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
	client := &smbmount.Client{HTTPClient: httpClient, BaseURL: rp.FilesURL}

	node, err := resolveSMBNode(ctx, client, o.node)
	if err != nil {
		return reformatNFSHTTPErr(err, rp.OlaresID, fmt.Sprintf("mount %s", target))
	}

	listReq := kind == "host"
	displayNode := node
	if displayNode == "" {
		displayNode = "(no node)"
	}
	if listReq {
		fmt.Fprintf(out, "discover: %s @ %s\n", target, displayNode)
	} else {
		fmt.Fprintf(out, "mount: %s @ %s\n", target, displayNode)
	}

	res, err := client.MountNFS(ctx, node, smbmount.NFSMountOptions{URL: target, List: listReq})
	if err != nil {
		return reformatNFSHTTPErr(err, rp.OlaresID, fmt.Sprintf("mount %s", target))
	}

	if res.Listed {
		return renderNFSExportList(out, target, res.Exports, o.jsonOut)
	}

	fmt.Fprintf(out, "  ✓ mounted; the export is now visible at external/%s/<entry>/\n", node)
	fmt.Fprintf(out, "    confirm with: olares-cli files ls external/%s/\n", node)
	return nil
}

// renderNFSExportList prints the discovered exports and returns a
// non-zero (error) result so a shell `if !` branch can detect the
// "pick one and re-run" case — same scriptable UX as SMB's code-300
// path. Each export is rendered as a remountable `host:/export`
// string; already-mounted exports are annotated.
func renderNFSExportList(out io.Writer, host string, exports []smbmount.NFSExport, jsonOut bool) error {
	if jsonOut {
		type row struct {
			Path    string `json:"path"`
			Mounted bool   `json:"mounted"`
		}
		rows := make([]row, 0, len(exports))
		for _, e := range exports {
			rows = append(rows, row{Path: nfsRemountURL(host, e.Path), Mounted: e.Mounted})
		}
		b, _ := json.MarshalIndent(map[string]any{"host": host, "exports": rows}, "", "  ")
		fmt.Fprintln(out, string(b))
		return fmt.Errorf("nfs mount returned an export list; re-run with one of the paths above")
	}

	if len(exports) == 0 {
		fmt.Fprintf(out, "server returned no exports for host %s\n", host)
		return fmt.Errorf("nfs mount returned no exports for %s", host)
	}
	fmt.Fprintf(out, "server returned %d export(s) — pick one and re-run mount:\n", len(exports))
	for _, e := range exports {
		suffix := ""
		if e.Mounted {
			suffix = "  (already mounted)"
		}
		fmt.Fprintf(out, "  %s%s\n", nfsRemountURL(host, e.Path), suffix)
	}
	return fmt.Errorf("nfs mount returned an export list; re-run with one of the paths above")
}

// nfsRemountURL turns a server-reported export path into a target
// the user can pass straight back to `nfs mount`. The server may
// return either a full `host:/export` string or a bare `/export`
// dir; in the latter case we splice the host the user typed back
// on. Mirrors LarePass's formatNfsMountPathItem.
func nfsRemountURL(host, path string) string {
	path = strings.TrimSpace(path)
	if strings.Contains(path, ":/") {
		return path
	}
	// host the user typed may itself be a full `host:/x` if they
	// re-listed; strip to the host portion before splicing.
	h := host
	if i := strings.Index(h, ":/"); i >= 0 {
		h = h[:i]
	}
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return h + ":" + path
}

// nfsUnmountOptions captures flag state for `files nfs unmount`.
type nfsUnmountOptions struct {
	node string
}

func newNFSUnmountCommand(f *cmdutil.Factory) *cobra.Command {
	o := &nfsUnmountOptions{}
	cmd := &cobra.Command{
		Use:   "unmount <name> [--node <node>]",
		Short: "unmount a previously-mounted NFS entry from external/<node>/",
		Long: `Unmount an NFS entry from ` + "`external/<node>/`" + `.

` + "`<name>`" + ` is the entry name visible in ` + "`files ls external/<node>/`" + ` —
use that command first to discover the exact name.

Wire shape:

    POST /api/unmount/external/<node>/<name>/?external_type=nfs
    body: {}

Examples:

    olares-cli files ls external/main/
    olares-cli files nfs unmount nfs-192-168-1-10-data --node main
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNFSUnmount(cmd.Context(), f, cmd.OutOrStdout(), args[0], o)
		},
	}
	cmd.Flags().StringVar(&o.node, "node", "", "node hosting the mount (defaults to the first /api/nodes/ entry)")
	return cmd
}

func runNFSUnmount(
	ctx context.Context,
	f *cmdutil.Factory,
	out io.Writer,
	name string,
	o *nfsUnmountOptions,
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("entry name is empty")
	}
	if strings.ContainsAny(name, "/\\") {
		return fmt.Errorf("entry name %q must not contain '/' or '\\\\'; pass the bare entry name as it appears under external/<node>/", name)
	}

	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return err
	}
	httpClient, err := f.HTTPClient(ctx)
	if err != nil {
		return err
	}
	client := &smbmount.Client{HTTPClient: httpClient, BaseURL: rp.FilesURL}

	node, err := resolveSMBNode(ctx, client, o.node)
	if err != nil {
		return reformatNFSHTTPErr(err, rp.OlaresID, fmt.Sprintf("unmount %s", name))
	}
	if node == "" {
		return errors.New("could not resolve a node for unmount; pass --node <name> explicitly")
	}

	display := fmt.Sprintf("external/%s/%s", node, name)
	fmt.Fprintf(out, "unmount: %s\n", display)
	if err := client.Unmount(ctx, "external", node, name, "nfs"); err != nil {
		return reformatNFSHTTPErr(err, rp.OlaresID, fmt.Sprintf("unmount %s", display))
	}
	fmt.Fprintf(out, "  ✓ unmounted %s\n", display)
	return nil
}

// newNFSHistoryCommand groups the per-node NFS favorites verbs. The
// favorites store is the SAME per-node book SMB uses
// (`/api/smb_history/<node>/`) — NFS entries are URL-only (no saved
// credentials). `history list` shows only NFS-shaped entries by
// default (pass --all to include SMB favorites too).
func newNFSHistoryCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history",
		Short: "manage the per-node NFS favorites (shared book with files smb)",
		Long: `Manage the per-node "Favorite Servers" — the same list the LarePass
"Connect to Server" dialog keeps. The underlying store is shared with
` + "`files smb history`" + ` (one per-node book at ` + "`/api/smb_history/<node>/`" + `);
NFS favorites are URL-only (NFS needs no credentials).

` + "`history list`" + ` shows only NFS-shaped entries (host / host:/export) by
default; pass --all to include SMB favorites (// shares) too.

Wire shape:

    GET    /api/smb_history/<node>/                 → array of entries
    PUT    /api/smb_history/<node>/                 body: array (upsert)
    DELETE /api/smb_history/<node>/                 body: array of {url}

Examples:

    olares-cli files nfs history list
    olares-cli files nfs history add 192.168.1.10:/data
    olares-cli files nfs history rm  192.168.1.10:/data
`,
	}
	cmd.AddCommand(
		newNFSHistoryListCommand(f),
		newNFSHistoryAddCommand(f),
		newNFSHistoryRmCommand(f),
	)
	for _, sub := range cmd.Commands() {
		sub.SilenceUsage = true
	}
	return cmd
}

// isNFSFavorite reports whether a favorite URL is NFS-shaped (i.e.
// not an SMB `//host/share`). Used to filter the shared favorites
// book down to NFS entries for `nfs history list`.
func isNFSFavorite(url string) bool {
	return !strings.HasPrefix(strings.TrimSpace(url), "//")
}

type nfsHistoryListOptions struct {
	node    string
	jsonOut bool
	all     bool
}

func newNFSHistoryListCommand(f *cmdutil.Factory) *cobra.Command {
	o := &nfsHistoryListOptions{}
	cmd := &cobra.Command{
		Use:   "list [--node <node>] [--json] [--all]",
		Short: "list the per-node NFS favorites (NFS-shaped entries only unless --all)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNFSHistoryList(cmd.Context(), f, cmd.OutOrStdout(), o)
		},
	}
	cmd.Flags().StringVar(&o.node, "node", "", "node whose history to read (defaults to the first /api/nodes/ entry)")
	cmd.Flags().BoolVar(&o.jsonOut, "json", false, "print each entry as JSON (one per line)")
	cmd.Flags().BoolVar(&o.all, "all", false, "include SMB favorites (// shares), not just NFS-shaped entries")
	return cmd
}

func runNFSHistoryList(
	ctx context.Context,
	f *cmdutil.Factory,
	out io.Writer,
	o *nfsHistoryListOptions,
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return err
	}
	httpClient, err := f.HTTPClient(ctx)
	if err != nil {
		return err
	}
	client := &smbmount.Client{HTTPClient: httpClient, BaseURL: rp.FilesURL}
	node, err := resolveSMBNode(ctx, client, o.node)
	if err != nil {
		return reformatNFSHTTPErr(err, rp.OlaresID, "nfs history list")
	}
	if node == "" {
		return errors.New("could not resolve a node for NFS history; pass --node <name> explicitly")
	}
	entries, err := client.HistoryList(ctx, node)
	if err != nil {
		return reformatNFSHTTPErr(err, rp.OlaresID, "nfs history list")
	}
	filtered := make([]smbmount.HistoryEntry, 0, len(entries))
	for _, e := range entries {
		if o.all || isNFSFavorite(e.URL) {
			filtered = append(filtered, e)
		}
	}
	if o.jsonOut {
		enc := json.NewEncoder(out)
		for _, e := range filtered {
			if err := enc.Encode(e); err != nil {
				return err
			}
		}
		return nil
	}
	if len(filtered) == 0 {
		fmt.Fprintf(out, "(no NFS history entries for node %q)\n", node)
		return nil
	}
	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "URL\tTYPE")
	for _, e := range filtered {
		kind := "nfs"
		if !isNFSFavorite(e.URL) {
			kind = "smb"
		}
		fmt.Fprintf(tw, "%s\t%s\n", e.URL, kind)
	}
	return tw.Flush()
}

type nfsHistoryAddOptions struct {
	node string
}

func newNFSHistoryAddCommand(f *cmdutil.Factory) *cobra.Command {
	o := &nfsHistoryAddOptions{}
	cmd := &cobra.Command{
		Use:   "add <host:/export | host> [--node <node>]",
		Short: "add or update an NFS favorite (per-node, URL-only)",
		Long: `Add or update an NFS entry in the per-node favorites book.

NFS favorites are URL-only — there are no credentials to save. The
target follows the same shape as ` + "`nfs mount`" + ` (a bare host or a full
host:/export path).

Wire shape:

    PUT /api/smb_history/<node>/    body: [{url}]

Examples:

    olares-cli files nfs history add 192.168.1.10:/data
    olares-cli files nfs history add 192.168.1.10
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNFSHistoryAdd(cmd.Context(), f, cmd.OutOrStdout(), args[0], o)
		},
	}
	cmd.Flags().StringVar(&o.node, "node", "", "node whose history to write (defaults to the first /api/nodes/ entry)")
	return cmd
}

func runNFSHistoryAdd(
	ctx context.Context,
	f *cmdutil.Factory,
	out io.Writer,
	target string,
	o *nfsHistoryAddOptions,
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	target = strings.TrimSpace(target)
	// Validate the same way mount does so a typo'd favorite can't be
	// saved (it would never match at mount time anyway).
	if _, err := nfsTargetKind(target); err != nil {
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
	client := &smbmount.Client{HTTPClient: httpClient, BaseURL: rp.FilesURL}

	node, err := resolveSMBNode(ctx, client, o.node)
	if err != nil {
		return reformatNFSHTTPErr(err, rp.OlaresID, "nfs history add")
	}
	if node == "" {
		return errors.New("could not resolve a node for NFS history; pass --node <name> explicitly")
	}

	entry := smbmount.HistoryEntry{URL: target}
	if err := client.HistoryUpsert(ctx, node, []smbmount.HistoryEntry{entry}); err != nil {
		return reformatNFSHTTPErr(err, rp.OlaresID, fmt.Sprintf("nfs history add %s", target))
	}
	fmt.Fprintf(out, "  ✓ saved favorite %s on node %s\n", target, node)
	return nil
}

type nfsHistoryRmOptions struct {
	node string
}

func newNFSHistoryRmCommand(f *cmdutil.Factory) *cobra.Command {
	o := &nfsHistoryRmOptions{}
	cmd := &cobra.Command{
		Use:   "rm <url>... [--node <node>]",
		Short: "remove one or more NFS favorites by URL",
		Long: `Remove one or more entries from the per-node favorites book.

Wire shape:

    DELETE /api/smb_history/<node>/    body: [{url}, ...]

Multiple URLs in a single invocation are batched into one request.

Examples:

    olares-cli files nfs history rm 192.168.1.10:/data
    olares-cli files nfs history rm 192.168.1.10:/data 10.0.0.2:/backups
`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNFSHistoryRm(cmd.Context(), f, cmd.OutOrStdout(), args, o)
		},
	}
	cmd.Flags().StringVar(&o.node, "node", "", "node whose history to write (defaults to the first /api/nodes/ entry)")
	return cmd
}

func runNFSHistoryRm(
	ctx context.Context,
	f *cmdutil.Factory,
	out io.Writer,
	urls []string,
	o *nfsHistoryRmOptions,
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	cleaned := make([]string, 0, len(urls))
	for _, u := range urls {
		u = strings.TrimSpace(u)
		if u == "" {
			continue
		}
		cleaned = append(cleaned, u)
	}
	if len(cleaned) == 0 {
		return errors.New("no NFS urls given")
	}

	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return err
	}
	httpClient, err := f.HTTPClient(ctx)
	if err != nil {
		return err
	}
	client := &smbmount.Client{HTTPClient: httpClient, BaseURL: rp.FilesURL}
	node, err := resolveSMBNode(ctx, client, o.node)
	if err != nil {
		return reformatNFSHTTPErr(err, rp.OlaresID, "nfs history rm")
	}
	if node == "" {
		return errors.New("could not resolve a node for NFS history; pass --node <name> explicitly")
	}
	if err := client.HistoryRemove(ctx, node, cleaned); err != nil {
		return reformatNFSHTTPErr(err, rp.OlaresID, "nfs history rm")
	}
	for _, u := range cleaned {
		fmt.Fprintf(out, "  ✓ removed favorite %s on node %s\n", u, node)
	}
	return nil
}

// reformatNFSHTTPErr delegates to the shared mount-surface error
// reformatter (reformatSMBHTTPErr) — the 401/403/404 mapping and
// typed-credential handling are identical for SMB and NFS since they
// share the wire endpoints. Kept as a thin named wrapper so the NFS
// call sites read clearly.
func reformatNFSHTTPErr(err error, olaresID, op string) error {
	return reformatSMBHTTPErr(err, olaresID, op)
}
