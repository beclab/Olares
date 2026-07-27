package cmdutil

import (
	"context"
	"testing"
	"time"

	"github.com/beclab/Olares/cli/pkg/access"
	"github.com/beclab/Olares/cli/pkg/cliconfig"
	"github.com/beclab/Olares/cli/pkg/credential"
	"github.com/beclab/Olares/cli/pkg/olares"
)

func TestInLocationCooldown(t *testing.T) {
	t.Setenv("OLARES_CLI_HOME", t.TempDir())

	const id = "alice@olares.com"
	const base = int64(1_000_000)
	now := time.Unix(base, 0)

	cfg := &cliconfig.MultiProfileConfig{}
	cfg.Upsert(cliconfig.ProfileConfig{OlaresID: id})

	// No outage stamp → never in cooldown.
	if err := cliconfig.SaveMultiProfileConfig(cfg); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if inLocationCooldown(id, now) {
		t.Error("a profile with no outage stamp should not be in cooldown")
	}

	// Recent outage (10s ago) → still inside the 30s window.
	cfg.FindByOlaresID(id).LocationUnreachableAt = base - 10
	if err := cliconfig.SaveMultiProfileConfig(cfg); err != nil {
		t.Fatalf("save recent: %v", err)
	}
	if !inLocationCooldown(id, now) {
		t.Error("a 10s-old outage should be inside the cooldown window")
	}

	// Old outage (well past 30s) → window expired.
	cfg.FindByOlaresID(id).LocationUnreachableAt = base - 120
	if err := cliconfig.SaveMultiProfileConfig(cfg); err != nil {
		t.Fatalf("save old: %v", err)
	}
	if inLocationCooldown(id, now) {
		t.Error("a 120s-old outage should be outside the cooldown window")
	}

	// Unknown profile → not in cooldown (and no panic).
	if inLocationCooldown("ghost@olares.com", now) {
		t.Error("an unknown profile should not be in cooldown")
	}
}

// TestClearUnreachableReArms is the regression for the clearOnce bug: in a
// long-lived process, a second outage→recovery cycle must still clear the
// cooldown stamp (the old sync.Once collapsed every clear after the first into
// a no-op).
func TestClearUnreachableReArms(t *testing.T) {
	t.Setenv("OLARES_CLI_HOME", t.TempDir())

	const id = "alice@olares.com"
	cfg := &cliconfig.MultiProfileConfig{}
	cfg.Upsert(cliconfig.ProfileConfig{OlaresID: id})
	if err := cliconfig.SaveMultiProfileConfig(cfg); err != nil {
		t.Fatalf("seed: %v", err)
	}

	now := int64(1_000_000)
	tr := &refreshingTransport{
		olaresID: id,
		loc:      &locationState{},
		now:      func() time.Time { return time.Unix(now, 0) },
	}

	stamp := func() int64 {
		c, err := cliconfig.LoadMultiProfileConfig()
		if err != nil {
			t.Fatalf("load: %v", err)
		}
		return c.FindByOlaresID(id).LocationUnreachableAt
	}

	// Cycle 1: mark → stamp set; clear → stamp lifted.
	tr.markUnreachable()
	if got := stamp(); got != now {
		t.Fatalf("after first mark, stamp = %d, want %d", got, now)
	}
	tr.clearUnreachable()
	if got := stamp(); got != 0 {
		t.Fatalf("after first clear, stamp = %d, want 0", got)
	}

	// Cycle 2: the bug lived here — the second clear used to be a no-op.
	now = 2_000_000
	tr.markUnreachable()
	if got := stamp(); got != now {
		t.Fatalf("after second mark, stamp = %d, want %d", got, now)
	}
	tr.clearUnreachable()
	if got := stamp(); got != 0 {
		t.Fatalf("after second clear, stamp = %d, want 0 (clearOnce regression)", got)
	}
}

