package nfd

import (
	"log"

	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/spf13/cobra"
)

func NewCmdInstallNfd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install Node Feature Discovery",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.InstallNFD(); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
}
