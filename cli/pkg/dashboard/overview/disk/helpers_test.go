package disk

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beclab/Olares/cli/pkg/credential"
	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
	"github.com/beclab/Olares/cli/pkg/dashboard/format"
)

// newTestClient mints a *pkgdashboard.Client pointed at srv. Mirrors
// the cli/pkg/dashboard/dashboard_test.go fixture so this subpackage
// stays self-contained without reaching into the parent _test.go
// file (cross-package _test.go is not addressable from Go's build
// graph).
func newTestClient(srv *httptest.Server) *pkgdashboard.Client {
	rp := &credential.ResolvedProfile{
		OlaresID:     "alice@olares.com",
		DashboardURL: srv.URL,
	}
	return pkgdashboard.NewClient(srv.Client(), rp)
}

// fixtureFlags returns a CommonFlags whose Validate() has already
// run. JSON output + zero head is the assertion-friendly default;
// individual tests override what they need.
func fixtureFlags(t *testing.T) *pkgdashboard.CommonFlags {
	t.Helper()
	cf := &pkgdashboard.CommonFlags{Timezone: format.LocalLocation()}
	if err := cf.Validate(); err != nil {
		t.Fatalf("CommonFlags.Validate: %v", err)
	}
	return cf
}

// noUnexpectedPath flags any upstream path the test didn't
// explicitly stub, so a wire-shape regression surfaces in a single
// line rather than a silent 404 deep inside the fetcher.
func noUnexpectedPath(t *testing.T, w http.ResponseWriter, path string) {
	t.Helper()
	t.Errorf("unexpected upstream path %q", path)
	w.WriteHeader(http.StatusNotFound)
}