// TestClearUnreachableNoMarkIsNoOp verifies the CAS gate: a success with no
// preceding mark this run does not touch a stamp left by another writer.
func TestClearUnreachableNoMarkIsNoOp(t *testing.T) {
	t.Setenv("OLARES_CLI_HOME", t.TempDir())

	const id = "alice@olares.com"
	cfg := &cliconfig.MultiProfileConfig{}
	cfg.Upsert(cliconfig.ProfileConfig{OlaresID: id, LocationUnreachableAt: 555})
	if err := cliconfig.SaveMultiProfileConfig(cfg); err != nil {
		t.Fatalf("seed: %v", err)
	}

	// unreachableMarked defaults false (not armed this run).
	tr := &refreshingTransport{olaresID: id, loc: &locationState{}}
	tr.clearUnreachable()

	c, _ := cliconfig.LoadMultiProfileConfig()
	if got := c.FindByOlaresID(id).LocationUnreachableAt; got != 555 {
		t.Errorf("unmarked clear should leave the stamp untouched, got %d want 555", got)
	}
}

// stubProbe forces probeLocationFn to a fixed answer for the duration of the
// test and reports how many probes ran.
func stubProbe(t *testing.T, loc olares.Location, err error) *int {
	t.Helper()
	calls := 0
	prev := probeLocationFn
	probeLocationFn = func(context.Context, olares.ID, string, bool) (olares.Location, error) {
		calls++
		return loc, err
	}
	t.Cleanup(func() { probeLocationFn = prev })
	return &calls
}

// seedProfile writes a single profile with the given outage stamp.
func seedProfile(t *testing.T, id string, unreachableAt int64) {
	t.Helper()
	cfg := &cliconfig.MultiProfileConfig{}
	cfg.Upsert(cliconfig.ProfileConfig{OlaresID: id, LocationUnreachableAt: unreachableAt})
	if err := cliconfig.SaveMultiProfileConfig(cfg); err != nil {
		t.Fatalf("seed profile: %v", err)
	}
}

// TestBackfillDeferredByCooldownIsRetried is the regression for the
// once-and-only-once backfill: a probe skipped because of the outage cooldown
// must be re-attempted later in the same process. Before the fix, the single
// call inside ResolveProfile's sync.Once was the process's only chance, so a
// long-running command that started inside the 30s window stayed on the
// external defaults for its whole run.
func TestBackfillDeferredByCooldownIsRetried(t *testing.T) {
	t.Setenv("OLARES_CLI_HOME", t.TempDir())

	const id = "alice@olares.com"
	now := time.Now()
	seedProfile(t, id, now.Unix()-5) // outage 5s ago → inside the window

	calls := stubProbe(t, olares.LocationLAN, nil)
	ctx := context.Background()
	f := NewFactory()
	rp := &credential.ResolvedProfile{OlaresID: id, Source: "default"}

	// First pass (what ResolveProfile's Once does): cooldown defers the probe
	// and leaves the profile on its external defaults, but arms the retry.
	f.maybeBackfillLocation(ctx, rp)
	f.resolved = rp
	if *calls != 0 {
		t.Fatalf("cooldown should have skipped the probe, ran %d", *calls)
	}
	if rp.Location.Valid() {
		t.Fatalf("location should still be unknown, got %q", rp.Location)
	}
	if !f.locationBackfillPending.Load() {
		t.Fatal("a cooldown-deferred backfill must arm the retry")
	}

	// Still inside the window: stay deferred rather than burning a probe.
	if got := f.retryLocationBackfill(ctx, rp); got != rp {
		t.Error("a retry inside the cooldown window must keep the current profile")
	}
	if *calls != 0 {
		t.Fatalf("retry inside the window should not probe, ran %d", *calls)
	}
	if !f.locationBackfillPending.Load() {
		t.Fatal("the retry must stay armed while the window is still open")
	}

	// Window lapses (connectivity restored) → the retry finally probes.
	seedProfile(t, id, 0)
	got := f.retryLocationBackfill(ctx, rp)
	if *calls != 1 {
		t.Fatalf("expected exactly one probe after the window lapsed, ran %d", *calls)
	}
	if got.Location != olares.LocationLAN {
		t.Errorf("republished location = %q, want lan", got.Location)
	}
	if got == rp {
		t.Fatal("the probed profile must be published as a copy, not written into the pointer callers already hold")
	}
	if rp.Location.Valid() || rp.DesktopURL != "" {
		t.Errorf("the original profile must be left untouched, got location %q / desktop %q", rp.Location, rp.DesktopURL)
	}
	if got.DesktopURL == "" {
		t.Error("the copy should carry URLs re-derived for the probed location")
	}

	// Persisted, so the next process starts on the fast path directly.
	cfg, err := cliconfig.LoadMultiProfileConfig()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if p := cfg.FindByOlaresID(id); p == nil || p.Location != string(olares.LocationLAN) {
		t.Errorf("config location = %+v, want lan", p)
	}

	// An already-built client is moved onto the probed position too, instead
	// of waiting for a network error to trigger ensureSwitched.
	if ls := f.sharedLocationState(got); ls.loc != olares.LocationLAN {
		t.Errorf("shared location state = %q, want lan", ls.loc)
	}

	// Disarmed: a success is final, later calls are free.
	if f.locationBackfillPending.Load() {
		t.Error("a completed backfill must not stay armed")
	}
	if f.retryLocationBackfill(ctx, got); *calls != 1 {
		t.Errorf("no further probes expected after a successful backfill, ran %d", *calls)
	}
}

