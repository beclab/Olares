package amdgpu

import (
	"log"

	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/spf13/cobra"
)

func NewCmdAmdGpuUninstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall AMD ROCm stack via amdgpu-install",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.AmdGpuUninstall(); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	return cmd
}
