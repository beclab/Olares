package gpu

import (
	"log"

	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/spf13/cobra"
)

func NewCmdUninstallpu() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "uninstall GPU drivers for Olares",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.UninstallGpuDrivers(); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	return cmd
}
