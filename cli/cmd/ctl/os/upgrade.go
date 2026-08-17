package os

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/cmd/config"
	"github.com/beclab/Olares/cli/pkg/phase"
	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/beclab/Olares/cli/pkg/upgrade"
	"github.com/beclab/Olares/cli/pkg/utils"
	"github.com/beclab/Olares/cli/version"
	"github.com/spf13/cobra"
)

func NewCmdUpgradeOs() *cobra.Command {
	var stage string

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade Olares to a newer version",
		// Without this, a word this command does not recognise is silently
		// ignored and the whole upgrade runs. That turns a subcommand this
		// binary is too old to have — `upgrade plan`, which the orchestrator
		// asks every node for — from a clean failure into an unrequested
		// upgrade of the machine that was only being asked a question.
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			if stage == "" {
				err = pipelines.UpgradeOlaresPipeline(cmd.Context())
			} else {
				err = pipelines.UpgradeOlaresStagePipeline(cmd.Context(), stage)
			}
			if err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}

	flagSetter := config.NewFlagSetterFor(cmd)
	config.AddBaseDirFlagBy(flagSetter)
	config.AddVersionFlagBy(flagSetter)

	// This is how a cluster orchestrator drives one node. Left unset, the
	// command upgrades everything on this machine, which is both the
	// single-node path and what it did before clusters were scheduled.
	cmd.Flags().StringVar(&stage, "stage", "",
		"run only this stage of the upgrade flow on this node (see 'olares-cli upgrade plan')")

	cmd.AddCommand(NewCmdCurrentVersionUpgradeSpec())
	cmd.AddCommand(NewCmdUpgradeViable())
	cmd.AddCommand(NewCmdUpgradePrecheck())
	cmd.AddCommand(NewCmdUpgradePlan())
	return cmd
}

// NewCmdUpgradePlan prints the upgrade flow and what this version puts in
// each stage. The orchestrator reads it rather than holding a plan of its own:
// which tasks exist, and which stage each belongs to, is decided by the code
// that implements them, and that code ships in this binary.
func NewCmdUpgradePlan() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan",
		Short: fmt.Sprintf("Print the upgrade flow and its tasks for this olares-cli version (%s)", version.VERSION),
		RunE: func(cmd *cobra.Command, args []string) error {
			target, err := utils.ParseOlaresVersionString(version.VERSION)
			if err != nil {
				return fmt.Errorf("invalid olares-cli version '%s': %v", version.VERSION, err)
			}
			plan, err := upgrade.BuildPlan(target)
			if err != nil {
				return err
			}
			out, err := json.MarshalIndent(plan, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(out))
			return nil
		},
	}
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
			if err := pipelines.UpgradePreCheckPipeline(cmd.Context()); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	return cmd
}
