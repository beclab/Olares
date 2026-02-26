package amdgpu

import (
	"log"

	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/spf13/cobra"
)

func NewCmdAmdGpuInstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install AMD ROCm stack via amdgpu-install",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.AmdGpuInstall(); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	return cmd
}
