package integration

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings integration accounts delete <type> [name]`
//
// Backed by DELETE /api/account/<storeKey> on user-service's
// account.controller.ts. The SPA's stores/settings/integration.ts
// constructs the storeKey via getStoreKey(data):
//
//   integration-account:<type>:<name>      // when name is set
//   integration-account:<type>             // legacy / single-account-of-type
//
// We mirror that logic exactly so the CLI hits the same row the SPA's
// "Delete account" button targets.

func newAccountsDeleteCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <type> [name]",
		Short: "delete an integration account",
		Long: `Delete an integration account.

The first positional arg is the account type (e.g. awss3, tencent,
google, dropbox, space). The second is the account name. In practice
every account type the SPA can create sets a name during the Add-Account
flow, so name is effectively required to target the right row:

  - object-storage flows (awss3, tencent): name is the access-key id
    that was passed to "accounts add <type> --access-key-id …"
  - OAuth / wallet flows (dropbox, google, space): name is the per-user
    identity returned by the upstream provider during sign-in
    (Dropbox uid, Google account email, Olares Space did or username)

The bare "<type>" form (omit name) is only kept for backward compatibility
with legacy rows that were written before the SPA started setting a name
on every account; do not rely on it for accounts created through the
current SPA or this CLI.

Examples:
  olares-cli settings integration accounts delete awss3   AKIAIOSFODNN7EXAMPLE
  olares-cli settings integration accounts delete dropbox 123456789
  olares-cli settings integration accounts delete google  alice@example.com

Use "olares-cli settings integration accounts list" first to see the
exact (type, name) tuple of each account.
`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(c *cobra.Command, args []string) error {
			name := ""
			if len(args) > 1 {
				name = args[1]
			}
			return runAccountsDelete(c.Context(), f, args[0], name)
		},
	}
	return cmd
}

func runAccountsDelete(ctx context.Context, f *cmdutil.Factory, accountType, name string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	accountType = strings.TrimSpace(accountType)
	name = strings.TrimSpace(name)
	if accountType == "" {
		return fmt.Errorf("account type is required")
	}

	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}

	storeKey := "integration-account:" + accountType
	if name != "" {
		storeKey += ":" + name
	}
	path := "/api/account/" + url.PathEscape(storeKey)

	if err := doMutateEnvelope(ctx, pc.doer, "DELETE", path, nil, nil); err != nil {
		return err
	}
	if name != "" {
		fmt.Printf("Deleted %s account %q.\n", accountType, name)
	} else {
		fmt.Printf("Deleted %s account.\n", accountType)
	}
	return nil
}
