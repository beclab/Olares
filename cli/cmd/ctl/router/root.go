package router

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewRouterCommand assembles the `olares-cli router` subtree: Router's
// management surface plus the Model Console runtime behind locally installed
// models. Identity and transport come from the active profile, as in the
// market / files / settings trees.
func NewRouterCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "router",
		Short: "AI models through Router and the Model Console",
		Long: `Configure and operate the AI models this Olares can reach.

Router is the gateway every model call goes through, whether the model runs
on this machine or at a cloud provider. It is a Market application, so this
tree locates it at runtime instead of assuming a hostname — run "router
status" first if anything here behaves unexpectedly.

  status        where Router lives, whether it is healthy, and your role
  whoami        the identity and role Router sees for the active profile
  list          every model configured, across every provider
  capabilities  the capability flags a model row can declare
  default       which model answers when a request names none
  provider      the upstreams Router routes to, and the models they serve
  app           model applications that run models on this machine
  local         the Model Console inside one of those applications
  call          send work to a model: chat, embed, transcribe, speak, OCR
  key           API keys for software that calls Router
  quota         spend and rate ceilings on a key, a person, or a model
  caller        the applications that call Router
  usage         what has been called, and what it cost
  audit         who changed Router, and to what
  trace         the spans an agent framework reported for a call
  user          the people Router knows

Most of Router's management surface is admin-only. Requires Olares 1.12.7+.

Run "olares-cli router <verb> --help" for details.
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
	cmd.AddCommand(NewLocalCommand(f))
	cmd.AddCommand(NewCallCommand(f))
	cmd.AddCommand(NewKeyCommand(f))
	cmd.AddCommand(NewQuotaCommand(f))
	cmd.AddCommand(NewCallerCommand(f))
	cmd.AddCommand(NewUsageCommand(f))
	cmd.AddCommand(NewAuditCommand(f))
	cmd.AddCommand(NewTraceCommand(f))
	cmd.AddCommand(NewUserCommand(f))
	return cmd
}
