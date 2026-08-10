package model

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewModelCommand assembles the `olares-cli model` subtree: Router's
// management surface plus the Model Console runtime behind locally installed
// models. Identity and transport come from the active profile, as in the
// market / files / settings trees.
func NewModelCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "model",
		Short: "AI models through Router and the Model Console",
		Long: `Configure and operate the AI models this Olares can reach.

Router is the gateway every model call goes through, whether the model runs
on this machine or at a cloud provider. It is a Market application, so this
tree locates it at runtime instead of assuming a hostname — run "model
status" first if anything here behaves unexpectedly.

  status        where Router lives, whether it is healthy, and your role
  whoami        the identity and role Router sees for the active profile
  list          every model configured, across every provider
  capabilities  the capability flags a model row can declare
  default       which model answers when a request names none
  provider      the upstreams Router routes to, and the models they serve
  app           model applications that run models on this machine

Most of Router's management surface is admin-only. Requires Olares 1.12.6+.

Run "olares-cli model <verb> --help" for details.
`,
	}
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.PersistentPreRun = func(c *cobra.Command, args []string) {
		c.SilenceErrors = true
		c.SilenceUsage = true
	}

	cmd.AddCommand(NewStatusCommand(f))
	cmd.AddCommand(NewWhoamiCommand(f))
	cmd.AddCommand(NewListCommand(f))
	cmd.AddCommand(NewCapabilitiesCommand(f))
	cmd.AddCommand(NewDefaultCommand(f))
	cmd.AddCommand(NewProviderCommand(f))
	cmd.AddCommand(NewAppCommand(f))
	return cmd
}
