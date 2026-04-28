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
//
// Removed verbs (Phase 6 cleanup, 2026-04-28; see KNOWN_ISSUES.md KI-4 / KI-5 / KI-6):
//   - permissions <app>           backend GET /api/applications/permissions/<app> 整条路由 404
//                                  (user-service 已无 application/permission 控制器)；
//                                  SPA UI 入口被 v-if + 注释守卫成 dead path、
//                                  store getPermissions 已下架。
//   - providers list <app>        同源；后端 GET /api/applications/provider/registry/<app>
//                                  整条路由没了；SPA UI 入口已注释。
//   - secrets list/set/delete     旧路径 /admin/secret/<app> 被 desktop ingress 当 SPA
//                                  路由吞掉（GET 返 index.html、POST 返 nginx 405）；
//                                  SPA secretStore.checkSecretPermission() 永远 false，
//                                  ApplicationSecretPage UI 入口 dead。
package apps

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewAppsCommand returns the `settings apps` parent. Lifecycle (suspend
// /resume), env shipped in Phase 3a-3b. Per-app entrance config
// (entrances list / domain / policy / auth-level) ships in Phase 3b2-3d.
// permissions / providers / secrets removed in Phase 6 cleanup
// (2026-04-28; see package doc / KNOWN_ISSUES.md KI-4 / KI-5 / KI-6).
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
	cmd.AddCommand(NewSuspendCommand(f))
	cmd.AddCommand(NewResumeCommand(f))
	cmd.AddCommand(NewEnvCommand(f))
	cmd.AddCommand(NewEntrancesCommand(f))
	cmd.AddCommand(NewDomainCommand(f))
	cmd.AddCommand(NewPolicyCommand(f))
	cmd.AddCommand(NewAuthLevelCommand(f))
	return cmd
}
