package gpu

import (
	"log"

	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/spf13/cobra"
)

func NewCmdDisableNouveau() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable-nouveau",
		Short: "Blacklist and disable the nouveau kernel module",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.DisableNouveau(); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	return cmd
}


