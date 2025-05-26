package node

import "github.com/spf13/cobra"

func NewNodeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node",
		Short: "cluster node related operations",
	}
	cmd.AddCommand(NewCmdMasterInfo())
	cmd.AddCommand(NewCmdAddNode())
	return cmd
}
