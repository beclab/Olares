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
  list          every model configured, across every provider
  models        every name the model field accepts, as a caller sees it
  route         the names callers may send instead of a provider and model
  default       the categories a caller can ask for instead of a model
  provider      the upstreams Router routes to, and the models they serve
  spec          what a local model declares itself to be, and changing it
  local         the Model Console inside a model application on this machine
  call          send work to a model: text, embeddings, web, images, audio, OCR
  key           API keys for software that calls Router
  quota         ceilings on a key, a person, a model, or an application
  usage         what has been called, what it cost, and how long it is kept
  audit         who changed Router, and to what

Model applications are installed, cloned, upgraded and removed with
"olares-cli market"; the people on this Olares are "olares-cli settings users".

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
	cmd.AddCommand(NewListCommand(f))
	cmd.AddCommand(newModelsCommand(f))
	cmd.AddCommand(NewRouteCommand(f))
	cmd.AddCommand(NewDefaultCommand(f))
	cmd.AddCommand(NewProviderCommand(f))
	cmd.AddCommand(NewSpecCommand(f))
	cmd.AddCommand(NewLocalCommand(f))
	cmd.AddCommand(NewCallCommand(f))
	cmd.AddCommand(NewKeyCommand(f))
	cmd.AddCommand(NewQuotaCommand(f))
	cmd.AddCommand(NewUsageCommand(f))
	cmd.AddCommand(NewAuditCommand(f))
	return cmd
}
