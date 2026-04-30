package vpn

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/settings/internal/preflight"
	"github.com/beclab/Olares/cli/pkg/cliutil"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// `olares-cli settings vpn acl <app> ...`
//
// Per-app ACL editor — controls which dst ports an app exposes on the
// Headscale mesh, per protocol. Mirrors the Settings -> VPN -> "Application
// ACL" / Settings -> Application -> ACL page in the SPA
// (apps/.../stores/settings/acl.ts:95-188).
//
// Wire shape:
//
//	GET  /api/acl/app/status?name=<app>
//	    BFL envelope on the wire (the SPA's response interceptor
//	    explicitly does NOT unwrap this one — it short-circuits on
//	    `acl/app/status?name=` so the caller can inspect `data.code`).
//	    code == 0 ⇒ data.data is the AclInfo[] (proto + dst[]).
//	    code != 0 ⇒ "no ACL configured for this app" — NOT an error.
//	                The CLI surfaces this as an empty list, same as the
//	                SPA does (`appAclList = []`).
//
//	POST /api/acl/app/status
//	    body { name, acls: [{proto, dst[]}, ...] }
//	    Replaces the WHOLE per-app ACL vector. There is no add / remove
//	    endpoint upstream — the SPA's add / rm buttons funnel through
//	    this same POST after read-modify-write merge. We mirror that
//	    locally for `acl add` / `acl remove`.
//
// Request shape note: the SPA filters out entries with empty dst[]
// before posting (`acls.filter((e) => e.dst.length > 0)`). The CLI does
// the same so an explicit `--tcp ""` or a fully-cleared protocol
// disappears from the wire body, matching what the upstream sees from
// the SPA today.
//
// Role: the SPA renders these controls on per-app pages without an
// admin/owner gate; we leave preflight to the server. A normal user
// who isn't allowed to change another user's ACL gets a 403 and the
// usual "this command needs role X to ..." CTA.

// AclInfo is the per-protocol entry the upstream stores: a proto label
// (typically "tcp" or "udp") and the list of allowed destinations
// (typically port strings — the upstream is intentionally loose about
// the format because Headscale forwards them verbatim).
type AclInfo struct {
	Proto string   `json:"proto"`
	Dst   []string `json:"dst"`
}

// NewACLCommand returns the `vpn acl` parent. It accepts read-modify-write
// verbs (`add`, `remove`, `clear`) plus the literal `set` and `get`
// shapes; together they cover every action the SPA's per-app ACL page
// can take.
func NewACLCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "acl",
		Short: "per-app ACL editor (Settings -> VPN -> Application ACL)",
		Long: `Manage per-app ACL entries — the per-protocol allow-lists of
destinations (typically port strings) Headscale uses to gate mesh
traffic to an app's pods.

Subcommands:
  all                                          dump every ACL the active user owns
  get    <app>                                 show the per-app ACL vector
  set    <app> [--tcp PORT...] [--udp PORT...] replace the whole vector
  add    <app> [--tcp PORT...] [--udp PORT...] merge new dsts in (read-modify-write)
  remove <app> [--tcp PORT...] [--udp PORT...] drop dsts (read-modify-write)
  clear  <app>                                 remove every ACL entry

If you don't know the app name yet, run "vpn acl all" first to see
every ACL configured for the active user.

--tcp / --udp accept either repeated flags (--tcp 80 --tcp 443) or a
single comma-separated value (--tcp 80,443) — both forms work and the
CLI dedupes either way. Port strings are passed to the upstream
verbatim so any value Headscale accepts (single port, range like
"8000-8100", "*", etc.) round-trips unchanged.

The upstream replaces the whole ACL vector on every POST; the
add/remove convenience verbs read first and merge before posting so
unrelated entries survive untouched.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newACLAllCommand(f))
	cmd.AddCommand(newACLGetCommand(f))
	cmd.AddCommand(newACLSetCommand(f))
	cmd.AddCommand(newACLAddCommand(f))
	cmd.AddCommand(newACLRemoveCommand(f))
	cmd.AddCommand(newACLClearCommand(f))
	return cmd
}

// newACLAllCommand registers `vpn acl all`.
//
// Wraps GET /api/acl/all (user-service/src/bfl/acl.controller.ts:157
// getAclsAll). The SPA does not currently surface this endpoint via a
// page action, but the controller exists upstream; the verb gives the
// caller a single command that lists every per-app ACL configured
// under their account, complementing the per-app `vpn acl get <app>`.
//
// Wire shape: BFL envelope wrapping a map keyed by app name (the
// upstream's headscale acls vector). We stay structurally agnostic by
// surfacing the raw envelope's data field as JSON when --output json,
// and unmarshaling into a map[string][]AclInfo for the table view.
func newACLAllCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "all",
		Short: "list every per-app ACL configured for the active user",
		Long: `List the per-app ACL vectors for every app the active user owns.

