package market

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/market/uninstall"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func NewCmdMarketUninstall(f *cmdutil.Factory) *cobra.Command {
	opts := newMarketOptions(f)
	cmd := &cobra.Command{
		Use:     "uninstall {app-name}",
		Aliases: []string{"remove", "rm"},
		Short:   "Uninstall an app (DELETE /apps/{name})",
		Long: `Uninstall an application. Source is implicit — uninstall acts
on whichever per-user state row matches the app name (no -s exposed).

For C/S (client/server) v2 multi-chart apps, --cascade controls whether
the shared sub-charts are torn down alongside the user's own chart
(the JSON payload field is "all"). Default behavior mirrors the Market
SPA's csAppUninstall() dialog:

  - --cascade NOT passed: auto-decided. When the cluster has a single
    user AND the target app is a v2 multi-chart bundle (isCSV2 — see
    SPA's constants.ts), default to --cascade=true; otherwise default
    to false. A short reason is printed on stderr when the
    auto-decision flips the default to true. Probe errors (user count
    or app info) soft-fail to --cascade=false; the backend has the
    final say either way.
  - --cascade or --cascade=true: force enabled.
  - --cascade=false: force disabled (the canonical override for the
    single-user CS auto-default).

Use --delete-data to also remove the app's persistent data.

--watch blocks until the row reaches 'uninstalled' or disappears
entirely (acceptInitialAbsent=true: if the user's per-user row is
ALREADY gone — e.g. a previous non-cascading uninstall — watch
returns immediately rather than hanging waiting for a row that no
longer exists).

Examples:
  olares-cli market uninstall firefox                              # auto-cascade for single-user CS apps
  olares-cli market uninstall firefox --cascade=false              # force no cascade on a CS app
  olares-cli market uninstall firefox --cascade --delete-data
  olares-cli market uninstall firefox --watch                      # block until row uninstalled / absent
  olares-cli market uninstall firefox --watch -o json -q           # silent; exit code = verdict
  olares-cli market uninstall firefox --watch --watch-interval 1s --watch-timeout 10m`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUninstall(opts, cmd, args[0])
		},
	}
	opts.addOutputFlags(cmd)
	opts.addCascadeFlag(cmd)
	opts.addDeleteDataFlag(cmd)
	opts.addWatchFlags(cmd)
	return cmd
}

func runUninstall(opts *MarketOptions, cmd *cobra.Command, appName string) error {
	mc, err := opts.prepare()
	if err != nil {
		return opts.failOp("uninstall", appName, err)
	}

	ctx := context.Background()
	cascade := opts.Cascade
	cascadeExplicit := cmd != nil && cmd.Flags().Changed("cascade")
	if !cascadeExplicit {
		// Auto-cascade only ever flips OFF → ON (never the other way),
		// matching the SPA's "default all:true for single-user CS apps,
		// otherwise default all:false" rule in csAppUninstall(). On any
		// detection error we soft-fail to the original false default —
		// the user can always pass --cascade=true explicitly.
		if auto, why := shouldAutoCascade(ctx, opts, mc, appName); auto {
			cascade = true
			opts.info("--cascade auto-enabled: %s (pass --cascade=false to override)", why)
		}
	}

	opts.info("Uninstalling '%s' for user '%s'...", appName, mc.olaresID)
	if cascade {
		opts.info("  --cascade: will uninstall all sub-charts")
	}
	if opts.DeleteData {
		opts.info("  --delete-data: will delete persistent data")
	}

	// Resolve the installed app's source: enforces the "only operate on an
	// installed app" guard (a bugfix applying to 1.12.5 and 1.12.6 alike)
	// and yields the source the 1.12.6 body needs (uninstall exposes no -s).
	// The resolved source also sharpens the --watch state-row match.
	source, err := resolveInstalledSource(ctx, opts, mc, appName)
	if err != nil {
		return opts.failOp("uninstall", appName, err)
	}

	atLeast126, err := opts.factory.OlaresBackendAtLeast(ctx, "1.12.6")
	if err != nil {
		return opts.failOp("uninstall", appName, err)
	}

	// 1.12.6 added app_name + source to the uninstall body (this is the
	// change that broke `market uninstall` against the new backend);
	// opts.Version is empty unless a future --version flag supplies one
	// (the 1.12.6 builder includes it when set, the 1.12.5 builder ignores
	// it).
	method, path, body := uninstall.Build(atLeast126, appName, source, opts.Version, cascade, opts.DeleteData)
	resp, err := mc.doRequest(ctx, method, path, body)
	if err != nil {
		return opts.failOp("uninstall", appName, err)
	}

	result := newOperationResult(mc, "uninstall", appName, "", "", "uninstall requested", resp)
	// Uninstall is unique: the row may simply disappear from /market/state
	// once the backend cleans it up, so the watch target opts in to the
	// "absent means success (provided we saw it earlier)" shortcut.
	return runWithWatch(opts, mc, result, newWatchTarget(watchUninstall, appName, source))
}

