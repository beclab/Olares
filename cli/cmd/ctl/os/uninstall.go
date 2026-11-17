package os

import (
	"log"

	"github.com/beclab/Olares/cli/cmd/config"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/phase/cluster"
	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewCmdUninstallOs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall Olares",
		Run: func(cmd *cobra.Command, args []string) {
			var removeSelf func() error
			if viper.GetBool(common.FlagUninstallAll) {
				var err error
				removeSelf, err = prepareSelfRemoval()
				if err != nil {
					log.Fatalf("cannot prepare olares-cli removal: %v", err)
				}
			}
			err := pipelines.UninstallTerminusPipeline(cmd.Context())
			if err != nil {
				log.Fatalf("error: %v", err)
			}
			if removeSelf != nil {
				if err := removeSelf(); err != nil {
					log.Fatalf("Olares was uninstalled, but could not remove olares-cli: %v", err)
				}
			}
		},
	}

	flagSetter := config.NewFlagSetterFor(cmd)

	config.AddVersionFlagBy(flagSetter)
	config.AddBaseDirFlagBy(flagSetter)
	config.AddStorageFlagsBy(flagSetter)

	// these two flags' names are too general, and only used in cmd options, so we manually bind them to the viper
	// inside the pipeline creator, it still uses the flag vars to get the values
	cmd.Flags().Bool("all", false, "Uninstall Olares completely, including prepared dependencies and the CLI itself on Linux/macOS")
	viper.BindPFlag(common.FlagUninstallAll, cmd.Flags().Lookup("all"))
	cmd.Flags().String("phase", cluster.PhaseInstall.String(), "Uninstall from a specified phase and revert to the previous one. For example, using --phase install will remove the tasks performed in the 'install' phase, effectively returning the system to the 'prepare' state.")
	viper.BindPFlag(common.FlagUninstallPhase, cmd.Flags().Lookup("phase"))
	return cmd
}
