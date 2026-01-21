package amdgpu

import (
	"log"

	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/spf13/cobra"
)

func NewCmdAmdGpuStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show AMD GPU driver and ROCm status",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.AmdGpuStatus(); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	return cmd
}
