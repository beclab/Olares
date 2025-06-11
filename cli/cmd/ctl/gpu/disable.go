package gpu

import (
	"log"

	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/spf13/cobra"
)

func NewCmdDisableGpu() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Disable GPU drivers for Olares node",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.DisableGpuNode(); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	return cmd
}
