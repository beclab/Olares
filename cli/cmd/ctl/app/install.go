package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func NewCmdAppInstall() *cobra.Command {
	opts := &AppOptions{Output: "table"}
	cmd := &cobra.Command{
		Use:   "install {app-name}",
		Short: "Install an app",
		Long: `Install an application from a market source.

If --version is not specified, the latest available version is used.
For apps that declare environment variables, use --env to provide values.

Examples:
  olares-cli app install firefox
  olares-cli app install myapp --version 1.0.0 -s market.olares
  olares-cli app install myapp --env API_KEY=abc123 --env REGION=us-east`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInstall(opts, args[0])
		},
	}
	opts.addCommonFlags(cmd)
	opts.addOutputFlags(cmd)
	opts.addVersionFlag(cmd)
	opts.addEnvFlag(cmd)
	return cmd
}

func runInstall(opts *AppOptions, appName string) error {
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

	opts.info("Installing '%s' version '%s' from '%s' for user '%s'...", appName, version, source, mc.user)

	ctx := context.Background()
	resp, err := mc.InstallApp(ctx, appName, version, source, envs)
	if err != nil {
		if envErr := parseServerEnvError(resp, appName); envErr != nil {
			return opts.failOp("install", appName, envErr)
		}
		return opts.failOp("install", appName, err)
	}

	result := newOperationResult(mc, "install", appName, source, version, fmt.Sprintf("install requested for version %s", version), resp)
	return finishOperation(opts, mc, result)
}
