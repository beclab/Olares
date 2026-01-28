package gpu

import (
	"log"

	"github.com/beclab/Olares/cli/cmd/config"
	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/spf13/cobra"
)

func NewCmdInstallGpu() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install GPU drivers for Olares",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.InstallGpuDrivers(); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	flagSetter := config.NewFlagSetterFor(cmd)
	config.AddVersionFlagBy(flagSetter)
	config.AddBaseDirFlagBy(flagSetter)
	return cmd
}
