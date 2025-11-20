package disk

import "github.com/spf13/cobra"

func NewDiskCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disk",
		Short: "disk management operations",
	}

	cmd.AddCommand(NewListUnmountedDisksCommand())
	cmd.AddCommand(NewExtendDiskCommand())

	return cmd
}
