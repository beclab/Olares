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
		Short: "add this prepared worker node to a cluster",
		// `node join` is a superset of this command: it takes the same flags and
		// skips the download and prepare steps when the node is already prepared
		// for the cluster's version, so there is nothing left that only `add` can
		// do. It keeps working for existing automation, but is no longer part of
		// the documented surface.
		Hidden:     true,
		Deprecated: "use 'olares-cli node join' instead, which prepares this node first when needed",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.AddNodePipeline(cmd.Context()); err != nil {
				log.Fatal(err)
			}
		},
	}
	flagSetter := config.NewFlagSetterFor(cmd)
	config.AddVersionFlagBy(flagSetter)
	config.AddBaseDirFlagBy(flagSetter)
	config.AddCDNServiceFlagBy(flagSetter)
	config.AddMasterHostFlagsBy(flagSetter)

	return cmd
}
