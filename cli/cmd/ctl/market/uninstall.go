package market

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

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

	resp, err := mc.UninstallApp(ctx, appName, cascade, opts.DeleteData)
	if err != nil {
		return opts.failOp("uninstall", appName, err)
	}

	result := newOperationResult(mc, "uninstall", appName, "", "", "uninstall requested", resp)
	// Uninstall is unique: the row may simply disappear from /market/state
	// once the backend cleans it up, so the watch target opts in to the
	// "absent means success (provided we saw it earlier)" shortcut.
	return runWithWatch(opts, mc, result, newWatchTarget(watchUninstall, appName, opts.Source))
}

// shouldAutoCascade decides the default value of --cascade when the user
// did not pass the flag, mirroring the SPA's csAppUninstall() default:
// turn cascade ON when the cluster is single-user AND the app is a v2
// multi-chart bundle (isCSV2). Both lookups failing is non-fatal: we
// return false (the existing default) without surfacing the error, so
// uninstall still proceeds and the backend's own validation has the
// final say. The why string is only meaningful on the true path —
// callers should ignore it when auto is false.
//
// Order matters for cost: user-count probe is one HTTP call against
// /api/users/v2, the CS probe is two (/market/state to discover the
// app's source + /apps to read its catalog metadata). Skipping the CS
// probe when the cluster is multi-user is the cheap fast path.
func shouldAutoCascade(ctx context.Context, opts *MarketOptions, mc *MarketClient, appName string) (bool, string) {
	totals, err := fetchUserTotals(ctx, opts)
	if err != nil || totals == 0 {
		return false, ""
	}
	if totals > 1 {
		return false, ""
	}

	source, err := lookupAppSource(ctx, mc, appName)
	if err != nil || source == "" {
		return false, ""
	}

	appInfo, err := fetchAppInfo(ctx, mc, appName, source)
	if err != nil {
		return false, ""
	}
	if !isCSV2(appInfo) {
		return false, ""
	}
	return true, fmt.Sprintf("single-user instance + v2 multi-chart app (source %q)", source)
}

// lookupAppSource finds which Market source currently carries appName's
// per-user state row, so the auto-cascade probe knows where to read its
// catalog metadata from for the isCSV2 check. Empty string + nil error
// means "not currently installed" — that is a valid signal (the backend
// will surface a clean not-found error from the DELETE itself), so we
// don't pretend it's a CS app.
func lookupAppSource(ctx context.Context, mc *MarketClient, appName string) (string, error) {
	resp, err := mc.GetMarketState(ctx)
	if err != nil {
		return "", err
	}
	rows, err := parseStatusRows(resp, "", true)
	if err != nil {
		return "", err
	}
	for _, r := range rows {
		if r.Name == appName {
			return r.Source, nil
		}
	}
	return "", nil
}
