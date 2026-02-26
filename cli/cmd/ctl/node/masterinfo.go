package node

import (
	"log"

	"github.com/beclab/Olares/cli/cmd/config"
	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/spf13/cobra"
)

func NewCmdMasterInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "masterinfo",
		Short: "get information about master node, and check whether current node can be added to the cluster",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.MasterInfoPipeline(); err != nil {
				log.Fatal(err)
			}
		},
	}
	flagSetter := config.NewFlagSetterFor(cmd)
	config.AddBaseDirFlagBy(flagSetter)
	config.AddMasterHostFlagsBy(flagSetter)

	return cmd
}
