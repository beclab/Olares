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

// NewAppsCommand returns the `settings apps` parent. Lifecycle (suspend
// /resume), env, and secrets shipped in Phase 3a-3b. Per-app entrance
// config (permissions / providers / entrances list / domain / policy /
// auth-level) ships in Phase 3b2-3d.
func NewAppsCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apps",
		Short: "Per-app settings (lifecycle, permissions, entrances, env, secrets, ACL)",
		Long: `Inspect and configure individual installed apps.

Subcommands:
  list                                              list installed apps
  get <name>                                        show one app's settings record
  suspend <name>                                    suspend a running app
  resume  <name>                                    resume a suspended app
  env get|set <name>                                per-app environment variables
  secrets list|set|delete <app>                     per-app secret store

  permissions <app>                                 declared permissions vector
  providers list <app>                              registered provider vector
  entrances list <app>                              live entrance vector
  domain get|set|finish <app> <entrance>            per-entrance custom domain
  policy get|set <app> <entrance>                   per-entrance auth policy
  auth-level set <app> <entrance> --level X         per-entrance auth level

Note: install / upgrade / clone / cancel still live under "olares-cli market"
(per-user app-store API). "settings apps" is the *post-install* surface.
Per-app ACL (mesh allow-list) lives under "olares-cli settings vpn acl"
because the wire shape is shared with the VPN ACL family.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(NewListCommand(f))
	cmd.AddCommand(NewGetCommand(f))
	cmd.AddCommand(NewSuspendCommand(f))
	cmd.AddCommand(NewResumeCommand(f))
	cmd.AddCommand(NewEnvCommand(f))
	cmd.AddCommand(NewSecretsCommand(f))
	cmd.AddCommand(NewPermissionsCommand(f))
	cmd.AddCommand(NewProvidersCommand(f))
	cmd.AddCommand(NewEntrancesCommand(f))
	cmd.AddCommand(NewDomainCommand(f))
	cmd.AddCommand(NewPolicyCommand(f))
	cmd.AddCommand(NewAuthLevelCommand(f))
	return cmd
}
