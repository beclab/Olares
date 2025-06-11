package node

import (
	"log"

	"github.com/beclab/Olares/cli/cmd/ctl/options"
	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/spf13/cobra"
)

func NewCmdAddNode() *cobra.Command {
	o := options.NewAddNodeOptions()
	cmd := &cobra.Command{
		Use:   "add",
		Short: "add worker node to the cluster",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.AddNodePipeline(o); err != nil {
				log.Fatal(err)
			}
		},
	}
	o.AddFlags(cmd)
	return cmd
}
