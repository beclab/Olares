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
	cmd.AddCommand(NewCmdCurrentVersionUpgradeSpec())
	cmd.AddCommand(NewCmdUpgradeViable())
	cmd.AddCommand(NewCmdUpgradePrecheck())
	return cmd
}

func NewCmdCurrentVersionUpgradeSpec() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "spec",
		Aliases: []string{"current-spec"},
		Short:   fmt.Sprintf("Get the upgrade spec of the current olares-cli version (%s)", version.VERSION),
		RunE: func(cmd *cobra.Command, args []string) error {
			spec, err := upgrade.CurrentVersionSpec()
			if err != nil {
				return err
			}
			jsonOutput, _ := json.MarshalIndent(spec, "", "  ")
			fmt.Println(string(jsonOutput))
			return nil
		},
	}
	return cmd
}

func NewCmdUpgradeViable() *cobra.Command {
	var baseVersionStr string
	cmd := &cobra.Command{
		Use:   "viable",
		Short: fmt.Sprintf("Determine whether upgrade can be directly performed upon the base version (to %s)", version.VERSION),
		RunE: func(cmd *cobra.Command, args []string) error {
			if baseVersionStr == "" {
				var err error
				baseVersionStr, err = phase.GetOlaresVersion()
				if err != nil {
					return err
				}
			}
			baseVersion, err := semver.NewVersion(baseVersionStr)
			if err != nil {
				return fmt.Errorf("invalid base version '%s': %v", baseVersionStr, err)
			}
			cliVersion, err := semver.NewVersion(version.VERSION)
			if err != nil {
				return fmt.Errorf("invalid cli version '%s': %v", version.VERSION, err)
			}
			err = upgrade.Check(baseVersion, cliVersion)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Printf("upgrade from %s to %s is viable\n", baseVersion, cliVersion)
			return nil
		},
	}
	cmd.Flags().StringVarP(&baseVersionStr, "base", "b", "", "base version, defaults to the current Olares system's version")
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
