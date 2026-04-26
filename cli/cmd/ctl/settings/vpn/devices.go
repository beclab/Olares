package vpn

import (
	"context"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings vpn devices ...`
//
// Backed by user-service's HeadScaleController (headscale.controller.ts:13),
// which proxies the per-Olares Headscale instance via:
//
//   GET /headscale/machine                 → list devices
//   GET /headscale/machine/:id/routes      → list routes for one device
//
// Wire shapes (NO BFL envelope — Headscale returns raw JSON, which
// user-service forwards verbatim through ProviderClient.execute):
//
//   /headscale/machine            → { "machines": [HeadScaleDevice...] }
//   /headscale/machine/:id/routes → { "routes":   [Route...] }
//
// HeadScaleDevice fields the SPA reads (HeadScaleDeviceCard.vue):
//   id, name, givenName, ipAddresses[], lastSeen (ISO timestamp string),
//   forcedTags[]
//
// Route fields the SPA reads:
//   id, prefix, enabled
//
// Phase 1 ships read-only verbs:
//
//   vpn devices list                       (Phase 1)
//   vpn devices routes <device-id>         (Phase 1)
//
// Phase 3 adds the writes:
//
//   vpn devices rename <id> <name>         (Phase 3 — devices_writes.go)
//   vpn devices delete <id>                (Phase 3 — devices_writes.go)
//   vpn devices tags set <id> --tag ...    (Phase 3 — devices_writes.go)
//
// Route enable / disable lives under `settings vpn routes` rather than
// here because Headscale routes are addressed by route id, not device
// id (the device-id-keyed `vpn devices routes <id>` verb only *lists*
// the device's routes — toggling them goes through /headscale/routes/
// <route-id>/{enable,disable}).
func NewDevicesCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "devices",
		Short: "Manage Headscale devices on this Olares mesh",
		Long: `Inspect and manage Headscale devices that have joined this Olares user's
mesh.

Subcommands:
  list                                   list all devices                          (Phase 1)
  routes <device-id>                     list routes advertised by one device      (Phase 1)
  rename <device-id> <new-name>          rename a device                           (Phase 3)
  delete <device-id>                     remove a device                           (Phase 3)
  tags set <device-id> --tag <name>...   replace the device's forcedTags           (Phase 3)
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newDevicesListCommand(f))
	cmd.AddCommand(newDevicesRoutesCommand(f))
	addDevicesWriteCommands(cmd, f)
	return cmd
}

func newDevicesListCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list Headscale devices on this Olares user's mesh",
		Long: `List Headscale devices joined to this Olares user.

Each row shows the device id, given name (the user-friendly label
Headscale uses on its UI), the canonical hostname, the most recent
lastSeen timestamp, and any forcedTags.

Pass --output json for the full HeadScaleDevice record (preauthKey,
nodeKey, machineKey, ephemeral flag, advertised routes, etc.).
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runDevicesList(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func newDevicesRoutesCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "routes <device-id>",
		Short: "list routes advertised by one device",
		Long: `List the routes (subnet prefixes) one Headscale device is advertising,
plus whether each route is currently enabled by the controller.

