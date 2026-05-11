package app

import "github.com/spf13/cobra"

func NewAppCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "app",
		Short: "application management operations",
		Long:  "Manage applications via the Olares Market API. Supports listing, inspecting, installing, uninstalling, upgrading, uploading charts, and controlling app lifecycle.",
	}
	cmd.SilenceUsage = true
	cmd.PersistentPreRun = func(c *cobra.Command, args []string) {
		c.SilenceUsage = true
	}

	cmd.AddCommand(NewCmdAppList())
	cmd.AddCommand(NewCmdAppCategories())
	cmd.AddCommand(NewCmdAppGet())
	cmd.AddCommand(NewCmdAppInstall())
	cmd.AddCommand(NewCmdAppUninstall())
	cmd.AddCommand(NewCmdAppUpgrade())
	cmd.AddCommand(NewCmdAppClone())
	cmd.AddCommand(NewCmdAppCancel())
	cmd.AddCommand(NewCmdAppStop())
	cmd.AddCommand(NewCmdAppResume())
	cmd.AddCommand(NewCmdAppUpload())
	cmd.AddCommand(NewCmdAppDelete())
	cmd.AddCommand(NewCmdAppSync())
	cmd.AddCommand(NewCmdAppStatus())
	return cmd
}
