package market

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/market/cancel"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

const (
	// stateResuming mirrors the SPA's APP_STATUS.RESUME.DEFAULT: the
	// per-user state-row value while a resume is in flight.
	stateResuming = "resuming"
	// resumeCancelMinVersion is the first Olares line whose resume-cancel
	// UX the CLI mirrors. The SPA exposed the cancel button on a resuming
	// app in 1.12.7 (reusing DELETE /apps/{name}/install); cancelling a
	// resuming app on older / undetectable backends is rejected fail-closed
	// so the CLI stays in lockstep with that rollout.
	resumeCancelMinVersion = "1.12.7"
)

// gateResumeCancel enforces the 1.12.7 minimum for cancelling a *resuming*
// app. It is a no-op for every other state — only resume cancellation is
// version-gated; the rest of the in-flight pipeline (pending / downloading /
// installing / upgrading / ...) cancels on any backend. Fail-closed on
// older / undetectable backends via cmdutil.RequireMinVersion.
func gateResumeCancel(ctx context.Context, f *cmdutil.Factory, state string) error {
	if strings.TrimSpace(state) != stateResuming {
		return nil
	}
	return cmdutil.RequireMinVersion(ctx, f, cmdutil.MinVersionGate{
		Verb:       "market cancel",
		MinVersion: resumeCancelMinVersion,
		Reason:     "canceling a resuming app",
	})
}

func NewCmdMarketCancel(f *cmdutil.Factory) *cobra.Command {
	opts := newMarketOptions(f)
	cmd := &cobra.Command{
		Use:   "cancel {app-name}",
		Short: "Cancel the current in-progress app operation (install / upgrade / uninstall / ...)",
		Long: `Cancel the current in-progress operation for an app
(DELETE /apps/{name}/install). Source is normally read from the
per-user state row; on 1.12.6+ pass --source when that row is absent
or /market/state is unreadable and you still need to cancel.

The cancel watcher is the widest in the market tree: any "row stopped
moving" state counts as success, including *Canceled, *Failed (the
underlying op died, cancel "won by default") and the stable resting
states running / stopped / uninstalled (cancel raced and lost, OR
rollback landed). Failure is ONLY surfaced for *CancelFailed (the
cancel request itself was rejected). This avoids the common hang where
a cancel races with downloadFailed / partial-rollback-to-stopped.

The terminal row carries the *underlying* op (install / upgrade / ...)
as its opType, not "cancel" — matchOpType is off, no race-tracking
gate applies.

Cancelling an app that is *resuming* requires Olares >= 1.12.7 (the
resume-cancel UX the SPA shipped in 1.12.7, reusing this same DELETE).
On older or undetectable backends the CLI rejects it fail-closed (pass
--olares-version to override). Every other in-flight state is
unaffected.

Examples:
  olares-cli market cancel firefox                         # fire-and-forget; returns once backend accepts
  olares-cli market cancel firefox --watch                 # block until row settles (any terminal state except *CancelFailed)
  olares-cli market cancel firefox --watch -o json         # JSON; finalState surfaces where the row actually landed
  olares-cli market cancel firefox --watch -q              # silent; exit 0 unless *CancelFailed
  olares-cli market cancel firefox --watch --watch-interval 1s --watch-timeout 2m`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCancel(opts, args[0])
		},
	}
	opts.addSourceFlag(cmd, "market source id (1.12.6+ when state row is absent)")
	opts.addOutputFlags(cmd)
	opts.addWatchFlags(cmd)
	return cmd
}

func runCancel(opts *MarketOptions, appName string) error {
	mc, err := opts.prepare()
	if err != nil {
		return opts.failOp("cancel", appName, err)
	}

	opts.info("Canceling in-progress operation for '%s' (user '%s')...", appName, mc.olaresID)

	ctx := context.Background()

	atLeast126, err := opts.factory.OlaresBackendAtLeast(ctx, "1.12.6")
	if err != nil {
		return opts.failOp("cancel", appName, err)
	}

	// 1.12.6 moved the cancel body to {app_name, source, version} (the
	// 1.12.5 body only sent {sync}, which the new backend rejects with
	// "Missing required fields: app_name is required"). Resolve source and
	// the installed version from the per-user state row when available; an
	// explicit --source wins. version is best-effort — the builder only
	// includes it when non-empty.
	//
	// On 1.12.5 the wire format does not need source, so a failed
	// /market/state read must not block cancel (the old flow never
	// depended on state). On 1.12.6 the body requires source; when the
	// row is gone and --source was not passed, report idempotent success
	// rather than sending a sourceless request the backend will reject.
	// An explicit --source bypasses the state-read failure path too.
	source := strings.TrimSpace(opts.Source)
	version := ""
	row, lookupErr := lookupInstalledApp(ctx, mc, appName)
	if lookupErr != nil && atLeast126 && source == "" {
		return opts.failOp("cancel", appName, lookupErr)
	}
	if row != nil {
		if source == "" {
			source = strings.TrimSpace(row.Source)
		}
		version = strings.TrimSpace(row.Version)
		// Cancelling a resuming app is a 1.12.7 UX; reject it fail-closed
		// on older/undetectable backends. Other in-flight states are
		// unaffected (gateResumeCancel no-ops for them).
		if err := gateResumeCancel(ctx, opts.factory, row.State); err != nil {
			return opts.failOp("cancel", appName, err)
		}
	}

	if atLeast126 && source == "" {
		opts.info("'%s' has no in-progress operation for this user; nothing to cancel", appName)
		result := newOperationResult(mc, "cancel", appName, "", "", "nothing in progress; nothing to cancel", nil)
		return finishOperation(opts, mc, result)
	}

	method, path, body := cancel.Build(atLeast126, appName, source, version)
	resp, err := mc.doRequest(ctx, method, path, body)
	if err != nil {
		return opts.failOp("cancel", appName, err)
	}

	result := newOperationResult(mc, "cancel", appName, "", "", "cancel requested", resp)
	// Cancel's terminal row carries the *underlying* OpType (install /
	// upgrade / ...), not "cancel", so the watch target opts out of
	// strict OpType matching via matchOpType=false in newWatchTarget.
	return runWithWatch(opts, mc, result, newWatchTarget(watchCancel, appName, source))
}
