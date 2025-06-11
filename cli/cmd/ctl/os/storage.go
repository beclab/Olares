package os

import (
	"log"

	"github.com/beclab/Olares/cli/cmd/ctl/options"
	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/spf13/cobra"
)

func NewCmdInstallStorage() *cobra.Command {
	o := options.NewInstallStorageOptions()
	cmd := &cobra.Command{
		Use:   "storage",
		Short: "install a storage backend for the Olares shared filesystem, or in the case of external storage, validate the config",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.CliInstallStoragePipeline(o); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	o.AddFlags(cmd)

	return cmd
}
