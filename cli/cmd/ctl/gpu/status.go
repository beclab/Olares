package gpu

import (
	"log"

	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/spf13/cobra"
)

func NewCmdGpuStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Print GPU drivers status",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.GpuDriverStatus(); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	return cmd
}
