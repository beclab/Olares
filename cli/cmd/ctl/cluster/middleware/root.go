// Package middleware implements `olares-cli cluster middleware ...`.
//
// Olares' middleware controller exposes a small per-user inventory
// of managed databases / queues / object stores under
// /middleware/v1/* on the ControlHub origin. Today we only model
// the `list` verb. Mutating verbs (password rotation lives at
// /middleware/v1/<type>/password in the SPA) belong to a future
// `cluster middleware password set` once we lock down the
// confirmation UX — that's destructive and ships with Phase 6.
package middleware

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewMiddlewareCommand assembles `olares-cli cluster middleware`.
func NewMiddlewareCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "middleware",
		Aliases: []string{"middlewares", "mw"},
		Short:   "List Olares middleware instances visible to the active profile",
		Long: `List Olares middleware instances (managed databases / queues /
object stores) visible to the active profile.

Backed by https://control-hub.<terminus>/middleware/v1/list, the
same endpoint the ControlHub SPA's "Middleware" page uses. The
endpoint returns a custom envelope ({code, data:[MiddlewareItem]})
rather than a K8s-native shape — see
apps/packages/app/src/apps/controlPanelCommon/network/middleware.ts.

Sensitive fields (admin password) are redacted by default. Pass
--show-passwords to include them in -o json output (table mode never
shows passwords).
`,
	}
	cmd.SilenceUsage = true
	cmd.PersistentPreRun = func(c *cobra.Command, args []string) {
		c.SilenceUsage = true
	}

	cmd.AddCommand(NewListCommand(f))

	return cmd
}
