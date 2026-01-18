package nfd

import "github.com/spf13/cobra"

func NewNfdCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nfd",
		Short: "Node Feature Discovery operations",
	}
	cmd.AddCommand(NewCmdInstallNfd())
	return cmd
}
