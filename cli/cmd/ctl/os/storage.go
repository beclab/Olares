package os

import (
	"log"

	"github.com/beclab/Olares/cli/cmd/config"
	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/spf13/cobra"
)

func NewCmdInstallStorage() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "storage",
		Short: "install a storage backend for the Olares shared filesystem, or in the case of external storage, validate the config",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.CliInstallStoragePipeline(); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	flagSetter := config.NewFlagSetterFor(cmd)

	config.AddVersionFlagBy(flagSetter)
	config.AddBaseDirFlagBy(flagSetter)
	config.AddStorageFlagsBy(flagSetter)

	return cmd
}
