package node

import (
	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/pipelines"
	"github.com/spf13/cobra"
	"log"
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
