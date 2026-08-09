// Package dev implements `olares-cli dev ...` — the loop for testing a
// locally built image on a running Olares.
//
// Registration note: this group is added to the root command
// UNCONDITIONALLY, alongside `cluster` / `market` / `chart`, and not
// inside root.go's `!remoteOnly` guard. The guard marks verbs that need
// an Olares host filesystem; putting `dev` behind it would make the
// whole group a host verb, and host verbs only reach users when the OS
// itself is upgraded (the host binary is cut by the OS release with
// publish-npm: false, then replaced by `olares-cli upgrade`). Outside
// the guard, `dev` ships on the CLI's own npm cadence.
//
// The visible cost is that an `npx @olares/cli` user can see
// `--transport local`, which can never work there. That is handled by
// failing the transport with an explanatory message rather than by
// hiding the group — a wrong-looking flag is cheaper than a verb that
// is missing for months on the machines that most want it.
package dev

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewDevCommand assembles `olares-cli dev`.
func NewDevCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dev",
		Short: "Test locally built images on a running Olares",
		Long: `Push a locally built container image onto your Olares and point a
workload at it, then put everything back when you're done.

The loop:

  olares-cli dev push  beclab/app-service:dev
  olares-cli dev deploy beclab/app-service:dev --replaces beclab/app-service:0.6.22 -w
  olares-cli dev status
  olares-cli dev revert os-framework/app-service --kind sts

push moves bytes; deploy repoints workloads. They are separate verbs
because the failure modes are unrelated — a failed push is a transport
or permission problem on the node, a failed deploy is an API or RBAC
problem — and because one push often feeds several deploys.

Transports (--transport, default auto):

  local     import straight into this machine's containerd; only
            works when the CLI runs on the Olares node itself
  ssh       stream the image to the node over SSH (configure it once
            with ` + "`olares-cli dev node set`" + `)
  registry  not implemented yet
  api       not implemented yet

auto tries local, then ssh, and explains what it would take to make
either work if neither is available.

Repointing a workload always goes through the same authenticated
ControlHub path as ` + "`cluster workload set-image`" + `, whether the CLI
runs on the node or on your laptop. Only the image transport differs
by location.
`,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("unknown command %q for %q", args[0], c.CommandPath())
			}
			return c.Help()
		},
	}
	cmd.SilenceUsage = true
	cmd.PersistentPreRun = func(c *cobra.Command, args []string) {
		c.SilenceUsage = true
	}

	cmd.AddCommand(NewComponentsCommand())
	cmd.AddCommand(NewNodeCommand())
	cmd.AddCommand(NewPushCommand(f))
	cmd.AddCommand(NewDeployCommand(f))
	cmd.AddCommand(NewRevertCommand(f))
	cmd.AddCommand(NewStatusCommand(f))

	return cmd
}
