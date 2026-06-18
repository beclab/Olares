// Package apps implements the `olares-cli settings apps` subtree (Settings ->
// Applications). This is the largest sub-tree by verb count: the SPA's
// per-app pages drive lifecycle (suspend/resume/uninstall), entrance +
// domain config, and per-app env. Backed by user-service's
// app.controller.ts / application.controller.ts / bfl/application.controller.ts /
// bfl/env.controller.ts.
//
// Lifecycle verbs (suspend / resume / uninstall) live here rather than under
// `market` because the underlying API is on the Settings desktop ingress,
// not the per-user market endpoint — see plan.md's "Settings vs Market"
// disambiguation.
package apps

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/market"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewAppsCommand returns the `settings apps` parent: lifecycle
// (suspend / resume), per-app env, and per-app entrance config
// (entrances list, domain, policy, auth-level).
func NewAppsCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apps",
		Short: "Per-app settings (lifecycle, entrances, env, ACL)",
		Long: `Inspect and configure individual installed apps.

Subcommands:
  list                                              list installed apps
  get <name>                                        show one app's settings record
  suspend <name>                                    suspend a running app
  resume  <name>                                    resume a suspended app
  env get|set <name>                                per-app environment variables

  entrances list <app>                              live entrance vector
  domain get|list|set|finish <app> [<entrance>]     per-entrance custom domain
  policy get|list|set <app> [<entrance>]            per-entrance auth policy
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
	// suspend / resume hold NO settings-side logic: 1.12.6's settings page
	// routes stop/resume through the Market flow (findAppByName ->
	// AppService.stopApp/resumeApp). We mirror that by reusing the market
	// stop/resume commands verbatim (single source of truth for the
	// shared/csv2 cascade), only renaming the verb. They talk to the
	// app-store v2 API via the active profile, same as `market stop/resume`.
	suspend := market.NewCmdMarketStop(f)
	suspend.Use = "suspend <name>"
	suspend.Aliases = nil
	suspend.Short = "Suspend (stop) a running app (delegates to `market stop`)"
	cmd.AddCommand(suspend)
	resume := market.NewCmdMarketResume(f)
	resume.Use = "resume <name>"
	resume.Aliases = nil
	resume.Short = "Resume a suspended app (delegates to `market resume`)"
	cmd.AddCommand(resume)
	cmd.AddCommand(NewEnvCommand(f))
	cmd.AddCommand(NewEntrancesCommand(f))
	cmd.AddCommand(NewDomainCommand(f))
	cmd.AddCommand(NewPolicyCommand(f))
	cmd.AddCommand(NewAuthLevelCommand(f))
	return cmd
}
