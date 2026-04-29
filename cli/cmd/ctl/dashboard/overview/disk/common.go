// Package disk hosts the cobra wiring for `olares-cli dashboard
// overview disk` (root + main + partitions). Business logic lives
// in cli/pkg/dashboard/overview/disk/. This package is a thin shell
// that owns cobra wiring + the area-private *Client factory + the
// per-area unknown-subcommand hint.
//
// `var common` is wired by NewDiskCommand at construction; reads
// flow through cobra's persistent-flag inheritance which mutates
// the pointed-at struct before any leaf RunE runs.
package disk

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

var common *pkgdashboard.CommonFlags

// prepareClient is the area-private *pkgdashboard.Client factory.
func prepareClient(ctx context.Context, f *cmdutil.Factory) (*pkgdashboard.Client, error) {
	if f == nil {
		return nil, fmt.Errorf("internal error: disk not wired with cmdutil.Factory")
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

// unknownSubcommandRunE prints a typed typo hint + returns
// ErrAlreadyReported.
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
