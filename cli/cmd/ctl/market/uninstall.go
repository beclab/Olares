package market

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/market/cancel"
	"github.com/beclab/Olares/cli/cmd/ctl/market/uninstall"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// inFlightCancelableStates mirrors app-service's CancelableStates: while the
// app is mid-operation, app-service rejects a direct uninstall and only
// accepts a cancel. `market uninstall` orchestrates around this (cancel
// first, then finish with a real uninstall if the cancel only stopped the
// app) so "uninstall == fully remove" holds regardless of state.
var inFlightCancelableStates = map[string]bool{
	"pending":      true,
	"downloading":  true,
	"installing":   true,
	"initializing": true,
	"upgrading":    true,
	"applyingEnv":  true,
	stateResuming:  true,
}

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

On Olares 1.12.6+ a CS/shared app (simpleInfo apiVersion=='v2' || shared)
is ALWAYS cascaded — the backend forces all=true and the SPA disables
the checkbox — so --cascade=false is overridden (a stderr note reports
the force). The --cascade=false override only takes effect on 1.12.5.

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

	atLeast126, err := opts.factory.OlaresBackendAtLeast(ctx, "1.12.6")
	if err != nil {
		return opts.failOp("uninstall", appName, err)
	}

	// Uninstall must stay idempotent: unlike stop/resume it is a valid flow
	// to re-run after the per-user row is already gone (e.g. `--cascade` to
	// tear down the shared sub-charts of a CS app after a prior uninstall
	// cleared the per-user row) or to clean up a half-installed / failed
	// row. So source is resolved leniently here instead of going through
	// resolveInstalledSource's strict "must be installed" guard — an absent
	// row must NOT abort the command, otherwise the watcher's
	// acceptInitialAbsent "already gone == success" path can never run.
	//
	// Caveat for 1.12.6: that body REQUIRES source, which (absent --source)
	// only the per-user row can supply. Once the row is gone we can't build a
	// valid request, so the `--cascade` re-run is only actionable when the
	// caller passes --source explicitly — see the source=="" guard below.
	//
	// version is read from the same per-user state row and fed to the 1.12.6
	// body (the SPA's onUninstall() sends the installed version). The 1.12.6
	// builder includes it only when non-empty; the 1.12.5 builder ignores it.
	source := strings.TrimSpace(opts.Source)
	version := strings.TrimSpace(opts.Version)
	var curState string
	row, lookupErr := lookupInstalledApp(ctx, mc, appName)
	if lookupErr != nil {
		return opts.failOp("uninstall", appName, lookupErr)
	}
	if row != nil {
		if source == "" {
			source = strings.TrimSpace(row.Source)
		}
		if version == "" {
			version = strings.TrimSpace(row.Version)
		}
		curState = strings.TrimSpace(row.State)
	}

	cascadeExplicit := cmd != nil && cmd.Flags().Changed("cascade")
	cascade, why := resolveCascade(ctx, opts, mc, appName, atLeast126, opts.Cascade, cascadeExplicit)
	if why != "" {
		if cascadeExplicit && !opts.Cascade {
			// 1.12.6 force-true: CS/shared apps always cascade, so an
			// explicit --cascade=false is overridden (matches the SPA
			// disabling the checkbox for CS/shared apps).
			opts.info("--cascade force-enabled: %s (CS/shared apps always cascade)", why)
		} else {
			opts.info("--cascade auto-enabled: %s", why)
		}
	}

	// In-flight app: app-service rejects a direct uninstall while an
	// operation is in progress and only accepts cancel. Orchestrate the full
	// removal from the CLI (cancel first, then a real uninstall if the cancel
	// only stopped the app) so "uninstall == fully remove" holds regardless
	// of state.
	if row != nil && inFlightCancelableStates[curState] {
		return runUninstallViaCancel(opts, mc, appName, source, version, cascade, curState, atLeast126)
	}

	opts.info("Uninstalling '%s' for user '%s'...", appName, mc.olaresID)
	if cascade {
		opts.info("  --cascade: will uninstall all sub-charts")
	}
	if opts.DeleteData {
		opts.info("  --delete-data: will delete persistent data")
	}

	// 1.12.6 requires source in the uninstall body, and the only place we
	// can learn it is the user's own install state. If the app isn't
	// installed for this user we can't know its source (or even whether it
	// was a multi-chart CS app), so there is nothing actionable to delete:
	// report an idempotent success and let --watch's acceptInitialAbsent
	// confirm it, rather than sending an invalid (sourceless) request or
	// erroring out.
	//
	// Exception: when the caller explicitly asked to cascade, they are
	// almost certainly re-running uninstall to tear down shared sub-charts
	// left behind after a prior uninstall already cleared the per-user row
	// (the "re-run --cascade" flow documented above). Silently reporting
	// "nothing to uninstall" there would hide the leftover shared charts, so
	// surface an actionable error pointing at --source — the one way to
	// supply the source the 1.12.6 body needs once the row is gone.
	if atLeast126 && source == "" {
		if cascade {
			return opts.failOp("uninstall", appName, fmt.Errorf(
				"'%s' is no longer in your installed apps, so its source can't be resolved; re-run with --source <name> to cascade-clean its shared sub-charts", appName))
		}
		opts.info("'%s' is not installed for this user; nothing to uninstall", appName)
		result := newOperationResult(mc, "uninstall", appName, "", "", "not installed; nothing to uninstall", nil)
		return runWithWatch(opts, mc, result, newWatchTarget(watchUninstall, appName, source))
	}

	method, path, body := uninstall.Build(atLeast126, appName, source, version, cascade, opts.DeleteData)
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

