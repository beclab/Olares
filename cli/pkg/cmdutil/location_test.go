package cmdutil

import (
	"testing"
	"time"

	"github.com/beclab/Olares/cli/pkg/access"
	"github.com/beclab/Olares/cli/pkg/cliconfig"
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

func TestLocationProbeBudget(t *testing.T) {
	want := access.MaxProbeDuration() + time.Second
	if locationProbeBudget != want {
		t.Errorf("locationProbeBudget = %v, want %v (MaxProbeDuration + 1s)", locationProbeBudget, want)
	}
}
