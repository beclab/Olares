// Package password implements `olares-cli cluster middleware password ...`.
//
// Sub-noun under `cluster middleware` so the password-management verbs
// (today: `set`; future: maybe `rotate`, `reveal`) cluster together
// rather than scattering across the umbrella's top-level help. Mirrors
// the SPA's grouping in apps/.../controlHub/pages/Middleware/Overview.vue
// where password rotation is a discrete dialog launched from the
// instance card, not a top-level Middleware action.
package password

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewPasswordCommand assembles `olares-cli cluster middleware password`.
// Today's set is the Phase 6 slice (just `set`).
func NewPasswordCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "password",
		Short: "Rotate / set passwords on Olares middleware instances",
		Long: `Manage passwords on Olares middleware instances visible to the
active profile.

Endpoints (under https://control-hub.<terminus>):
  set    POST /middleware/v1/<type>/password
           body: {name, namespace, middleware, user, password}

Server-decides authority: a 403 means the operator's token can't
rotate passwords on this instance; we surface the error verbatim.
CLI does NOT preflight any local authorization check.
`,
	}
	cmd.SilenceUsage = true
	cmd.PersistentPreRun = func(c *cobra.Command, args []string) {
		c.SilenceUsage = true
	}

	cmd.AddCommand(NewSetCommand(f))

	return cmd
}
