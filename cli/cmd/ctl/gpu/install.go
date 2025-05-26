package gpu

import (
	"log"

	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/pipelines"
	"github.com/spf13/cobra"
)

func NewCmdInstallGpu() *cobra.Command {
	o := options.NewInstallGpuOptions()
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install GPU drivers for Olares",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.InstallGpuDrivers(o); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	o.AddFlags(cmd)
	return cmd
}
