package gpu

import (
	"github.com/spf13/cobra"
)

func NewCmdGpu() *cobra.Command {
	rootGpuCmd := &cobra.Command{
		Use:   "gpu",
		Short: "install / uninstall / status / disbale gpu commands for olares.",
	}

	rootGpuCmd.AddCommand(NewCmdInstallGpu())
	rootGpuCmd.AddCommand(NewCmdUninstallpu())
	rootGpuCmd.AddCommand(NewCmdEnableGpu())
	rootGpuCmd.AddCommand(NewCmdDisableGpu())
	rootGpuCmd.AddCommand(NewCmdGpuStatus())
	return rootGpuCmd
}
