package node

import (
	"log"

	"github.com/beclab/Olares/cli/cmd/config"
	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/spf13/cobra"
)

func NewCmdAddNode() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "add worker node to the cluster",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.AddNodePipeline(); err != nil {
				log.Fatal(err)
			}
		},
	}
	flagSetter := config.NewFlagSetterFor(cmd)
	config.AddVersionFlagBy(flagSetter)
	config.AddBaseDirFlagBy(flagSetter)
	config.AddMasterHostFlagsBy(flagSetter)

	return cmd
}