// shouldAutoCascade decides the default value of --cascade when the user
// did not pass the flag, mirroring the SPA's csAppUninstall() / csAppStop()
// default: turn cascade ON when the cluster is single-user AND the app
// is a v2 multi-chart bundle (isCSV2). All probes failing is non-fatal:
// we return false (the existing default) without surfacing the error,
// so uninstall / stop still proceeds and the backend's own validation
// has the final say. The why string is only meaningful on the true
// path — callers should ignore it when auto is false.
//
// Order matters for cost: user-count probe is one HTTP call against
// /api/users/v2, the CS probe is two (/market/state to discover the
// app's source + /apps to read its catalog metadata). Skipping the CS
// probe when the cluster is multi-user is the cheap fast path.
func shouldAutoCascade(ctx context.Context, opts *MarketOptions, mc *MarketClient, appName string) (bool, string) {
	return shouldAutoCascadeWith(ctx, newCascadeProbe(opts, mc), appName)
}

// cascadeProbe is the unit-testable seam for shouldAutoCascade. The
// three closures map 1:1 to the production calls (fetchUserTotals
// against /api/users/v2 on DesktopURL, lookupInstalledApp against
// /market/state, fetchAppInfo against /apps). Tests construct a
// cascadeProbe directly to drive shouldAutoCascadeWith without having
// to stand up a full cmdutil.Factory + http stack.
type cascadeProbe struct {
	fetchTotals  func(ctx context.Context) (int, error)
	lookupRow    func(ctx context.Context, appName string) (*installedAppRow, error)
	fetchAppMeta func(ctx context.Context, name, source string) (map[string]interface{}, error)
}

func newCascadeProbe(opts *MarketOptions, mc *MarketClient) cascadeProbe {
	return cascadeProbe{
		fetchTotals: func(ctx context.Context) (int, error) { return fetchUserTotals(ctx, opts) },
		lookupRow: func(ctx context.Context, name string) (*installedAppRow, error) {
			return lookupInstalledApp(ctx, mc, name)
		},
		fetchAppMeta: func(ctx context.Context, name, source string) (map[string]interface{}, error) {
			return fetchAppInfo(ctx, mc, name, source)
		},
	}
}

// shouldAutoCascadeWith is the testable core of shouldAutoCascade.
//
// Critical clone parity bit: the catalog (/apps) is indexed by the
// source app name (e.g. `windows`), NOT by the per-instance clone name
// the user typed (`windowsefe992`). The state row carries both — the
// canonical row.Name and a row.RawName that, for clones, is the
// source app. We MUST use RawName for the isCSV2 catalog lookup
// whenever it differs from Name, otherwise the /apps response comes
// back empty, isCSV2 returns false, and the auto-cascade default
// silently diverges from the SPA's csAppUninstall() / csAppStop()
// (which read app_info.app_entry.{apiVersion,subCharts} from the
// SOURCE app's `AppFullInfo` — that's also keyed by RawName under the
// hood). The same RawName-preferred catalog-key trick is used by
// `preflightUpgrade` (preflight.go) and `fetchInstalledApps`
// (list.go) — keep all three in lockstep.
func shouldAutoCascadeWith(ctx context.Context, p cascadeProbe, appName string) (bool, string) {
	totals, err := p.fetchTotals(ctx)
	if err != nil || totals == 0 {
		return false, ""
	}
	if totals > 1 {
		return false, ""
	}

	row, err := p.lookupRow(ctx, appName)
	if err != nil || row == nil || row.Source == "" {
		return false, ""
	}

	lookupName := strings.TrimSpace(row.RawName)
	if lookupName == "" {
		lookupName = row.Name
	}
	if lookupName == "" {
		lookupName = appName
	}

	appInfo, err := p.fetchAppMeta(ctx, lookupName, row.Source)
	if err != nil {
		return false, ""
	}
	if !isCSV2(appInfo) {
		return false, ""
	}
	// Surface the catalog name (RawName for clones) in the reason
	// string so the stderr hint matches what the user would see in the
	// SPA's dialog — a clone like windowsefe992 reads "via source app
	// 'windows'" which makes the cascade decision auditable.
	if lookupName != "" && lookupName != appName {
		return true, fmt.Sprintf("single-user instance + v2 multi-chart app (via source app %q in source %q)", lookupName, row.Source)
	}
	return true, fmt.Sprintf("single-user instance + v2 multi-chart app (source %q)", row.Source)
}
