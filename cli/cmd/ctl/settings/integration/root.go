// Package integration implements the `olares-cli settings integration`
// subtree (Settings -> Integration). Backed by user-service's
// account.controller.ts. The cookie / cloud-binding / NFT subsets are
// browser-/wallet-bound and stay out of CLI scope; only the headless account
// CRUD is in.
package integration

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewIntegrationCommand returns the `settings integration` parent:
// inspect, add (object-storage credentials only) and delete external
// integration accounts.
func NewIntegrationCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "integration",
		Short: "External integration accounts (Settings -> Integration)",
		Long: `Manage integration accounts (S3 / Dropbox / Google Drive / Tencent COS / ...).

Subcommands:
  accounts list
  accounts get <type> [name]
  accounts add awss3   [flags]
  accounts add tencent [flags]
  accounts delete <type> [name]

OAuth flows (Google Drive, Dropbox), the cookie store and the
Olares-Space / NFT cloud-binding flows stay in the SPA — they are
browser- and wallet-bound by design.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(NewAccountsCommand(f))
	return cmd
}
