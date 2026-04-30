// Package applications hosts the cobra wiring for `olares-cli
// dashboard applications`; business logic lives in
// cli/pkg/dashboard/applications/.
package applications

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// common is wired by NewApplicationsCommand; cobra's persistent-flag
// inheritance mutates the pointed-at struct before any leaf RunE fires.
var common *pkgdashboard.CommonFlags

// prepareClient is the area-private *pkgdashboard.Client factory.
func prepareClient(ctx context.Context, f *cmdutil.Factory) (*pkgdashboard.Client, error) {
	if f == nil {
		return nil, fmt.Errorf("internal error: applications not wired with cmdutil.Factory")
	}
	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return nil, err
	}
	hc, err := f.HTTPClient(ctx)
	if err != nil {
		return nil, err
	}
	return pkgdashboard.NewClient(hc, rp), nil
}

// unknownSubcommandRunE prints a typed typo hint + returns ErrAlreadyReported.
func unknownSubcommandRunE(c *cobra.Command, args []string) error {
	if len(args) == 0 {
		return c.Help()
	}
	msg := fmt.Sprintf("Error: unknown subcommand %q for %q", args[0], c.CommandPath())
	if suggestions := c.SuggestionsFor(args[0]); len(suggestions) > 0 {
		msg += "\n\nDid you mean this?\n\t" + strings.Join(suggestions, "\n\t")
	}
	fmt.Fprintln(c.ErrOrStderr(), msg)
	fmt.Fprintf(c.ErrOrStderr(), "\nRun '%s --help' for usage.\n", c.CommandPath())
	return pkgdashboard.ErrAlreadyReported
}
