package market

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/market/restart"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// appStateStopped is the per-user state-row value for a suspended app
// (APP_STATUS.STOP.COMPLETED in the SPA). RestartApp uses it to leave stopped
// apps stopped, mirroring OverlayGatewayPage's `!isStopped` guard.
const appStateStopped = "stopped"

// restartMinOlaresVersion is the first Olares line that exposes POST
// /apps/restart. Restart shipped as part of the overlay feature set in 1.12.6;
// on 1.12.5 the endpoint doesn't exist.
const restartMinOlaresVersion = "1.12.6"

func NewCmdMarketRestart(f *cmdutil.Factory) *cobra.Command {
	opts := newMarketOptions(f)
	cmd := &cobra.Command{
		Use:   "restart {app-name}",
		Short: "Restart a running app (sends POST /apps/restart)",
		Long: `Restart an installed application.

Source is implicit: restart acts on whichever per-user state row matches
the app name, regardless of source (no -s flag exposed).

Requires Olares >= 1.12.6: POST /apps/restart is part of the overlay
feature set introduced in 1.12.6 (the SPA shares one wire body between
resume and restart: {app_name, source}). On 1.12.5 the command fails fast
with a version error. An app that uses a GPU/accelerator may need a device
selected; use --compute-binding <node>:<device>[:<mem>] (repeatable) to
pin it, or answer the interactive prompt.

--watch blocks until the restart's stop-then-resume cycle actually
completes. Because a finished restart looks identical to the app's
pre-restart resting row ('running' with opType 'resume'), --watch
captures the row's statusTime before the request and only reports success
once the row is 'running' with a strictly newer statusTime. Failure is
reported for either phase (stopFailed in the stop phase; resumeFailed /
resumingCanceled / resumingCancelFailed in the resume phase).

Examples:
  olares-cli market restart firefox                    # fire-and-forget; returns once backend accepts
  olares-cli market restart firefox --watch            # block until row reports running
  olares-cli market restart comfyui --compute-binding node-1:gpu-0 --watch   # pin a device (1.12.6+)`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRestart(opts, args[0])
		},
	}
	opts.addOutputFlags(cmd)
	opts.addComputeBindingFlag(cmd)
	opts.addWatchFlags(cmd)
	return cmd
}

func runRestart(opts *MarketOptions, appName string) error {
	mc, err := opts.prepare()
	if err != nil {
		return opts.failOp("restart", appName, err)
	}

	opts.info("Restarting '%s' for user '%s'...", appName, mc.olaresID)

	ctx := context.Background()

	// `market restart` posts to /apps/restart, which is part of the overlay
	// feature set introduced in Olares 1.12.6 (the SPA shares one wire body
	// between resume and restart). Older backends don't expose the endpoint,
	// so fail fast with an actionable message instead of surfacing a confusing
	// 404 from the POST below.
	if err := requireRestartBackendVersion(ctx, opts); err != nil {
		return opts.failOp("restart", appName, err)
	}

	// Resolve the installed app's row: enforces the "only operate on an
	// installed app" guard and yields both the source the restart body needs
	// (restart exposes no -s) and the row's current statusTime, which we
	// capture as the --watch baseline BEFORE issuing the restart. A completed
	// restart rests at `running, OpType=resume` — identical to this
	// pre-restart row — so the watcher can only tell them apart by requiring
	// a strictly-newer statusTime (see watchRestart / requireNewerThanBaseline).
	row, err := resolveInstalledRow(ctx, mc, appName)
	if err != nil {
		return opts.failOp("restart", appName, err)
	}
	source := row.Source
	baselineStatusTime := parseStatusTime(row.StatusTime)

	// --compute-binding is a 1.12.6+ concept; the command is already gated to
	// >= 1.12.6 above, so no separate 1.12.5 rejection is needed here.
	bindings, err := parseComputeBindingFlags(opts.ComputeBinding)
	if err != nil {
		return opts.failOp("restart", appName, err)
	}

	resp, err := sendRestart(ctx, mc, appName, source, bindings)
	// Recover once from a computeBindingRequired / Unavailable 422, exactly as
	// `market resume` does (restart is 1.12.6+, so this path always applies).
	if err != nil {
		if checkType, raw := parseFailedCheck(resp); isComputeBindingPrompt(checkType) {
			if len(bindings) > 0 {
				return opts.failOp("restart", appName, computeBindingRejected(raw, appName, bindings))
			}
			sel, berr := resolveComputeBinding(raw, checkType, appName, opts.isInteractive())
			if berr != nil {
				return opts.failOp("restart", appName, berr)
			}
			resp, err = sendRestart(ctx, mc, appName, source, sel)
		}
	}
	if err != nil {
		return opts.failOp("restart", appName, err)
	}

	result := newOperationResult(mc, "restart", appName, "", "", "restart requested", resp)
	target := newWatchTarget(watchRestart, appName, source)
	target.baselineStatusTime = baselineStatusTime
	return runWithWatch(opts, mc, result, target)
}

