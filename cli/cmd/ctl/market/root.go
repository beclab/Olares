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

This command tree is the profile-based parallel of "olares-cli app": same
verbs (install / upgrade / uninstall / list / status / clone / upload / ...),
but identity and the API endpoint are resolved from the currently-selected
profile (switch with "olares-cli profile use <name>") instead of from
kubeconfig + --user.

Authentication uses the access token from "olares-cli profile login" and the
same edge auth chain the Olares web app uses (Authelia + l4-bfl-proxy).`,
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
