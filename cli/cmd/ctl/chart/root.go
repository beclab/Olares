package chart

import "github.com/spf13/cobra"

// NewChartCommand assembles the `olares-cli chart` subtree. Verbs here are
// developer-side helpers for working with the OlaresManifest.yaml + Helm
// chart layout the app store ingests; they don't talk to a running Olares
// and don't go through the profile-based HTTP API.
func NewChartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chart",
		Short: "Olares chart developer utilities",
		Long: `Developer-side helpers for working with Olares chart packages
(the same OlaresManifest.yaml + Helm chart layout the app store ingests).

These verbs are local-only — they do not talk to a running Olares cluster
and do not require a profile login.`,
	}
	// Lint failures are domain errors, not usage errors — don't dump the
	// full help block on every validation failure.
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.PersistentPreRun = func(c *cobra.Command, args []string) {
		c.SilenceErrors = true
		c.SilenceUsage = true
	}
	cmd.AddCommand(NewCmdChartLint())
	return cmd
}