Wraps GET /api/acl/all (user-service AclController.getAclsAll). The
upstream returns a map keyed by app name; the default table flattens
it to one row per (app, proto). Pass --output json for the raw map.

If the upstream returns an envelope error code (e.g. "not found")
this verb surfaces an empty list, the same way "acl get <app>" does
for individual apps.

Examples:
  olares-cli settings vpn acl all
  olares-cli settings vpn acl all -o json
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "list per-app ACLs"); err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runACLAll(ctx, f, output), "list per-app ACLs")
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runACLAll(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	all, err := getAllACLViaDoer(ctx, pc.doer)
	if err != nil {
		return err
	}
	switch format {
	case FormatJSON:
		if all == nil {
			all = map[string][]AclInfo{}
		}
		return printJSON(os.Stdout, all)
	default:
		return renderACLAll(os.Stdout, all)
	}
}

// getAllACLViaDoer wraps the GET /api/acl/all envelope. Same envelope
// quirk as `getAppACLViaDoer`: a non-zero `code` is treated as "no ACL
// configured" rather than a hard error, matching the SPA's defensive
// pattern of falling back to an empty list.
//
// We try two payload shapes the upstream has emitted historically:
//
//  1. map[string][]AclInfo  — keyed by app name, the natural shape.
//  2. []struct{name, acls}  — older flat list; we coerce to the map
//                             so the renderer has one code path.
func getAllACLViaDoer(ctx context.Context, d Doer) (map[string][]AclInfo, error) {
	var env bflEnvelope
	if err := d.DoJSON(ctx, "GET", "/api/acl/all", nil, &env); err != nil {
		return nil, err
	}
	if env.Code != 0 && env.Code != 200 {
		return map[string][]AclInfo{}, nil
	}
	if len(env.Data) == 0 || string(env.Data) == "null" {
		return map[string][]AclInfo{}, nil
	}
	// Try map shape first.
	asMap := map[string][]AclInfo{}
	if err := json.Unmarshal(env.Data, &asMap); err == nil {
		return asMap, nil
	}
	// Fallback: array of {name, acls}.
	var asArr []struct {
		Name string    `json:"name"`
		Acls []AclInfo `json:"acls"`
	}
	if err := json.Unmarshal(env.Data, &asArr); err != nil {
		return nil, fmt.Errorf("decode acl/all data: %w", err)
	}
	out := make(map[string][]AclInfo, len(asArr))
	for _, e := range asArr {
		if strings.TrimSpace(e.Name) == "" {
			continue
		}
		out[e.Name] = e.Acls
	}
	return out, nil
}

func renderACLAll(w io.Writer, all map[string][]AclInfo) error {
	if len(all) == 0 {
		_, err := fmt.Fprintln(w, "no ACL configured for any app")
		return err
	}
	apps := make([]string, 0, len(all))
	for app := range all {
		apps = append(apps, app)
	}
	sort.Strings(apps)
	if _, err := fmt.Fprintf(w, "%-24s  %-8s  %s\n", "APP", "PROTO", "DST"); err != nil {
		return err
	}
	for _, app := range apps {
		acls := all[app]
		if len(acls) == 0 {
			if _, err := fmt.Fprintf(w, "%-24s  %-8s  %s\n", app, "-", "(none)"); err != nil {
				return err
			}
			continue
		}
		for _, a := range acls {
			if _, err := fmt.Fprintf(w, "%-24s  %-8s  %s\n", app, nonEmpty(a.Proto), joinNonEmpty(a.Dst, ",")); err != nil {
				return err
			}
		}
	}
	return nil
}

func newACLGetCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "get <app>",
		Short: "show the per-app ACL vector",
		Long: `Show the ACL entries currently configured for an app. Output is one
row per protocol with a comma-joined dst list. --output json returns the
raw [{proto, dst:[...]}] vector.

If no ACL is configured for the app, the upstream returns code != 0
("not found"); the CLI surfaces this as an empty list, matching the
SPA's "appAclList = []" default — it is NOT a hard error.

Example:
  olares-cli settings vpn acl get my-app
  olares-cli settings vpn acl get my-app -o json
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "get per-app ACL"); err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runACLGet(ctx, f, args[0], output), "get per-app ACL")
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runACLGet(ctx context.Context, f *cmdutil.Factory, app, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	app = strings.TrimSpace(app)
	if app == "" {
		return fmt.Errorf("app name is required")
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	acls, err := getAppACLViaDoer(ctx, pc.doer, app)
	if err != nil {
		return err
	}
	switch format {
	case FormatJSON:
		if acls == nil {
			acls = []AclInfo{}
		}
		return printJSON(os.Stdout, acls)
	default:
		return renderACL(os.Stdout, app, acls)
	}
}

// getAppACLViaDoer reads the per-app ACL vector. A non-zero `code` from
// upstream is treated as "no ACL configured" and returns ([]AclInfo,
// nil), matching the SPA's behavior. Genuine transport / 4xx / 5xx
// errors still bubble up because DoJSON formats those before we get the
// envelope back.
func getAppACLViaDoer(ctx context.Context, d Doer, app string) ([]AclInfo, error) {
	path := "/api/acl/app/status?name=" + url.QueryEscape(app)
	var env bflEnvelope
	if err := d.DoJSON(ctx, "GET", path, nil, &env); err != nil {
		return nil, err
	}
	if env.Code != 0 && env.Code != 200 {
		// "not found" — same as SPA's else-branch.
		return []AclInfo{}, nil
	}
	if len(env.Data) == 0 {
		return []AclInfo{}, nil
	}
	var out []AclInfo
	if err := json.Unmarshal(env.Data, &out); err != nil {
		return nil, fmt.Errorf("decode acl data: %w", err)
	}
	return out, nil
}

func renderACL(w io.Writer, app string, acls []AclInfo) error {
	if len(acls) == 0 {
		_, err := fmt.Fprintf(w, "no ACL configured for app %q\n", app)
		return err
	}
	if _, err := fmt.Fprintf(w, "%-8s  %s\n", "PROTO", "DST"); err != nil {
		return err
	}
	for _, a := range acls {
		if _, err := fmt.Fprintf(w, "%-8s  %s\n", nonEmpty(a.Proto), joinNonEmpty(a.Dst, ",")); err != nil {
			return err
		}
	}
	return nil
}

func newACLSetCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		tcp []string
		udp []string
	)
	cmd := &cobra.Command{
		Use:   "set <app>",
		Short: "replace the per-app ACL vector",
		Long: `Replace the entire per-app ACL vector with the supplied protocol
allow-lists. Pass --tcp / --udp (repeatable, comma-separated also OK)
for each protocol. Omitting a protocol drops it from the upstream
config.

Example:
  olares-cli settings vpn acl set my-app --tcp 80,443 --udp 53
  olares-cli settings vpn acl set my-app --tcp 8080
  olares-cli settings vpn acl set my-app                          # equivalent to clear
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "set per-app ACL"); err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runACLSet(ctx, f, args[0], tcp, udp), "set per-app ACL")
		},
	}
	cmd.Flags().StringSliceVar(&tcp, "tcp", nil, "TCP destinations (repeat or comma-separate)")
	cmd.Flags().StringSliceVar(&udp, "udp", nil, "UDP destinations (repeat or comma-separate)")
	return cmd
}