// runUninstallViaCancel handles `uninstall` against an in-flight app.
// app-service rejects a direct uninstall while an operation is in progress
// and only accepts cancel, so we orchestrate the removal entirely from the
// CLI:
//
//  1. cancel and block until the row settles;
//  2. if the cancel tore the app down (the pending/downloading/installing
//     flow cancels into a *Canceled state with the namespace deleted, or the
//     row reaches uninstalled / vanishes) we are done — that is equivalent to
//     uninstalled;
//  3. if the cancel only stopped the app (initializing/upgrading/applyingEnv/
//     resuming cancel into stopped, or a race left it running) the app is
//     still present, so issue the real uninstall — now allowed from
//     stopped/running — and finish under --watch.
//
// The cancel step always blocks (it must, to decide step 3) regardless of
// --watch.
func runUninstallViaCancel(opts *MarketOptions, mc *MarketClient, appName, source, version string, cascade bool, curState string, atLeast126 bool) error {
	ctx := context.Background()

	opts.info("'%s' is %s (operation in progress); canceling it first...", appName, curState)
	cancelMethod, cancelPath, cancelBody := cancel.Build(atLeast126, appName, source, version)
	if _, err := mc.doRequest(ctx, cancelMethod, cancelPath, cancelBody); err != nil {
		return opts.failOp("uninstall", appName, fmt.Errorf("cancel in-progress operation: %w", err))
	}

	// watchCancel's success set is the widest in the tree (any "stopped
	// moving" state); it fails only on *CancelFailed.
	cancelRow, err := waitForTerminal(ctx, mc, opts, newWatchTarget(watchCancel, appName, source))
	if err != nil {
		return finishWatchError(opts, mc, "uninstall", appName, source, err)
	}
	settled := strings.TrimSpace(cancelRow.State)

	// Install-flow cancel deletes the namespace -> the app is gone. Treat
	// *Canceled / uninstalled / a vanished row as a completed uninstall.
	if settled == "" || settled == "uninstalled" || canceledStates[settled] {
		final := newOperationResult(mc, "uninstall", appName, source, "", "", nil)
		final.Status = "success"
		final.State = settled
		final.FinalState = settled
		final.Source = source
		final.Message = fmt.Sprintf("uninstall completed via cancel (state=%s)", valueOrUnknown(settled))
		if !opts.Quiet {
			opts.printResult(final)
		}
		return nil
	}

	// Cancel only stopped the app (or raced and left it running); the app is
	// still present, so finish the job with a real uninstall.
	opts.info("'%s' settled at %s after cancel; issuing uninstall to fully remove it...", appName, settled)

	method, path, body := uninstall.Build(atLeast126, appName, source, version, cascade, opts.DeleteData)
	resp, err := mc.doRequest(ctx, method, path, body)
	if err != nil {
		return opts.failOp("uninstall", appName, err)
	}
	result := newOperationResult(mc, "uninstall", appName, source, "", "uninstall requested", resp)
	return runWithWatch(opts, mc, result, newWatchTarget(watchUninstall, appName, source))
}

