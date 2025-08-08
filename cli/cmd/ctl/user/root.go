package user

import "github.com/spf13/cobra"

func NewUserCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "user management operations",
	}
	cmd.AddCommand(NewCmdCreateUser())
	cmd.AddCommand(NewCmdDeleteUser())
	cmd.AddCommand(NewCmdListUsers())
	cmd.AddCommand(NewCmdGetUser())
	// cmd.AddCommand(NewCmdUpdateUserLimits())
	return cmd
}
