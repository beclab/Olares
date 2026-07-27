package whoami

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/beclab/Olares/cli/pkg/cliconfig"
	"github.com/beclab/Olares/cli/pkg/olares"
)

// fakeDoer answers the two detect round-trips from a canned body per path, so
// a detect pass can be driven without a network.
type fakeDoer map[string]string

func (f fakeDoer) DoJSON(_ context.Context, _, path string, _, out interface{}) error {
	body, ok := f[path]
	if !ok {
		return fmt.Errorf("unexpected request to %s", path)
	}
	return json.Unmarshal([]byte(body), out)
}

// stubDetectClient points DetectAndCache's fetches at d for the test's
// duration.
func stubDetectClient(t *testing.T, d Doer) {
	t.Helper()
	prev := newDetectClient
	newDetectClient = func(string, string, string, bool, olares.Location) Doer { return d }
	t.Cleanup(func() { newDetectClient = prev })
}

func detectFixture() fakeDoer {
	return fakeDoer{
		Endpoint:           `{"code":0,"data":{"name":"alice","owner_role":"owner"}}`,
		OlaresInfoEndpoint: `{"code":0,"data":{"osVersion":"1.12.6"}}`,
	}
}

// TestDetectAndCachePersists is the happy path: everything fetched, everything
// written, no error.
func TestDetectAndCachePersists(t *testing.T) {
	t.Setenv("OLARES_CLI_HOME", t.TempDir())
	const id = "alice@olares.com"

	cfg := &cliconfig.MultiProfileConfig{}
	cfg.Upsert(cliconfig.ProfileConfig{OlaresID: id})
	if err := cliconfig.SaveMultiProfileConfig(cfg); err != nil {
		t.Fatalf("seed: %v", err)
	}
	stubDetectClient(t, detectFixture())

	d, err := DetectAndCache(context.Background(), DetectInput{
		Cfg:           cfg,
		OlaresID:      id,
		AccessToken:   "AT-1",
		KnownLocation: olares.LocationHost,
	})
	if err != nil {
		t.Fatalf("DetectAndCache: %v", err)
	}
	if d.Location != string(olares.LocationHost) || d.Role != "owner" || d.BackendVersion != "1.12.6" {
		t.Fatalf("display = %+v, want host/owner/1.12.6", d)
	}

	persisted, err := cliconfig.LoadMultiProfileConfig()
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	p := persisted.FindByOlaresID(id)
	if p == nil {
		t.Fatal("profile vanished")
	}
	if p.Location != string(olares.LocationHost) || p.OwnerRole != "owner" || p.BackendVersion != "1.12.6" {
		t.Errorf("persisted = %+v, want host/owner/1.12.6", p)
	}
}

// TestDetectAndCacheReportsFailedCacheWrite is the regression for the
// swallowed `_ =` on SetDetectResults: when the locked write fails, the
// server-side facts are still valid but nothing reached config.json, so the
// caller must be told — otherwise the CLI prints freshly detected values under
// "(re-detected just now)" while every later command still reads the old ones.
//
// The write is made to fail the deterministic way SetDetectResults can fail:
// there is no profile on disk for this olaresId.
func TestDetectAndCacheReportsFailedCacheWrite(t *testing.T) {
	t.Setenv("OLARES_CLI_HOME", t.TempDir())
	stubDetectClient(t, detectFixture())

	d, err := DetectAndCache(context.Background(), DetectInput{
		Cfg:           &cliconfig.MultiProfileConfig{},
		OlaresID:      "ghost@olares.com",
		AccessToken:   "AT-1",
		KnownLocation: olares.LocationExternal,
	})
	if d == nil {
		t.Fatal("the fetched facts are still valid; expected a display")
	}
	if err == nil {
		t.Fatal("a failed cache write must be reported, not swallowed")
	}
	if !errors.Is(err, ErrCacheWrite) {
		t.Errorf("error = %v, want one matching ErrCacheWrite", err)
	}
}

// TestDetectAndCacheReportsBothFailures: a fetch failure and a cache-write
// failure in the same pass must both survive into the returned error, so a
// caller testing for ErrCacheWrite isn't fooled by the fetch error arriving
// first.
func TestDetectAndCacheReportsBothFailures(t *testing.T) {
	t.Setenv("OLARES_CLI_HOME", t.TempDir())

	// Role answers, version doesn't → a fetch error; no profile on disk → a
	// cache-write error.
	partial := detectFixture()
	delete(partial, OlaresInfoEndpoint)
	stubDetectClient(t, partial)

	d, err := DetectAndCache(context.Background(), DetectInput{
		Cfg:           &cliconfig.MultiProfileConfig{},
		OlaresID:      "ghost@olares.com",
		AccessToken:   "AT-1",
		KnownLocation: olares.LocationExternal,
	})
	if d == nil || d.Role != "owner" {
		t.Fatalf("the role fetch succeeded; expected it on the display, got %+v", d)
	}
	if !errors.Is(err, ErrCacheWrite) {
		t.Errorf("error = %v, want the cache-write failure to survive alongside the fetch error", err)
	}
	if err == nil || !strings.Contains(err.Error(), OlaresInfoEndpoint) {
		t.Errorf("error = %v, want the version fetch failure retained too", err)
	}
}