The device-id is the numeric / string id from "settings vpn devices list".
Pass --output json for the raw Route records (advertised flag, isPrimary,
created/updated timestamps).
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runDevicesRoutes(c.Context(), f, args[0], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

// headScaleDevice mirrors the field set the SPA renders. Headscale's
// upstream struct has more fields (preAuthKey, machineKey, ...); we
// surface them through --output json by keeping all decoded fields on
// the JSON path even though the table only shows a subset.
//
// We use json.RawMessage for the catch-all of unknown fields so that
// pure pass-through to JSON output preserves them. But Go's json
// package doesn't let us "merge" RawMessage on top of strongly-typed
// fields without a custom UnmarshalJSON; for simplicity we keep this
// as a typed struct only — JSON output will just show the typed fields.
// If users need richer JSON we can switch to map[string]any later.
type headScaleDevice struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	GivenName   string   `json:"givenName"`
	IPAddresses []string `json:"ipAddresses"`
	LastSeen    string   `json:"lastSeen"`
	Online      bool     `json:"online"`
	ForcedTags  []string `json:"forcedTags"`
	User        struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"user"`
}

type devicesListResp struct {
	Machines []headScaleDevice `json:"machines"`
}

func runDevicesList(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
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

	var resp devicesListResp
	if err := pc.doer.DoJSON(ctx, "GET", "/headscale/machine", nil, &resp); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, resp.Machines)
	default:
		return renderDevicesTable(os.Stdout, resp.Machines)
	}
}

func renderDevicesTable(w io.Writer, devices []headScaleDevice) error {
	if len(devices) == 0 {
		_, err := fmt.Fprintln(w, "no devices")
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "ID\tGIVEN NAME\tHOSTNAME\tONLINE\tIPS\tTAGS\tLAST SEEN"); err != nil {
		return err
	}
	for _, d := range devices {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			nonEmpty(d.ID),
			nonEmpty(d.GivenName),
			nonEmpty(d.Name),
			boolStr(d.Online),
			joinNonEmpty(d.IPAddresses, ","),
			joinNonEmpty(stripTagPrefix(d.ForcedTags), ","),
			fmtIsoTime(d.LastSeen),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

// stripTagPrefix mirrors the SPA's tag display: Headscale stores tags as
// "tag:foo" but the UI hides the "tag:" prefix. We do the same in the
// table so the CLI output reads like the dashboard.
func stripTagPrefix(tags []string) []string {
	if len(tags) == 0 {
		return tags
	}
	out := make([]string, len(tags))
	for i, t := range tags {
		out[i] = stripPrefix(t, "tag:")
	}
	return out
}

func stripPrefix(s, prefix string) string {
	if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
		return s[len(prefix):]
	}
	return s
}

// fmtIsoTime parses an ISO-8601 / RFC3339 timestamp string from
// Headscale and re-renders it in the user's local timezone. Empty or
// unparseable values render as "-" so the table doesn't break.
func fmtIsoTime(s string) string {
	if s == "" {
		return "-"
	}
	for _, layout := range []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.999999999Z",
		"2006-01-02T15:04:05.999999Z",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.Local().Format(time.RFC3339)
		}
	}
	return s
}

// route mirrors the Headscale Route shape the SPA reads in
// HeadScaleDeviceCard.vue (id, prefix, enabled). Additional Headscale
// fields (advertised, isPrimary, created/updated) come through under
// --output json via the typed fields below.
type route struct {
	ID         int    `json:"id"`
	Prefix     string `json:"prefix"`
	Advertised bool   `json:"advertised"`
	Enabled    bool   `json:"enabled"`
	IsPrimary  bool   `json:"isPrimary"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

type devicesRoutesResp struct {
	Routes []route `json:"routes"`
}

func runDevicesRoutes(ctx context.Context, f *cmdutil.Factory, deviceID, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if deviceID == "" {
		return fmt.Errorf("device-id is required")
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}

	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}

	var resp devicesRoutesResp
	path := "/headscale/machine/" + deviceID + "/routes"
	if err := pc.doer.DoJSON(ctx, "GET", path, nil, &resp); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, resp.Routes)
	default:
		return renderRoutesTable(os.Stdout, resp.Routes)
	}
}

func renderRoutesTable(w io.Writer, routes []route) error {
	if len(routes) == 0 {
		_, err := fmt.Fprintln(w, "no routes advertised")
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "ID\tPREFIX\tADVERTISED\tENABLED\tPRIMARY"); err != nil {
		return err
	}
	for _, r := range routes {
		if _, err := fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\n",
			r.ID,
			nonEmpty(r.Prefix),
			boolStr(r.Advertised),
			boolStr(r.Enabled),
			boolStr(r.IsPrimary),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}
