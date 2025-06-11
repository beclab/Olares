package node

import (
	"log"

	"github.com/beclab/Olares/cli/cmd/ctl/options"
	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/spf13/cobra"
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
