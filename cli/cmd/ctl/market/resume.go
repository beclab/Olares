package market

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/market/resume"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func NewCmdMarketResume(f *cmdutil.Factory) *cobra.Command {
	opts := newMarketOptions(f)
	cmd := &cobra.Command{
		Use:     "resume {app-name}",
		Aliases: []string{"start"},
		Short:   "Resume a stopped app (sends POST /apps/resume)",
		Long: `Resume a stopped (suspended) application.

Source is implicit: resume acts on whichever per-user state row matches
the app name, regardless of source (no -s flag exposed).

--watch blocks until the row settles at 'running' (or one of the
resumeFailed / resumingCanceled / resumingCancelFailed failure states).
The watcher is idempotent: 'resume' against an already-running row
returns immediately with success ({state=running, opType=""}), rather
than hanging until --watch-timeout fires.

Compute binding (Olares 1.12.6+ only): apps that use a GPU/accelerator
may need a device selected when resumed. Use --compute-binding
<node>:<device>[:<mem>] (repeatable) to pin it. The optional mem is a
MemorySlice allocation: a bare number is Gi, or add a Gi/Mi suffix
(e.g. node:gpu-0:8, node:gpu-0:512Mi). If omitted and the backend asks
for a binding, an interactive terminal prompts you to choose from the
available devices, while a non-interactive session fails with the list so
you can re-run with the flag. <node> and <device> are the NODE / DEVICE-ID
values from 'olares-cli settings compute list'. On Olares 1.12.5 the resume
path is unchanged and --compute-binding is rejected.

Examples:
  olares-cli market resume firefox                         # fire-and-forget; returns once backend accepts
  olares-cli market resume firefox --watch                 # block until row reports running
  olares-cli market resume firefox --watch -o json         # JSON OperationResult with finalState/finalOpType
  olares-cli market resume firefox --watch -q              # silent; exit code only
  olares-cli market resume comfyui --compute-binding node-1:gpu-0 --watch         # pin a device (1.12.6+)
  olares-cli market resume comfyui --compute-binding node-1:gpu-0:8 --watch       # MemorySlice: 8 Gi
  olares-cli market resume firefox --watch --watch-interval 1s --watch-timeout 2m`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runResume(opts, args[0])
		},
	}
	opts.addOutputFlags(cmd)
	opts.addComputeBindingFlag(cmd)
	opts.addWatchFlags(cmd)
	return cmd
}

func runResume(opts *MarketOptions, appName string) error {
	mc, err := opts.prepare()
	if err != nil {
		return opts.failOp("resume", appName, err)
	}

	opts.info("Resuming '%s' for user '%s'...", appName, mc.olaresID)

	ctx := context.Background()

	// Resolve the installed app's source: enforces the "only operate on an
	// installed app" guard (a bugfix applying to 1.12.5 and 1.12.6 alike)
	// and yields the source the 1.12.6 body needs (resume exposes no -s).
	// The resolved source also sharpens the --watch state-row match.
	source, err := resolveInstalledSource(ctx, opts, mc, appName)
	if err != nil {
		return opts.failOp("resume", appName, err)
	}

	atLeast126, err := opts.factory.OlaresBackendAtLeast(ctx, "1.12.6")
	if err != nil {
		return opts.failOp("resume", appName, err)
	}

	// --compute-binding is a 1.12.6+ concept. Parse it up front and reject it
	// on 1.12.5, whose resume path we deliberately leave untouched.
	bindings, err := parseComputeBindingFlags(opts.ComputeBinding)
	if err != nil {
		return opts.failOp("resume", appName, err)
	}
	if len(bindings) > 0 && !atLeast126 {
		return opts.failOp("resume", appName, fmt.Errorf("--compute-binding requires Olares 1.12.6+; this backend uses a different (unchanged) resume path — re-run without --compute-binding"))
	}

	resp, err := sendResume(ctx, mc, atLeast126, appName, source, bindings)
	// 1.12.6+: recover once from a computeBindingRequired / Unavailable 422.
	// If the user supplied an explicit --compute-binding that the backend
	// rejected, surface why instead of retrying; otherwise resolve a binding
	// (interactive prompt or actionable error) and retry once. Skipped on 1.12.5.
	if err != nil && atLeast126 {
		if checkType, raw := parseFailedCheck(resp); isComputeBindingPrompt(checkType) {
			if len(bindings) > 0 {
				return opts.failOp("resume", appName, computeBindingRejected(raw, appName, bindings))
			}
			sel, berr := resolveComputeBinding(raw, checkType, appName, opts.isInteractive())
			if berr != nil {
				return opts.failOp("resume", appName, berr)
			}
			resp, err = sendResume(ctx, mc, atLeast126, appName, source, sel)
		}
	}
	if err != nil {
		return opts.failOp("resume", appName, err)
	}

	result := newOperationResult(mc, "resume", appName, "", "", "resume requested", resp)
	return runWithWatch(opts, mc, result, newWatchTarget(watchResume, appName, source))
}

// sendResume builds the version-appropriate resume body (1.12.6 moved to
// {app_name, source, computeBinding?}; 1.12.5 keeps {appName}) and posts it.
// computeBinding is only ever populated on 1.12.6+.
func sendResume(ctx context.Context, mc *MarketClient, atLeast126 bool, appName, source string, bindings []BindingSelection) (*APIResponse, error) {
	var cb any
	if len(bindings) > 0 {
		cb = bindings
	}
	method, path, body := resume.Build(atLeast126, appName, source, cb)
	return mc.doRequest(ctx, method, path, body)
}
