package os

import (
	"log"

	"github.com/beclab/Olares/cli/cmd/config"
	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/spf13/cobra"
)

func NewCmdPrecheck() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "precheck",
		Short: "precheck the installation compatibility of the system",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.StartPreCheckPipeline(); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	flagSetter := config.NewFlagSetterFor(cmd)
	config.AddVersionFlagBy(flagSetter)
	config.AddBaseDirFlagBy(flagSetter)
	return cmd
}
