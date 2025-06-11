package gpu

import (
	"log"

	"github.com/beclab/Olares/cli/cmd/ctl/options"
	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/spf13/cobra"
)

func NewCmdUpgradeGpu() *cobra.Command {
	o := options.NewInstallGpuOptions()
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "upgrade GPU drivers for Olares",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.UpgradeGpuDrivers(o); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	o.AddFlags(cmd)
	return cmd
}
