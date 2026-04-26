package market

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func NewCmdMarketUpgrade(f *cmdutil.Factory) *cobra.Command {
	opts := newMarketOptions(f)
	cmd := &cobra.Command{
		Use:   "upgrade {app-name}",
		Short: "Upgrade an installed app",
		Long: `Upgrade an installed application to a new version.

If --version is not specified, the latest available version is used.

Examples:
  olares-cli market upgrade myapp
  olares-cli market upgrade myapp --version 2.0.0
  olares-cli market upgrade myapp --env API_KEY=new-key`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpgrade(opts, args[0])
		},
	}
	opts.addCommonFlags(cmd)
	opts.addOutputFlags(cmd)
	opts.addVersionFlag(cmd)
	opts.addEnvFlag(cmd)
	opts.addWatchFlags(cmd)
	return cmd
}

func runUpgrade(opts *MarketOptions, appName string) error {
	mc, err := opts.prepare()
	if err != nil {
		return opts.failOp("upgrade", appName, err)
	}

	source := resolveCatalogSource(opts)
	if strings.TrimSpace(opts.Source) == "" {
		opts.info("Using source: %s", source)
	}

	version := strings.TrimSpace(opts.Version)
	if version != "" {
		if err := validateVersion(version); err != nil {
			return opts.failOp("upgrade", appName, err)
		}
	} else {
		v, err := resolveVersionInSource(mc, appName, source)
		if err != nil {
			return opts.failOp("upgrade", appName, fmt.Errorf("cannot determine version in source '%s': %w (use --version to specify)", source, err))
		}
		version = v
		opts.info("Using latest version: %s", version)
	}

	envs, err := parseEnvFlags(opts.Envs)
	if err != nil {
		return opts.failOp("upgrade", appName, err)
	}

	opts.info("Upgrading '%s' to version '%s' from '%s' for user '%s'...", appName, version, source, mc.olaresID)

	ctx := context.Background()
	resp, err := mc.UpgradeApp(ctx, appName, version, source, envs)
	if err != nil {
		if envErr := parseServerEnvError(resp, appName); envErr != nil {
			return opts.failOp("upgrade", appName, envErr)
		}
		return opts.failOp("upgrade", appName, err)
	}

	result := newOperationResult(mc, "upgrade", appName, source, version, fmt.Sprintf("upgrade requested for version %s", version), resp)
	return runWithWatch(opts, mc, result, newWatchTarget(watchUpgrade, appName, source))
}
