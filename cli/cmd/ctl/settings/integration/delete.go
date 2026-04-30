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
google, dropbox, space). The second is the optional account name —
required for object-storage flows where one user can have multiple
accounts of the same type, omittable for OAuth flows that only ever
have a single account per type per user.

Examples:
  olares-cli settings integration accounts delete awss3 AKIAIOSFODNN7EXAMPLE
  olares-cli settings integration accounts delete dropbox

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
