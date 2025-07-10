package os

import (
	"encoding/json"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/cmd/ctl/options"
	"github.com/beclab/Olares/cli/pkg/phase"
	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/beclab/Olares/cli/pkg/upgrade"
	"github.com/beclab/Olares/cli/version"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"log"
	"os"
)

type UpgradeOsOptions struct {
	UpgradeOptions *options.UpgradeOptions
}

func NewUpgradeOsOptions() *UpgradeOsOptions {
	return &UpgradeOsOptions{
		UpgradeOptions: options.NewUpgradeOptions(),
	}
}

func NewCmdUpgradeOs() *cobra.Command {
	o := NewUpgradeOsOptions()
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade Olares to a newer version",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.UpgradeOlaresPipeline(o.UpgradeOptions); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	o.UpgradeOptions.AddFlags(cmd)
	cmd.AddCommand(NewCmdUpgradePrecheck())
	cmd.AddCommand(NewCmdGetUpgradePath())
	return cmd
}

func NewCmdGetUpgradePath() *cobra.Command {
	var baseVersionStr, targetVersionStr string
	cmd := &cobra.Command{
		Use:   "path",
		Short: "Get the upgrade path (required intermediate versions) from base version to target version",
		RunE: func(cmd *cobra.Command, args []string) error {
			var baseVersion, targetVersion *semver.Version
			var err error
			if baseVersionStr == "" {
				baseVersionStr, err = phase.GetOlaresVersion()
				if err != nil {
					return errors.New("failed to get current Olares version, please specify the base version explicitly")
				}
			}
			baseVersion, err = semver.NewVersion(baseVersionStr)
			if err != nil {
				return fmt.Errorf("invalid base version: %v", err)
			}
			cliVersion, err := semver.NewVersion(version.VERSION)
			if err != nil {
				fmt.Printf("invalid olares-cli version \"%s\" for upgrade: %v\n", version.VERSION, err)
				os.Exit(0)
			}

			if targetVersionStr == "" {
				targetVersion = cliVersion
			} else {
				targetVersion, err = semver.NewVersion(targetVersionStr)
				if err != nil {
					return fmt.Errorf("invalid target version: %v", err)
				}
				if targetVersion.GreaterThan(cliVersion) {
					fmt.Printf("target version (%s) is greater than olares-cli version (%s), unable to upgrade, please upgrade olares-cli first", targetVersion, cliVersion)
					os.Exit(0)
				}
			}
			if baseVersion.GreaterThanEqual(targetVersion) {
				return fmt.Errorf("base version %s is no less than target version %s, no need to upgrade", baseVersionStr, targetVersionStr)
			}
			if targetVersion.GreaterThan(cliVersion) {
				fmt.Printf("target version (%s) is greater than olares-cli version (%s), unable to upgrade, please upgrade olares-cli first!\n", targetVersion, cliVersion)
				os.Exit(0)
			}
			path := upgrade.GetUpgradePathFor(baseVersion, targetVersion)
			encoder := json.NewEncoder(cmd.OutOrStdout())
			encoder.SetIndent("", "  ")
			return encoder.Encode(path)
		},
	}

	cmd.Flags().StringVarP(&baseVersionStr, "base-version", "b", baseVersionStr, "base version to be upgraded, defaults to the current Olares version if inside Olares cluster")
	cmd.Flags().StringVarP(&targetVersionStr, "target-version", "t", targetVersionStr, fmt.Sprintf("target version to upgrade to, defaults to the latest Olares version when this olares-cli was released: %s", version.VERSION))

	return cmd
}

func NewCmdUpgradePrecheck() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "precheck",
		Short: "Precheck Olares for Upgrade",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.UpgradePreCheckPipeline(); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	return cmd
}
