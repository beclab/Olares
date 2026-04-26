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

// NewAppsCommand returns the `settings apps` parent. Subcommands land in
// later phases; today the parent simply prints help.
func NewAppsCommand(_ *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apps",
		Short: "Per-app settings (lifecycle, permissions, entrances, env, secrets, ACL)",
		Long: `Inspect and configure individual installed apps.

Subcommands will be added in subsequent phases:
  Phase 1: list, status
  Phase 3: suspend / resume / uninstall, permissions, entrances, providers,
           domain (get|set), policy (get|set), auth-level set, env, secrets,
           acl (get|set)

Note: install / upgrade / clone / cancel still live under "olares-cli market"
(per-user app-store API). "settings apps" is the *post-install* surface.
`,
	}
	cmd.SilenceUsage = true
	return cmd
}
