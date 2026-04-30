package wizard

import "github.com/spf13/cobra"

func NewWizardCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wizard",
		Short: "activation wizard operations",
	}
	cmd.AddCommand(NewCmdActivate())
	return cmd
}