func runACLSet(ctx context.Context, f *cmdutil.Factory, app string, tcp, udp []string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	app = strings.TrimSpace(app)
	if app == "" {
		return fmt.Errorf("app name is required")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	acls := buildACLPayload(tcp, udp)
	if err := postAppACLViaDoer(ctx, pc.doer, app, acls); err != nil {
		return err
	}
	if len(acls) == 0 {
		fmt.Fprintf(os.Stdout, "cleared per-app ACL for %s\n", app)
	} else {
		fmt.Fprintf(os.Stdout, "set per-app ACL for %s: %s\n", app, summarizeACL(acls))
	}
	return nil
}

func newACLAddCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		tcp []string
		udp []string
	)
	cmd := &cobra.Command{
		Use:   "add <app>",
		Short: "merge new dsts into the per-app ACL (read-modify-write)",
		Long: `Add destinations to the per-app ACL. Existing entries on protocols you
don't pass survive untouched; entries on protocols you do pass are
unioned with the existing dst list (de-duped).

Example:
  olares-cli settings vpn acl add my-app --tcp 8080
  olares-cli settings vpn acl add my-app --tcp 80,443 --udp 53
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "add per-app ACL entries"); err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runACLAdd(ctx, f, args[0], tcp, udp), "add per-app ACL entries")
		},
	}
	cmd.Flags().StringSliceVar(&tcp, "tcp", nil, "TCP destinations to add (repeat or comma-separate)")
	cmd.Flags().StringSliceVar(&udp, "udp", nil, "UDP destinations to add (repeat or comma-separate)")
	return cmd
}

func runACLAdd(ctx context.Context, f *cmdutil.Factory, app string, tcp, udp []string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	app = strings.TrimSpace(app)
	if app == "" {
		return fmt.Errorf("app name is required")
	}
	additions := buildACLPayload(tcp, udp)
	if len(additions) == 0 {
		return fmt.Errorf("nothing to add — pass --tcp and/or --udp")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	current, err := getAppACLViaDoer(ctx, pc.doer, app)
	if err != nil {
		return err
	}
	merged := mergeACL(current, additions)
	if err := postAppACLViaDoer(ctx, pc.doer, app, merged); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "updated per-app ACL for %s: %s\n", app, summarizeACL(merged))
	return nil
}

func newACLRemoveCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		tcp []string
		udp []string
	)
	cmd := &cobra.Command{
		Use:     "remove <app>",
		Aliases: []string{"rm"},
		Short:   "drop dsts from the per-app ACL (read-modify-write)",
		Long: `Remove destinations from the per-app ACL. A protocol whose dst list
empties out is dropped entirely from the upstream config (matching the
SPA's behavior of filtering acls with an empty dst[]).

Example:
  olares-cli settings vpn acl remove my-app --tcp 80
  olares-cli settings vpn acl rm my-app --udp 53
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "remove per-app ACL entries"); err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runACLRemove(ctx, f, args[0], tcp, udp), "remove per-app ACL entries")
		},
	}
	cmd.Flags().StringSliceVar(&tcp, "tcp", nil, "TCP destinations to remove (repeat or comma-separate)")
	cmd.Flags().StringSliceVar(&udp, "udp", nil, "UDP destinations to remove (repeat or comma-separate)")
	return cmd
}