// requireRestartBackendVersion gates `market restart` on Olares >=
// restartMinOlaresVersion. POST /apps/restart is part of the overlay feature
// set that landed in 1.12.6; on 1.12.5 the endpoint doesn't exist and the call
// would 404. Fail-closed (mirrors settings' RequireMinVersion / files'
// requireArchiveBackendVersion): an undetectable version is rejected because
// the feature provably doesn't exist on anything older, with --olares-version
// as the escape hatch.
func requireRestartBackendVersion(ctx context.Context, opts *MarketOptions) error {
	ok, err := opts.factory.OlaresBackendAtLeast(ctx, restartMinOlaresVersion)
	if err != nil {
		return fmt.Errorf(
			"market restart requires Olares >= %s (the overlay feature set), but the backend version could not be determined: %w; pass --%s <version> to set it manually (e.g. --%s %s)",
			restartMinOlaresVersion, err, cmdutil.FlagOlaresVersion, cmdutil.FlagOlaresVersion, restartMinOlaresVersion)
	}
	if !ok {
		got := "unknown"
		if v, verr := opts.factory.OlaresBackendVersion(ctx); verr == nil && v != nil {
			got = v.Original()
		}
		return fmt.Errorf(
			"market restart requires Olares >= %s (the overlay feature set, incl. POST /apps/restart), but this backend is %s",
			restartMinOlaresVersion, got)
	}
	return nil
}

// sendRestart posts the restart body ({app_name, source, computeBinding?}) to
// /apps/restart. computeBinding is only populated on 1.12.6+.
func sendRestart(ctx context.Context, mc *MarketClient, appName, source string, bindings []BindingSelection) (*APIResponse, error) {
	var cb any
	if len(bindings) > 0 {
		cb = bindings
	}
	method, path, body := restart.Build(appName, source, cb)
	return mc.doRequest(ctx, method, path, body)
}

// RestartApp restarts an installed app when it is currently running, mirroring
// the SPA's OverlayGatewayPage flow: after toggling an app's overlay the store
// restarts the app (via AppService.restartApp) so the change takes effect, but
// only when the app is not stopped — stopped apps keep the persisted setting and
// pick it up on their next start.
//
// It is a best-effort, fire-and-forget submit (no --watch, no compute-binding
// recovery). The return value reports whether a restart was actually issued:
//   - (false, nil): app is not installed, not in an installed state, or stopped
//     -> nothing to restart.
//   - (true, nil):  a restart request was accepted by the backend.
//   - (_, err):     the lookup or restart POST failed. Callers such as
//     `settings network overlay app enable/disable` treat this as a soft
//     warning because the primary operation (the overlay toggle) already
//     succeeded.
//
// Exported so the settings network overlay commands can reuse the market
// transport/auth path instead of re-implementing the app-store v2 client.
func RestartApp(ctx context.Context, f *cmdutil.Factory, appName string) (restarted bool, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	opts := newMarketOptions(f)
	opts.Quiet = true
	mc, err := opts.prepare()
	if err != nil {
		return false, err
	}
	row, err := lookupInstalledApp(ctx, mc, appName)
	if err != nil {
		return false, err
	}
	if row == nil || !isInstalledState(row.State) {
		return false, nil
	}
	if row.State == appStateStopped {
		return false, nil
	}
	if _, err := sendRestart(ctx, mc, appName, row.Source, nil); err != nil {
		return false, err
	}
	return true, nil
}
