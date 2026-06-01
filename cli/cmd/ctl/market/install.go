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

	opts.info("Installing '%s' version '%s' from '%s' for user '%s'...", appName, version, source, mc.olaresID)

	ctx := context.Background()
	resp, err := mc.InstallApp(ctx, appName, version, source, envs)
	if err != nil {
		if envErr := parseServerEnvError(resp, appName); envErr != nil {
			return opts.failOp("install", appName, envErr)
		}
		return opts.failOp("install", appName, err)
	}

	result := newOperationResult(mc, "install", appName, source, version, fmt.Sprintf("install requested for version %s", version), resp)
	return runWithWatch(opts, mc, result, newWatchTarget(watchInstall, appName, source))
}
