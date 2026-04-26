package apps

import (
	"context"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings apps get <name>`
//
// There is no per-app "get" endpoint in user-service or BFL — the SPA's
// Application detail page reads a single record from the in-memory list
// it loaded from /api/myapps (see ApplicationDetailPage.vue's reliance on
// `applicationStore.applications.find(...)`). We follow the same model:
// fetch the full list, filter by name, and render a single record.
//
// The trade-off: each `apps get` call pays the cost of a list fetch.
// That's fine for an interactive CLI (the full list is small — typically
// dozens of apps, not thousands), and it sidesteps having to keep two
// separate decoders for "list" and "single" payload shapes.
//
// Role: any authenticated user; no preflight.
func NewGetCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "get <name>",
		Short: "show one app's settings record (Settings -> Apps -> <name>)",
		Long: `Show the AppInfo record for a single installed app.

The name is the lowercase app id (e.g. "files", "vault", "studio") — the
same name that appears in the NAME column of "settings apps list".

Pass --output json for the full record including entrances, sharedEntrances,
ports, owner, namespace, and the various flags (isClusterScoped,
isSysApp, mobileSupported).
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runGet(c.Context(), f, args[0], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runGet(ctx context.Context, f *cmdutil.Factory, name, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if name == "" {
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

	// Same payload as `apps list` — the per-app endpoint doesn't exist
	// upstream. We fetch the unfiltered list (--all + system) so the
	// user can `apps get` system apps that are otherwise hidden by the
	// list view's defaults.
	var rows []appInfo
	if err := doGetEnvelope(ctx, pc.doer, "/api/myapps", &rows); err != nil {
		return err
	}

	var match *appInfo
	for i := range rows {
		if rows[i].Name == name {
			match = &rows[i]
			break
		}
	}
	if match == nil {
		return fmt.Errorf("app %q not found (try `olares-cli settings apps list --all` to see the full set)", name)
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, match)
	default:
		return renderAppDetail(os.Stdout, *match)
	}
}

func renderAppDetail(w io.Writer, a appInfo) error {
	rows := [][2]string{
		{"Name", nonEmpty(a.Name)},
		{"Title", nonEmpty(a.Title)},
		{"State", nonEmpty(a.State)},
		{"Owner", nonEmpty(a.Owner)},
		{"Namespace", nonEmpty(a.Namespace)},
		{"Deployment", nonEmpty(a.Deployment)},
		{"URL", nonEmpty(a.URL)},
		{"System App", boolStr(a.IsSysApp)},
		{"Cluster-Scoped", boolStr(a.IsClusterScoped)},
		{"Mobile Supported", boolStr(a.MobileSupported)},
		{"Required GPU", nonEmpty(a.RequiredGpu)},
		{"Ports", fmtPorts(a.Ports)},
	}
	for _, r := range rows {
		if _, err := fmt.Fprintf(w, "%-19s %s\n", r[0]+":", r[1]); err != nil {
			return err
		}
	}
	if len(a.Entrances) > 0 {
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "Entrances:")
		if err := renderEntrancesTable(w, a.Entrances); err != nil {
			return err
		}
	}
	if len(a.SharedEntrances) > 0 {
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "Shared Entrances:")
		if err := renderEntrancesTable(w, a.SharedEntrances); err != nil {
			return err
		}
	}
	return nil
}

func renderEntrancesTable(w io.Writer, entries []appEntrance) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "  NAME\tTITLE\tSTATE\tAUTH LEVEL\tINVISIBLE\tURL"); err != nil {
		return err
	}
	for _, e := range entries {
		if _, err := fmt.Fprintf(tw, "  %s\t%s\t%s\t%s\t%s\t%s\n",
			nonEmpty(e.Name),
			nonEmpty(e.Title),
			nonEmpty(e.State),
			nonEmpty(e.AuthLevel),
			boolStr(e.Invisible),
			nonEmpty(e.URL),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func fmtPorts(ports []int) string {
	if len(ports) == 0 {
		return "-"
	}
	out := ""
	for i, p := range ports {
		if i > 0 {
			out += ","
		}
		out += fmt.Sprintf("%d", p)
	}
	return out
}
