// Package integration implements the `olares-cli settings integration`
// subtree (Settings -> Integration). Backed by user-service's
// account.controller.ts and cookie.controller.ts. The cloud-binding / NFT
// subsets are wallet-bound and stay out of CLI scope.
package integration

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewIntegrationCommand returns the `settings integration` parent:
// external integration accounts plus the per-domain cookie store that
// downloads and Wise collection read from.
func NewIntegrationCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "integration",
		Short: "External integration accounts and cookies (Settings -> Integration)",
		Long: `Manage integration accounts (S3 / Dropbox / Google Drive / Tencent COS / ...)
and the browser cookies used to download content behind a login.

Subcommands:
  accounts list
  accounts get <type> [name]
  accounts add awss3   [flags]
  accounts add tencent [flags]
  accounts delete <type> [name]
  cookie import --domain <d> --file <path>
  cookie list
  cookie rm <domain>
  cookie validate <domain>

OAuth flows (Google Drive, Dropbox) and the Olares-Space / NFT
cloud-binding flows stay in the SPA — they are browser- and
wallet-bound by design.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(NewAccountsCommand(f))
	cmd.AddCommand(NewCookieCommand(f))
	return cmd
}
