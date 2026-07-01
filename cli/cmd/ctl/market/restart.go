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

func NewCmdMarketRestart(f *cmdutil.Factory) *cobra.Command {
	opts := newMarketOptions(f)
	cmd := &cobra.Command{
		Use:   "restart {app-name}",
		Short: "Restart a running app (sends POST /apps/restart)",
		Long: `Restart an installed application.

Source is implicit: restart acts on whichever per-user state row matches
the app name, regardless of source (no -s flag exposed).

The restart endpoint is version-agnostic (the SPA shares one body between
resume and restart): POST /apps/restart with {app_name, source}. On
Olares 1.12.6+ an app that uses a GPU/accelerator may need a device
selected; use --compute-binding <node>:<device>[:<mem>] (repeatable) to
pin it, or answer the interactive prompt. On 1.12.5 --compute-binding is
rejected.

--watch blocks until the row settles back at 'running' (or one of the
resumeFailed / resumingCanceled / resumingCancelFailed failure states).

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

	// Resolve the installed app's source: enforces the "only operate on an
	// installed app" guard and yields the source the restart body needs
	// (restart exposes no -s). The resolved source also sharpens --watch.
	source, err := resolveInstalledSource(ctx, opts, mc, appName)
	if err != nil {
		return opts.failOp("restart", appName, err)
	}

	atLeast126, err := opts.factory.OlaresBackendAtLeast(ctx, "1.12.6")
	if err != nil {
		return opts.failOp("restart", appName, err)
	}

	// --compute-binding is a 1.12.6+ concept; reject it on 1.12.5.
	bindings, err := parseComputeBindingFlags(opts.ComputeBinding)
	if err != nil {
		return opts.failOp("restart", appName, err)
	}
	if len(bindings) > 0 && !atLeast126 {
		return opts.failOp("restart", appName, fmt.Errorf("--compute-binding requires Olares 1.12.6+; re-run without --compute-binding"))
	}

	resp, err := sendRestart(ctx, mc, appName, source, bindings)
	// 1.12.6+: recover once from a computeBindingRequired / Unavailable 422,
	// exactly as `market resume` does.
	if err != nil && atLeast126 {
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
	return runWithWatch(opts, mc, result, newWatchTarget(watchResume, appName, source))
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
