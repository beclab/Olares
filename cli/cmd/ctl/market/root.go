package market

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewMarketCommand assembles the `olares-cli market` subtree. Identity (which
// Olares user) and transport (which cluster) are resolved from the
// currently-selected profile via cmdutil.Factory (switch with
// `olares-cli profile use <name>`) rather than per-command flags, so this
// tree intentionally diverges from `cli/cmd/ctl/app` (which still discovers
// the cluster via kubeconfig + X-Bfl-User). The two trees can therefore be
// reviewed side-by-side; once `market` is GA the `app` tree should retire.
//
// Note: the legacy `sync` verb is intentionally not exposed here. It depends
// on chart-repo-service which is only reachable from inside the cluster
// network, not via the user-subdomain edge route the rest of the market API
// goes through.
func NewMarketCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "market",
		Short: "Manage Olares applications via the per-user Market API",
		Long: `Manage applications through the Olares Market app-store API.

This command tree is the profile-based parallel of "olares-cli app":
identity (which Olares user) and transport (which cluster) are
resolved from the currently-selected profile (switch with
"olares-cli profile use <name>") instead of from kubeconfig + --user.
Authentication uses the access token from "olares-cli profile login"
and the same edge auth chain the Olares web app uses
(Authelia + l4-bfl-proxy).

Verb families:

  catalog (read-only)   list, categories, get         browse /market/data
  runtime (read-only)   status                        read /market/state
  lifecycle (mutating)  install, upgrade, uninstall,  POST/PUT/DELETE on
                        clone, stop, resume, cancel    /apps/{name}/*
  charts (mutating)     upload, delete                 SPA Local Sources

Universal flags:

  -o, --output {table,json}   every verb. JSON output is parseable.
  -q, --quiet                 every verb. Suppress output; exit code wins.
  -s, --source <id>           source-aware verbs (catalog + install /
                              upgrade / clone / upload / delete).
                              Valid ids: market.olares, cli, upload, studio.
  -a, --all-sources           list, categories, status.
      --no-headers            list, categories, get (table-rendering verbs).
  -w, --watch [+ --watch-timeout, --watch-interval]
                              lifecycle verbs. Block until terminal state.

The "my apps" inventory lives under "list --mine" (-m): same set the
Market UI's "My Terminus" tab shows (broader than "completed installs"
— includes in-flight rows and post-install failures, hides only the
6 SPA uninstalledAppStates). Use it instead of "status" when answering
"what apps do I have".

Run "olares-cli market <verb> --help" for verb-specific details
(source resolution, --watch idempotency, auto-cascade behavior,
pre-flight gates, etc.). For deep-dive documentation, see
cli/skills/olares-market/SKILL.md.`,
	}
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.PersistentPreRun = func(c *cobra.Command, args []string) {
		c.SilenceErrors = true
		c.SilenceUsage = true
	}

	cmd.AddCommand(NewCmdMarketList(f))
	cmd.AddCommand(NewCmdMarketCategories(f))
	cmd.AddCommand(NewCmdMarketGet(f))
	cmd.AddCommand(NewCmdMarketInstall(f))
	cmd.AddCommand(NewCmdMarketUninstall(f))
	cmd.AddCommand(NewCmdMarketUpgrade(f))
	cmd.AddCommand(NewCmdMarketClone(f))
	cmd.AddCommand(NewCmdMarketCancel(f))
	cmd.AddCommand(NewCmdMarketStop(f))
	cmd.AddCommand(NewCmdMarketResume(f))
	cmd.AddCommand(NewCmdMarketUpload(f))
	cmd.AddCommand(NewCmdMarketDelete(f))
	cmd.AddCommand(NewCmdMarketStatus(f))
	return cmd
}
