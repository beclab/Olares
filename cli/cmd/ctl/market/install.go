package market

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func NewCmdMarketInstall(f *cmdutil.Factory) *cobra.Command {
	opts := newMarketOptions(f)
	cmd := &cobra.Command{
		Use:   "install {app-name}",
		Short: "Install an app from a market source (POST /apps/{name}/install)",
		Long: `Install an application from a market source.

Source resolution: -s pins the source; omitting it falls back to the
auto-selected catalog source (typically 'market.olares'). To install
a locally-uploaded chart, pass '-s upload' (the same bucket
'market upload' writes to — see 'olares-cli market upload --help').

If --version is omitted, the latest catalog version is used. The CLI
validates the version is strict semver before sending the request.

For apps that declare environment variables, use --env KEY=VALUE
(repeatable) to provide values. Required env vars not supplied will
surface a structured error from the backend (HTTP 422 / type=appenv)
that the CLI parses into a 'missing required env var(s): ...' message
listing exactly which vars and their value-source constraints.

Compute mode (Olares 1.12.6+ only): apps that can run on more than one
accelerator (e.g. cpu vs nvidia) need a mode picked at install time. Use
--compute-mode <type> to pin it; if omitted, an interactive terminal
prompts you to choose from the installable modes, while a non-interactive
session (-q, -o json, or a pipe) fails with the list so you can re-run
with the flag. On Olares 1.12.5 the install path is unchanged and
--compute-mode is rejected.

--watch blocks until the row settles at 'running' (success) or one of
the *Failed / *Canceled states (failure). Image-pull-heavy charts
(Stable Diffusion, Ollama, ...) often need --watch-timeout > 15m.

Examples:
  olares-cli market install firefox                                          # fire-and-forget; returns OperationResult{status:"accepted"}
  olares-cli market install firefox --watch                                  # happy path: block until running
  olares-cli market install firefox --version 1.0.11 --env DEBUG=1 --watch
  olares-cli market install firefox -o json                                  # one accepted-payload JSON doc
  olares-cli market install firefox --watch -o json | jq -r '.finalState'    # scripted success check
  olares-cli market install firefox --watch -q                               # silent + block; exit code = terminal verdict
  olares-cli market install comfyui --compute-mode nvidia --watch            # pin GPU mode (1.12.6+)
  olares-cli market install myapp -s upload --watch                          # install from locally-uploaded chart
  olares-cli market install ollama-webui --watch --watch-timeout 30m         # image-pull-heavy
  olares-cli market install firefox --watch --watch-interval 1s --watch-timeout 5m   # tight CI bounds`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInstall(opts, args[0])
		},
	}
	opts.addCommonFlags(cmd)
	opts.addOutputFlags(cmd)
	opts.addVersionFlag(cmd)
	opts.addEnvFlag(cmd)
	opts.addComputeModeFlag(cmd)
	opts.addWatchFlags(cmd)
	return cmd
}

func runInstall(opts *MarketOptions, appName string) error {
	mc, err := opts.prepare()
	if err != nil {
		return opts.failOp("install", appName, err)
	}

	source := resolveCatalogSource(opts)
	if strings.TrimSpace(opts.Source) == "" {
		opts.info("Using source: %s", source)
	}

	version := strings.TrimSpace(opts.Version)
	if version != "" {
		if err := validateVersion(version); err != nil {
			return opts.failOp("install", appName, err)
		}
	} else {
		v, err := resolveVersionInSource(mc, appName, source)
		if err != nil {
			return opts.failOp("install", appName, fmt.Errorf("cannot determine version in source '%s': %w (use --version to specify)", source, err))
		}
		version = v
		opts.info("Using latest version: %s", version)
	}

	envs, err := parseEnvFlags(opts.Envs)
	if err != nil {
		return opts.failOp("install", appName, err)
	}

	ctx := context.Background()

	// Compute-mode selection is a 1.12.6+ feature. Detect the backend
	// version so we never touch the (untouched) 1.12.5 install path: on
	// 1.12.5 we send no selectedGpuType and never interpret a
	// computeModeSelect 422.
	atLeast126, verr := opts.factory.OlaresBackendAtLeast(ctx, "1.12.6")
	if verr != nil {
		if strings.TrimSpace(opts.ComputeMode) != "" {
			return opts.failOp("install", appName, fmt.Errorf("cannot determine Olares backend version to honor --compute-mode: %w", verr))
		}
		atLeast126 = false
	}
	computeMode := strings.TrimSpace(opts.ComputeMode)
	if computeMode != "" && !atLeast126 {
		return opts.failOp("install", appName, fmt.Errorf("--compute-mode requires Olares 1.12.6+; this backend uses a different (unchanged) install path — re-run without --compute-mode"))
	}

	opts.info("Installing '%s' version '%s' from '%s' for user '%s'...", appName, version, source, mc.olaresID)

	selected := ""
	if atLeast126 {
		selected = computeMode
	}
	resp, err := mc.InstallApp(ctx, appName, version, source, selected, envs)
	// 1.12.6+: recover once from a computeModeSelect 422 by resolving the
	// mode (from --compute-mode, an interactive prompt, or a clear error)
	// and retrying. Skipped entirely on 1.12.5.
	if err != nil && atLeast126 {
		if checkType, raw := parseFailedCheck(resp); isComputeModeSelect(checkType) {
			mode, merr := resolveComputeMode(raw, appName, computeMode, opts.isInteractive())
			if merr != nil {
				return opts.failOp("install", appName, merr)
			}
			opts.info("Selected compute mode: %s", mode)
			resp, err = mc.InstallApp(ctx, appName, version, source, mode, envs)
		}
	}
	if err != nil {
		if envErr := parseServerEnvError(resp, appName); envErr != nil {
			return opts.failOp("install", appName, envErr)
		}
		return opts.failOp("install", appName, err)
	}

	result := newOperationResult(mc, "install", appName, source, version, fmt.Sprintf("install requested for version %s", version), resp)
	return runWithWatch(opts, mc, result, newWatchTarget(watchInstall, appName, source))
}