func runACLRemove(ctx context.Context, f *cmdutil.Factory, app string, tcp, udp []string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	app = strings.TrimSpace(app)
	if app == "" {
		return fmt.Errorf("app name is required")
	}
	removals := buildACLPayload(tcp, udp)
	if len(removals) == 0 {
		return fmt.Errorf("nothing to remove — pass --tcp and/or --udp")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	current, err := getAppACLViaDoer(ctx, pc.doer, app)
	if err != nil {
		return err
	}
	pruned := subtractACL(current, removals)
	if err := postAppACLViaDoer(ctx, pc.doer, app, pruned); err != nil {
		return err
	}
	if len(pruned) == 0 {
		fmt.Fprintf(os.Stdout, "removed last per-app ACL entries for %s (now empty)\n", app)
	} else {
		fmt.Fprintf(os.Stdout, "updated per-app ACL for %s: %s\n", app, summarizeACL(pruned))
	}
	return nil
}

func newACLClearCommand(f *cmdutil.Factory) *cobra.Command {
	var assumeYes bool
	cmd := &cobra.Command{
		Use:   "clear <app>",
		Short: "remove every per-app ACL entry",
		Long: `Remove every ACL entry for the app. Equivalent to "set" with no flags.
The app loses every previously-allowed dst on every protocol; existing
mesh sessions stay open until renegotiation, but new connections are
denied until you re-add ACL entries.

Prompts for confirmation by default; pass --yes for automation.

Example:
  olares-cli settings vpn acl clear my-app
  olares-cli settings vpn acl clear my-app --yes
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "clear per-app ACL"); err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runACLClear(ctx, f, args[0], assumeYes), "clear per-app ACL")
		},
	}
	cmd.Flags().BoolVar(&assumeYes, "yes", false, "skip the confirmation prompt (required for non-TTY automation)")
	return cmd
}

func runACLClear(ctx context.Context, f *cmdutil.Factory, app string, assumeYes bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	app = strings.TrimSpace(app)
	if app == "" {
		return fmt.Errorf("app name is required")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	if err := cliutil.ConfirmDestructive(os.Stderr, os.Stdin,
		fmt.Sprintf("Remove every per-app ACL entry for %q? Existing mesh sessions stay open but new connections will be denied until ACLs are re-added.", app),
		assumeYes); err != nil {
		return err
	}
	if err := postAppACLViaDoer(ctx, pc.doer, app, nil); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "cleared per-app ACL for %s\n", app)
	return nil
}

// postAppACLViaDoer is the wire-only core of every per-app ACL write.
// SPA filters out empty-dst entries before posting; we do the same so
// the upstream sees the same shape it does from the SPA today.
func postAppACLViaDoer(ctx context.Context, d Doer, app string, acls []AclInfo) error {
	body := struct {
		Name string    `json:"name"`
		Acls []AclInfo `json:"acls"`
	}{
		Name: app,
		Acls: pruneEmptyACL(acls),
	}
	return doMutateEnvelope(ctx, d, "POST", "/api/acl/app/status", body, nil)
}

// buildACLPayload normalizes user-supplied --tcp / --udp slices into the
// upstream wire shape: one AclInfo per protocol that actually has at
// least one non-empty dst, in a stable proto order ("tcp" before "udp"
// before others) so the success message + JSON output round-trip
// deterministically.
func buildACLPayload(tcp, udp []string) []AclInfo {
	out := make([]AclInfo, 0, 2)
	if dsts := dedupeNonEmpty(tcp); len(dsts) > 0 {
		out = append(out, AclInfo{Proto: "tcp", Dst: dsts})
	}
	if dsts := dedupeNonEmpty(udp); len(dsts) > 0 {
		out = append(out, AclInfo{Proto: "udp", Dst: dsts})
	}
	return out
}

// mergeACL unions `additions` into `current`, returning a new slice.
// Protocols present only in current survive untouched; protocols
// present in both have their dst lists unioned (de-duped, first-seen
// order). Protocols present only in additions are appended at the end
// in the order they came in.
func mergeACL(current, additions []AclInfo) []AclInfo {
	if len(additions) == 0 {
		return cloneACL(current)
	}
	byProto := indexACL(current)
	order := protoOrder(current)
	for _, add := range additions {
		key := strings.ToLower(strings.TrimSpace(add.Proto))
		if existing, ok := byProto[key]; ok {
			existing.Dst = unionPreservingOrder(existing.Dst, add.Dst)
			byProto[key] = existing
			continue
		}
		byProto[key] = AclInfo{Proto: add.Proto, Dst: append([]string(nil), add.Dst...)}
		order = append(order, key)
	}
	out := make([]AclInfo, 0, len(order))
	for _, k := range order {
		out = append(out, byProto[k])
	}
	return out
}

// subtractACL drops entries in `removals` from `current`. A protocol
// whose dst list empties out is removed entirely from the result so the
// upstream wire shape matches what the SPA would post (it filters out
// empty-dst entries client-side before POSTing).
func subtractACL(current, removals []AclInfo) []AclInfo {
	if len(removals) == 0 {
		return cloneACL(current)
	}
	dropByProto := map[string]map[string]struct{}{}
	for _, r := range removals {
		key := strings.ToLower(strings.TrimSpace(r.Proto))
		set, ok := dropByProto[key]
		if !ok {
			set = map[string]struct{}{}
			dropByProto[key] = set
		}
		for _, d := range r.Dst {
			d = strings.TrimSpace(d)
			if d != "" {
				set[d] = struct{}{}
			}
		}
	}
	out := make([]AclInfo, 0, len(current))
	for _, entry := range current {
		key := strings.ToLower(strings.TrimSpace(entry.Proto))
		drops, ok := dropByProto[key]
		if !ok {
			out = append(out, AclInfo{Proto: entry.Proto, Dst: append([]string(nil), entry.Dst...)})
			continue
		}
		kept := entry.Dst[:0:0]
		for _, d := range entry.Dst {
			if _, drop := drops[strings.TrimSpace(d)]; drop {
				continue
			}
			kept = append(kept, d)
		}
		if len(kept) > 0 {
			out = append(out, AclInfo{Proto: entry.Proto, Dst: kept})
		}
	}
	return out
}

func pruneEmptyACL(in []AclInfo) []AclInfo {
	if len(in) == 0 {
		return []AclInfo{}
	}
	out := in[:0:0]
	for _, a := range in {
		dsts := dedupeNonEmpty(a.Dst)
		if len(dsts) == 0 {
			continue
		}
		out = append(out, AclInfo{Proto: a.Proto, Dst: dsts})
	}
	return out
}

func cloneACL(in []AclInfo) []AclInfo {
	out := make([]AclInfo, 0, len(in))
	for _, a := range in {
		out = append(out, AclInfo{Proto: a.Proto, Dst: append([]string(nil), a.Dst...)})
	}
	return out
}

func indexACL(in []AclInfo) map[string]AclInfo {
	out := map[string]AclInfo{}
	for _, a := range in {
		key := strings.ToLower(strings.TrimSpace(a.Proto))
		out[key] = AclInfo{Proto: a.Proto, Dst: append([]string(nil), a.Dst...)}
	}
	return out
}

func protoOrder(in []AclInfo) []string {
	out := make([]string, 0, len(in))
	seen := map[string]struct{}{}
	for _, a := range in {
		key := strings.ToLower(strings.TrimSpace(a.Proto))
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	return out
}

func dedupeNonEmpty(in []string) []string {
	out := make([]string, 0, len(in))
	seen := map[string]struct{}{}
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, dup := seen[s]; dup {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func unionPreservingOrder(a, b []string) []string {
	out := make([]string, 0, len(a)+len(b))
	seen := map[string]struct{}{}
	for _, s := range a {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, dup := seen[s]; dup {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	for _, s := range b {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, dup := seen[s]; dup {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

// summarizeACL renders an AclInfo[] as "tcp=80,443 udp=53" for the
// success message. Order matches the slice as produced by the merge /
// subtract helpers, except we sort dst lists to keep the message
// stable across runs that exercised the same end-state.
func summarizeACL(in []AclInfo) string {
	if len(in) == 0 {
		return "(empty)"
	}
	parts := make([]string, 0, len(in))
	for _, a := range in {
		dsts := append([]string(nil), a.Dst...)
		sort.Strings(dsts)
		parts = append(parts, fmt.Sprintf("%s=%s", strings.ToLower(a.Proto), strings.Join(dsts, ",")))
	}
	return strings.Join(parts, " ")
}
