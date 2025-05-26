package node

import (
	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/pipelines"
	"github.com/spf13/cobra"
	"log"
)

func NewCmdMasterInfo() *cobra.Command {
	o := options.NewMasterInfoOptions()
	cmd := &cobra.Command{
		Use:   "masterinfo",
		Short: "get information about master node, and check whether current node can be added to the cluster",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.MasterInfoPipeline(o); err != nil {
				log.Fatal(err)
			}
		},
	}
	o.AddFlags(cmd)
	return cmd
}
