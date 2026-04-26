// Package apps implements the `olares-cli settings apps` subtree (Settings ->
// Applications). This is the largest sub-tree by verb count: the SPA's
// per-app pages drive lifecycle (suspend/resume/uninstall), permission +
// entrance + domain config, secrets, and per-app env. Backed by
// user-service's app.controller.ts / application.controller.ts /
// bfl/application.controller.ts / bfl/env.controller.ts / secret.controller.ts.
//
// Lifecycle verbs (suspend / resume / uninstall) live here rather than under
// `market` because the underlying API is on the Settings desktop ingress,
// not the per-user market endpoint — see plan.md's "Settings vs Market"
// disambiguation.
package apps

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewAppsCommand returns the `settings apps` parent. Phase 1 ships the
// list / get reads; deeper per-app config verbs (permissions, entrances,
// domain, env, secrets, ACL) and lifecycle (suspend / resume / uninstall)
// land in Phase 3.
func NewAppsCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apps",
		Short: "Per-app settings (lifecycle, permissions, entrances, env, secrets, ACL)",
		Long: `Inspect and configure individual installed apps.

Subcommands:
  list   list installed apps                                (Phase 1)
  get    show one app's settings record                     (Phase 1)

Subcommands landing in later phases:
  Phase 3: suspend / resume / uninstall, permissions, entrances, providers,
           domain (get|set), policy (get|set), auth-level set, env, secrets,
           acl (get|set)

Note: install / upgrade / clone / cancel still live under "olares-cli market"
(per-user app-store API). "settings apps" is the *post-install* surface.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(NewListCommand(f))
	cmd.AddCommand(NewGetCommand(f))
	return cmd
}
