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

// `olares-cli settings apps entrances list <app>` (parent + verb).
//
// Reads the FRESH per-entrance vector for an app from BFL via
// GET /api/applications/<app>/entrances, mirroring the SPA's
// applicationStore.getEntrances(app_name). Yes, /api/myapps already
// includes an entrances[] sub-field — but it's a snapshot from the most
// recent app-service list call, not always in sync with the live state.
// The SPA navigates to a per-app page and re-fetches via this endpoint
// for that reason; the CLI mirrors the behavior so `entrances list`
// reports the same numbers the SPA would.
//
// The response payload is wrapped as {items: TerminusEntrance[]} inside
// the standard BFL envelope. We unwrap data → items and surface them as
// rows. The set of fields per entrance overlaps with what /api/myapps
// returns; we render the same NAME/TITLE/STATE/AUTH LEVEL/INVISIBLE/URL
// columns as `apps get`'s entrance section so the two views are
// consistent.
//
// Setup verbs (`apps domain`, `apps policy`, `apps auth-level`) live in
// their own files but key off the entrance name surfaced here — the
// SPA flow is to land on this page first to learn the entrance names,
// then click through to per-entrance setup. The CLI matches.
//
// Role: any authenticated user can read their own apps' entrances; no
// preflight. Returns 403 if the active user doesn't own the app.

// entrancesEnvelope unwraps the {items: [...]} payload BFL returns
// inside the envelope's data field for /api/applications/<app>/entrances.
type entrancesEnvelope struct {
	Items []appEntrance `json:"items"`
}

// NewEntrancesCommand returns the `settings apps entrances` parent.
// Phase 3 ships only `list`; the per-entrance setup verbs live as
// top-level commands (`apps domain`, `apps policy`, `apps auth-level`)
// rather than nested under `entrances` because the SPA flattens them
// into separate side-panels rather than nesting under a single
// "entrance" tab.
func NewEntrancesCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "entrances",
		Short: "list per-entrance metadata for an app (Settings -> App -> Entrance)",
		Long: `Inspect the entrance vector an app exposes — the visible UIs / SSO
endpoints other users can reach. Each entrance has its own state
(running / stopped / etc.), authorization level (private / public /
internal), and optionally a custom domain.

Subcommands:
  list <app>    list every entrance the app exposes

For per-entrance setup (custom domain, two-factor policy, auth level)
see the sibling commands:
  olares-cli settings apps domain      get|set|finish <app> <entrance>
  olares-cli settings apps policy      get|set        <app> <entrance>
  olares-cli settings apps auth-level  set            <app> <entrance>
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newEntrancesListCommand(f))
	return cmd
}

func newEntrancesListCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "list <app>",
		Short: "list per-entrance metadata for an app",
		Long: `List every entrance the app exposes, with the same NAME / TITLE / STATE
/ AUTH LEVEL / INVISIBLE / URL columns "apps get" renders inline.
Useful when scripting against a specific entrance — for example, to
discover the entrance name to pass to "apps auth-level set".

The data comes from a FRESH /api/applications/<app>/entrances call, not
the cached snapshot inside /api/myapps; if you've just toggled an
auth-level via this CLI and want to confirm, call this verb (rather
than "apps get") for an authoritative read.

Pass --output json for the full TerminusEntrance[] vector including
icon, openMethod, and any per-entrance fields the SPA exposes.

Examples:
  olares-cli settings apps entrances list files
  olares-cli settings apps entrances list files -o json
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runEntrancesList(c.Context(), f, args[0], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runEntrancesList(ctx context.Context, f *cmdutil.Factory, app, outputRaw string) error {
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
	path := "/api/applications/" + url.PathEscape(app) + "/entrances"
	var env entrancesEnvelope
	if err := doGetEnvelope(ctx, pc.doer, path, &env); err != nil {
		return err
	}
	if env.Items == nil {
		env.Items = []appEntrance{}
	}
	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, env.Items)
	default:
		return renderEntrancesList(os.Stdout, app, env.Items)
	}
}

func renderEntrancesList(w io.Writer, app string, entries []appEntrance) error {
	if len(entries) == 0 {
		_, err := fmt.Fprintf(w, "no entrances for app %q\n", app)
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "NAME\tTITLE\tSTATE\tAUTH LEVEL\tINVISIBLE\tURL"); err != nil {
		return err
	}
	for _, e := range entries {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
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