// finishWatchError renders a terminal watch error (failure / timeout) as a
// failed OperationResult and returns errReported, mirroring runWithWatch's
// error path so multi-step flows surface structured output too.
func finishWatchError(opts *MarketOptions, mc *MarketClient, op, appName, source string, err error) error {
	failed := newOperationResult(mc, op, appName, source, "", "", nil)
	failed.Status = "failed"
	failed.Message = err.Error()
	var fail *watchFailureError
	if errors.As(err, &fail) {
		failed.State = fail.row.State
		failed.Progress = fail.row.Progress
		failed.FinalState = fail.row.State
		failed.FinalOpType = fail.row.OpType
	}
	var to *watchTimeoutError
	if errors.As(err, &to) && to.last != nil {
		failed.State = to.last.State
		failed.Progress = to.last.Progress
		failed.FinalState = to.last.State
		failed.FinalOpType = to.last.OpType
	}
	opts.printResult(failed)
	return errReported
}

// resolveCascade computes the final value of the `all` flag for uninstall /
// stop, branching on backend version:
//
//   - 1.12.6: the SPA forces the cascade ON (and disables the checkbox) for any
//     CS/shared app — uninstallChoicePrompt sets all=isCsV2 with allDisabled,
//     stopChoicePrompt forces all=true for CS v2 — where CS/shared is now read
//     from simpleInfo (apiVersion=='v2' || shared). We mirror that with
//     force-true: a CS/shared app cascades regardless of --cascade=false. A
//     non-CS app keeps the user's value (default false) since the flag is
//     meaningless there.
//   - 1.12.5: legacy behavior — only auto-enable (single-user + isCSV2 by
//     subCharts) when the user did NOT pass --cascade; an explicit value wins.
//
// The returned reason string is only meaningful (non-empty) when the function
// flipped the default ON; callers print it on stderr.
func resolveCascade(ctx context.Context, opts *MarketOptions, mc *MarketClient, appName string, atLeast126, userCascade, userExplicit bool) (bool, string) {
	if atLeast126 {
		if cs, why := isCsOrSharedWith(ctx, newCascadeProbe(opts, mc), appName); cs {
			return true, why
		}
		return userCascade, ""
	}
	if userExplicit {
		return userCascade, ""
	}
	if auto, why := shouldAutoCascadeWith(ctx, newCascadeProbe(opts, mc), appName); auto {
		return true, why
	}
	return userCascade, ""
}

// isCsOrSharedWith reports whether appName is a 1.12.6 CS/shared app, reading
// the predicate from simpleInfo (isCsOrSharedFromSimple). It reuses the same
// RawName-preferred catalog lookup as shouldAutoCascadeWith so clones resolve
// against their source app's catalog entry. All probe failures are non-fatal
// (return false): the backend's own validation has the final say.
func isCsOrSharedWith(ctx context.Context, p cascadeProbe, appName string) (bool, string) {
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
	if !isCsOrSharedFromSimple(appInfo) {
		return false, ""
	}
	if lookupName != "" && lookupName != appName {
		return true, fmt.Sprintf("CS/shared app (via source app %q in source %q)", lookupName, row.Source)
	}
	return true, fmt.Sprintf("CS/shared app (source %q)", row.Source)
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