// TestBackfillRetryGivesUpAfterFailedProbe: the deferred retry gets exactly
// one shot. If that probe fails we fall back to the external defaults for the
// rest of the process rather than re-paying a multi-second probe on every
// ResolveProfile call; runtime recovery is ensureSwitched's job from there.
func TestBackfillRetryGivesUpAfterFailedProbe(t *testing.T) {
	t.Setenv("OLARES_CLI_HOME", t.TempDir())

	const id = "alice@olares.com"
	seedProfile(t, id, 0)

	calls := stubProbe(t, "", &access.UnreachableError{OlaresID: id})
	ctx := context.Background()
	f := NewFactory()
	rp := &credential.ResolvedProfile{OlaresID: id, Source: "default"}
	f.resolved = rp
	f.locationBackfillPending.Store(true)

	if got := f.retryLocationBackfill(ctx, rp); got != rp {
		t.Error("a failed probe must leave the current profile in place")
	}
	if *calls != 1 {
		t.Fatalf("expected one probe, ran %d", *calls)
	}
	if f.locationBackfillPending.Load() {
		t.Error("a failed retry must disarm; otherwise every later call pays the probe")
	}
	f.retryLocationBackfill(ctx, rp)
	if *calls != 1 {
		t.Errorf("expected no second probe, ran %d", *calls)
	}
}

// TestNeedsLocationBackfill covers the eligibility rules: only an on-disk
// profile with an unknown location and no pinned auth URL gets probed.
func TestNeedsLocationBackfill(t *testing.T) {
	cases := []struct {
		name string
		rp   *credential.ResolvedProfile
		want bool
	}{
		{"nil", nil, false},
		{"unprobed on-disk profile", &credential.ResolvedProfile{Source: "default"}, true},
		{"env-resolved", &credential.ResolvedProfile{Source: "env"}, false},
		{"pinned auth URL", &credential.ResolvedProfile{Source: "default", AuthURLOverride: "https://auth.example"}, false},
		{"already probed", &credential.ResolvedProfile{Source: "default", Location: olares.LocationHost}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := needsLocationBackfill(tc.rp); got != tc.want {
				t.Errorf("needsLocationBackfill = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLocationProbeBudget(t *testing.T) {
	want := access.MaxProbeDuration() + time.Second
	if locationProbeBudget != want {
		t.Errorf("locationProbeBudget = %v, want %v (MaxProbeDuration + 1s)", locationProbeBudget, want)
	}
}
