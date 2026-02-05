package amdgpu

import "github.com/spf13/cobra"

func NewCmdAmdGpu() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "amdgpu",
		Short: "Manage AMD GPU ROCm stack",
	}
	cmd.AddCommand(NewCmdAmdGpuInstall())
	cmd.AddCommand(NewCmdAmdGpuUninstall())
	cmd.AddCommand(NewCmdAmdGpuStatus())
	return cmd
}


