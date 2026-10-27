// Package fan hosts the cobra wiring for `olares-cli dashboard
// overview fan` (root + live + curve). Business logic lives in
// cli/pkg/dashboard/overview/fan/. This package is a thin shell
// that owns cobra wiring + the area-private *Client factory + the
// per-area unknown-subcommand hint.
//
// `var common` is wired by NewFanCommand at construction; reads
// flow through cobra's persistent-flag inheritance which mutates
// the pointed-at struct before any leaf RunE runs.
package fan

import (
	"context"
	"fmt"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

var common *pkgdashboard.CommonFlags

// prepareClient is the area-private *pkgdashboard.Client factory.
func prepareClient(ctx context.Context, f *cmdutil.Factory) (*pkgdashboard.Client, error) {
	if f == nil {
		return nil, fmt.Errorf("internal error: fan not wired with cmdutil.Factory")
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
