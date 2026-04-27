package apps

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings apps permissions <app>` and
// `olares-cli settings apps providers list <app>`.
//
// Two read-only views of an installed app's permission surface, one per
// page in the SPA's Application detail view:
//
//  1. `permissions` — the permissions VECTOR the app declared in its
//     OlaresManifest (which provider data types it consumes). Backed by
//     GET /api/applications/permissions/<app>, which the SPA wires
//     through `applicationStore.getPermissions(app_name)`.
//
//  2. `providers list` — the provider REGISTRY the app exposes (which
//     data types it offers OTHER apps to consume). Backed by
//     GET /api/applications/provider/registry/<app>, wired through
//     `applicationStore.getProviderRegistryList(app_name)`.
//
// Both endpoints sit on the BFL proxy (no user-service controller) and
// return a standard {code, data} envelope. The provider-registry payload
// wraps the rows in a {items: [...]} object the way the SPA expects.
//
// Role: any authenticated user can read these for their own apps; no
// preflight. Server returns 403 if the active user doesn't own the app.

// AppPermission mirrors apps/.../constant/global.ts's AppPermission:
// the tuple of {dataType, group, version, ops[]} entries the app
// declared at install time. ops[] lists the upstream verbs the app
// promises to use (e.g. ["GetUser", "ListUsers"]).
type AppPermission struct {
	App         string       `json:"app"`
	Owner       string       `json:"owner"`
	Permissions []permission `json:"permissions"`
}

type permission struct {
	DataType string   `json:"dataType"`
	Group    string   `json:"group"`
	Ops      []string `json:"ops"`
	Version  string   `json:"version"`
}

// providerRegister mirrors PermissionProviderRegister from the SPA
// (apps/.../constant/global.ts) — one entry per data type the app
// itself REGISTERS as a provider. opApis is a list of {name, uri}
// pairs that other apps will call against this provider.
type providerRegister struct {
	DataType    string             `json:"dataType"`
	Deployment  string             `json:"deployment"`
	Endpoint    string             `json:"endpoint"`
	Group       string             `json:"group"`
	Kind        string             `json:"kind"`
	Namespace   string             `json:"namespace"`
	Description string             `json:"description"`
	OpAPIs      []providerRegister `json:"opApis"`
	Version     string             `json:"version"`
}

// providerRegistryEnvelope unwraps the {items: [...]} payload that BFL
// returns inside the standard envelope's data field.
type providerRegistryEnvelope struct {
	Items []providerRegister `json:"items"`
}

// NewPermissionsCommand returns `settings apps permissions <app>`.
func NewPermissionsCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "permissions <app>",
		Short: "show the permissions an app declared (Settings -> App -> Permissions)",
		Long: `Show the permission VECTOR an app declared in its OlaresManifest — i.e.
which provider data types the app consumes from other apps. Each entry
is a {dataType, group, version, ops[]} tuple where ops lists the
upstream verbs the app intends to call.

This is the same data the SPA's Application detail page renders under
the "Permissions" panel (apps/.../ApplicationDetailPage.vue, via
applicationStore.getPermissions()).

Note: this is the inverse of "providers list" — permissions are what
the app CONSUMES, providers are what it OFFERS.

Pass --output json for the raw AppPermission record.

Examples:
  olares-cli settings apps permissions files
  olares-cli settings apps permissions files -o json
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runPermissions(c.Context(), f, args[0], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runPermissions(ctx context.Context, f *cmdutil.Factory, app, outputRaw string) error {
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
	path := "/api/applications/permissions/" + url.PathEscape(app)
	var perm AppPermission
	if err := doGetEnvelope(ctx, pc.doer, path, &perm); err != nil {
		return err
	}
	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, perm)
	default:
		return renderPermissions(os.Stdout, perm)
	}
}

func renderPermissions(w io.Writer, p AppPermission) error {
	if _, err := fmt.Fprintf(w, "App:   %s\n", nonEmpty(p.App)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Owner: %s\n", nonEmpty(p.Owner)); err != nil {
		return err
	}
	if len(p.Permissions) == 0 {
		_, err := fmt.Fprintln(w, "\nNo declared permissions.")
		return err
	}
	fmt.Fprintln(w, "")
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "DATA TYPE\tGROUP\tVERSION\tOPS"); err != nil {
		return err
	}
	for _, e := range p.Permissions {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			nonEmpty(e.DataType),
			nonEmpty(e.Group),
			nonEmpty(e.Version),
			joinNonEmptyList(e.Ops),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

// NewProvidersCommand returns the `settings apps providers` parent.
// Phase 3 ships the read; further verbs are a deliberate non-goal
// because the registry is configured at install time, not runtime.
func NewProvidersCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "providers",
		Short: "provider registry the app OFFERS (Settings -> App -> Providers)",
		Long: `Inspect the provider registry an app exposes for other apps to consume.

Subcommands:
  list <app>    list every {dataType, group, version, opApis[]} entry the
                app registered (Settings -> App -> Providers panel)

The provider registry is configured by the app's OlaresManifest at
install time and is not editable at runtime — there are no set / add /
remove verbs upstream.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newProvidersListCommand(f))
	return cmd
}

func newProvidersListCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "list <app>",
		Short: "list providers the app registered for other apps to call",
		Long: `List the provider entries the app declared in its OlaresManifest.

Each row is a data type the app PROVIDES, including its endpoint
(host:port pair the consumer connects to), kind ("provider" /
"sysProvider"), and the list of opApis (the {name, uri} pairs other
apps will call against this provider).

Pass --output json for the full PermissionProviderRegister[] vector
including opApis.

Examples:
  olares-cli settings apps providers list profile
  olares-cli settings apps providers list profile -o json
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runProvidersList(c.Context(), f, args[0], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runProvidersList(ctx context.Context, f *cmdutil.Factory, app, outputRaw string) error {
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
	path := "/api/applications/provider/registry/" + url.PathEscape(app)
	var env providerRegistryEnvelope
	if err := doGetEnvelope(ctx, pc.doer, path, &env); err != nil {
		return err
	}
	switch format {
	case FormatJSON:
		if env.Items == nil {
			env.Items = []providerRegister{}
		}
		return printJSON(os.Stdout, env.Items)
	default:
		return renderProviders(os.Stdout, env.Items)
	}
}

func renderProviders(w io.Writer, items []providerRegister) error {
	if len(items) == 0 {
		_, err := fmt.Fprintln(w, "no provider registrations")
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "DATA TYPE\tGROUP\tVERSION\tKIND\tENDPOINT\tOP APIS"); err != nil {
		return err
	}
	for _, p := range items {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%d\n",
			nonEmpty(p.DataType),
			nonEmpty(p.Group),
			nonEmpty(p.Version),
			nonEmpty(p.Kind),
			nonEmpty(p.Endpoint),
			len(p.OpAPIs),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

// joinNonEmptyList renders a string slice for a one-line table cell. Empty
// or whitespace-only entries are dropped; the result is "-" for fully
// empty input so the column never collapses to nothing.
func joinNonEmptyList(in []string) string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return "-"
	}
	return strings.Join(out, ",")
}
