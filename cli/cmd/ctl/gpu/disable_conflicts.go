package gpu

import (
	"log"

	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/spf13/cobra"
)

func NewCmdDisableConflicts() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "disable-conflicts",
		Aliases: []string{"disable-nouveau"},
		Short:   "Blacklist and disable the in-tree GPU drivers (nouveau, nova) that conflict with the NVIDIA driver",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.DisableConflictingGPUDrivers(); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	return cmd
}
