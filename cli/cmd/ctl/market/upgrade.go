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
		Short: "Upgrade an installed app to a newer version (PUT /apps/{name}/upgrade)",
		Long: `Upgrade an installed application to a new version.

If --version is omitted, the latest catalog version for the resolved
source is used. Source resolution mirrors 'install': -s pins the
source; omitting it falls back to the auto-selected source.

Pre-flight (mirror of SPA's canUpgrade() in constant/config.ts):
runs BEFORE the PUT request, bails locally with a self-contained
error (formatted via failOp so -o json carries it in 'message' and -q
still surfaces the exit code). Four gates:

  1. Row exists                   — state row found via Name or RawName
                                    (clones included).
  2. State is upgradable          — running / stopped / stopFailed /
                                    upgradeFailed / applyEnvFailed
                                    (verbatim isUpgradableAppStates).
  3. Strict semver newer          — target > installed via
                                    Masterminds/semver/v3 strict parse;
                                    rejects downgrade and same-version
                                    no-ops.
  4. Not suspended / withdrawn    — catalog labels exclude both
                                    'suspend' and 'remove' (verbatim
                                    suspendApp predicate). Soft-fails
                                    on probe errors so flaky catalog
                                    reads don't block valid upgrades.

Env vars are intentionally NOT accepted here: this mirrors the Market
SPA's upgrade dialog, which sends only {app_name, source, version} and
preserves any existing env values server-side from the prior install.
Use 'olares-cli market env --set KEY=value <app>' (out-of-band) to
change env values — that's the same flow the SPA exposes via its
env-editor dialog. The CLI's UpgradeRequest wire payload has no
'envs' field, so passing envs accidentally is impossible.

Examples:
  olares-cli market upgrade firefox                                          # to catalog latest
  olares-cli market upgrade firefox --version 2.0.0
  olares-cli market upgrade firefox --watch                                  # block until row settles
  olares-cli market upgrade firefox --version 2.0.0 --watch -o json | jq -r '.finalState'
  olares-cli market upgrade firefox --version 2.0.0 --watch --watch-timeout 30m     # slow image pull
  olares-cli market upgrade firefox --watch -q                               # silent; exit code = verdict
  olares-cli market upgrade firefox --watch --watch-interval 1s --watch-timeout 10m`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpgrade(opts, args[0])
		},
	}
	opts.addCommonFlags(cmd)
	opts.addOutputFlags(cmd)
	opts.addVersionFlag(cmd)
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

	ctx := context.Background()
	// Pre-flight gate mirroring the SPA's canUpgrade(): refuse early
	// (with an actionable message) when the state row is missing, in a
	// non-upgradable state, when the target version is not strictly
	// newer than the installed version, or when the catalog row is
	// marked suspend / remove. Soft-fails on transient catalog probe
	// errors — see preflightUpgrade for rationale.
	if err := preflightUpgrade(ctx, opts, mc, appName, version, source); err != nil {
		return opts.failOp("upgrade", appName, err)
	}

	opts.info("Upgrading '%s' to version '%s' from '%s' for user '%s'...", appName, version, source, mc.olaresID)

	// Upgrade can never produce an `appenv` 422 because we don't ship
	// envs in the payload (parity with the SPA's upgradeApp). Failures
	// here are surfaced verbatim — no parseServerEnvError branch.
	resp, err := mc.UpgradeApp(ctx, appName, version, source)
	if err != nil {
		return opts.failOp("upgrade", appName, err)
	}

	result := newOperationResult(mc, "upgrade", appName, source, version, fmt.Sprintf("upgrade requested for version %s", version), resp)
	return runWithWatch(opts, mc, result, newWatchTarget(watchUpgrade, appName, source))
}
