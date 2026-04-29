package apps

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings apps list`
//
// Wraps user-service's GET /api/myapps (init.controller.ts:52).
//
// Server flow:
//   user-service /api/myapps → AppService.GetMyApps() →
//   BFL /bfl/backend/v1/myapps (POST internally) → returns
//   `{code: 0, data: {items: [AppInfo, ...]}}`. user-service unwraps
//   `data.data.items` into AppInfo[] and re-wraps with returnSucceed,
//   so the CLI receives a plain BFL envelope wrapping the array:
//
//     { "code": 0, "message": "Success", "data": [AppInfo, ...] }
//
// Note: this is the same SPA call surface the Settings UI uses to
// populate "Settings -> Apps". It overlaps in data with
// `olares-cli market list` but the focus is different: market is about
// install lifecycle (versions, sources), settings apps is about
// post-install configuration (entrances, ports, owner). We keep both
// surfaces because the SPA does too.
//
// Role: any authenticated user can list their own apps; no preflight.
func NewListCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output  string
		all     bool
		showSys bool
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list installed apps for the current user (Settings -> Apps)",
		Long: `List installed applications for the active profile's user.

By default this filters out apps in uninstalled / pending states (matching
the SPA's "filterUnInstalledMyApps" behavior). Pass --all to include
every app the cluster reports — useful for triaging stuck installs.

System apps (Files / Settings / Vault / etc.) are hidden by default
because they're not user-actionable; pass --show-system to include them.

Pass --output json for the full AppInfo struct including entrances,
ports, owner, namespace, and isClusterScoped — useful for scripting
against per-app config.
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runList(c.Context(), f, output, all, showSys)
		},
	}
	addOutputFlag(cmd, &output)
	cmd.Flags().BoolVar(&all, "all", false, "include apps in uninstalled / pending states")
	cmd.Flags().BoolVar(&showSys, "show-system", false, "include system apps (Files, Settings, Vault, etc.)")
	return cmd
}

// appEntrance mirrors AppInfo.entrances element shape from
// user-service/src/app.service.ts:118-130. Defined here rather than in a
// shared types file because Phase 1 only renders a small subset; later
// phases will add their own narrower types as the verbs land.
type appEntrance struct {
	AuthLevel  string `json:"authLevel"`
	Icon       string `json:"icon"`
	ID         string `json:"id"`
	Invisible  bool   `json:"invisible"`
	Name       string `json:"name"`
	OpenMethod string `json:"openMethod"`
	Reason     string `json:"reason"`
	State      string `json:"state"`
	Title      string `json:"title"`
	URL        string `json:"url"`
}

// servicePort mirrors BFL's ServicePort struct
// (framework/bfl/pkg/app_service/v1/types.go:48-60). The wire shape has
// been a struct array since 60d37998 (2025-12-11, BFL→main repo merge);
// CLI shipped with `[]int` by mistake and only got away with it while
// every probed app had ports=[] (empty arrays decode into any slice).
// Once a real ports entry hit the wire, decode failed with
// `cannot unmarshal object into Go struct field appInfo.ports of type int`
// (KI-18, first surfaced in scripts/local_report_phase14b.md).
type servicePort struct {
	Name       string `json:"name"`
	Host       string `json:"host"`
	Port       int32  `json:"port"`
	ExposePort int32  `json:"exposePort,omitempty"`
	Protocol   string `json:"protocol,omitempty"` // "tcp" | "udp" (BFL default tcp)
}

// appInfo mirrors user-service's AppInfo struct (app.service.ts:116).
// Field-for-field JSON shape so --output json is a 1:1 passthrough.
type appInfo struct {
	Deployment       string        `json:"deployment"`
	Entrances        []appEntrance `json:"entrances"`
	SharedEntrances  []appEntrance `json:"sharedEntrances"`
	Icon             string        `json:"icon"`
	ID               string        `json:"id"`
	IsClusterScoped  bool          `json:"isClusterScoped"`
	IsSysApp         bool          `json:"isSysApp"`
	MobileSupported  bool          `json:"mobileSupported"`
	Name             string        `json:"name"`
	Namespace        string        `json:"namespace"`
	Owner            string        `json:"owner"`
	Ports            []servicePort `json:"ports"`
	RequiredGpu      string        `json:"requiredGpu"`
	State            string        `json:"state"`
	Target           string        `json:"target"`
	Title            string        `json:"title"`
	URL              string        `json:"url"`
}

// uninstalledStates mirrors the SPA-side uninstalledAppStates set used
// by stores/settings/application.ts's uninstalledAppState getter. We
// keep the values verbatim — if upstream adds a new "uninstalled-ish"
// state, this list must be updated to keep --all consistent with
// what the SPA hides.
var uninstalledStates = map[string]struct{}{
	"uninstalled":  {},
	"installing":   {},
	"pending":      {},
	"upgrading":    {},
	"reinstalling": {},
}

func isUninstalled(state string) bool {
	_, ok := uninstalledStates[strings.ToLower(state)]
	return ok
}

func runList(ctx context.Context, f *cmdutil.Factory, outputRaw string, includeAll, includeSys bool) error {
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

	var rows []appInfo
	if err := doGetEnvelope(ctx, pc.doer, "/api/myapps", &rows); err != nil {
		return err
	}

	filtered := make([]appInfo, 0, len(rows))
	for _, r := range rows {
		if !includeAll && isUninstalled(r.State) {
			continue
		}
		if !includeSys && r.IsSysApp {
			continue
		}
		filtered = append(filtered, r)
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, filtered)
	default:
		return renderAppsTable(os.Stdout, filtered)
	}
}

func renderAppsTable(w io.Writer, rows []appInfo) error {
	if len(rows) == 0 {
		_, err := fmt.Fprintln(w, "no apps")
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "NAME\tTITLE\tSTATE\tOWNER\tENTRANCES\tURL"); err != nil {
		return err
	}
	for _, r := range rows {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%d\t%s\n",
			nonEmpty(r.Name),
			nonEmpty(r.Title),
			nonEmpty(r.State),
			nonEmpty(r.Owner),
			len(r.Entrances),
			nonEmpty(r.URL),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}
