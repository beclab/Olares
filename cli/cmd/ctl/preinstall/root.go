package preinstall

import (
	"github.com/spf13/cobra"
)

// NewPreinstallCommand assembles the `olares-cli preinstall` subtree.
// Verbs here are local developer utilities for offline Market bootstrap
// bundles; they do not talk to a running Olares and do not require a
// profile login.
func NewPreinstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "preinstall",
		Short: "Offline Market preinstall bundle utilities",
		Long: `Developer-side helpers for validating offline Market preinstall
bundles (the static preinstall/market directory shipped with installer
media).

These verbs are local-only — they do not talk to a running Olares cluster
and do not require a profile login.`,
	}
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.PersistentPreRun = func(c *cobra.Command, args []string) {
		c.SilenceErrors = true
		c.SilenceUsage = true
	}
	cmd.AddCommand(NewCmdPreinstallCheck())
	return cmd
}
