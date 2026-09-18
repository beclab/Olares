// Package overview hosts the cobra subtree for `olares-cli dashboard
// overview`. Per the settings-style layout (cli/cmd/ctl/settings/<area>/),
// every cobra leaf lives in its own .go file and the directory tree
// mirrors the command tree:
//
//	overview/                       (this package — root.go + 7 leaves)
//	  ├── disk/                     (subgroup; own Go package)
//	  ├── fan/
//	  └── gpu/
//
// Business logic lives in cli/pkg/dashboard/overview/. This package
// is a thin shell — it owns cobra wiring + the area-private
// *Client factory. `var common`
// is a *pkgdashboard.CommonFlags pointer set by NewOverviewCommand at
// construction; cobra's persistent-flag inheritance from the dashboard
// root populates the pointed-at struct before any leaf RunE fires.
package overview

import (
	"context"
	"fmt"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// common is the shared CommonFlags pointer; mutated by cobra's
// persistent-flag inheritance from the dashboard root before every
// leaf RunE.
var common *pkgdashboard.CommonFlags

// prepareClient is the area-private *pkgdashboard.Client factory.
// Each cmd subpackage writes its own (settings precedent) so the
// transport seam between cmdutil.Factory and pkgdashboard.Client is
// inspectable per area.
func prepareClient(ctx context.Context, f *cmdutil.Factory) (*pkgdashboard.Client, error) {
	if f == nil {
		return nil, fmt.Errorf("internal error: overview not wired with cmdutil.Factory")
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
