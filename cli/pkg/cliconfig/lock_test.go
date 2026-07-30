package cliconfig

import "testing"

func seedProfile(t *testing.T, ids ...string) {
	t.Helper()
	cfg := &MultiProfileConfig{}
	for _, id := range ids {
		cfg.Upsert(ProfileConfig{OlaresID: id})
	}
	if err := SaveMultiProfileConfig(cfg); err != nil {
		t.Fatalf("seed: %v", err)
	}
}

func TestUpdateLockedReadModifyWrite(t *testing.T) {
	t.Setenv(homeEnv, t.TempDir())
	const id = "alice@olares.com"
	seedProfile(t, id)

	if err := UpdateLocked(func(cfg *MultiProfileConfig) error {
		cfg.FindByOlaresID(id).Name = "renamed"
		return nil
	}); err != nil {
		t.Fatalf("UpdateLocked: %v", err)
	}

	reloaded, _ := LoadMultiProfileConfig()
	if got := reloaded.FindByOlaresID(id).Name; got != "renamed" {
		t.Errorf("Name = %q, want renamed", got)
	}
}

// TestUpdateLockedRereadsLatest proves UpdateLocked operates on a fresh disk
// read, not a snapshot captured before the lock — so a writer holding a stale
// in-memory config can't clobber a concurrent writer's change.
func TestUpdateLockedRereadsLatest(t *testing.T) {
	t.Setenv(homeEnv, t.TempDir())
	const id = "alice@olares.com"
	seedProfile(t, id)

	// A stale handle loaded before any writes.
	stale, _ := LoadMultiProfileConfig()

	// Another writer sets the backend version out-of-band.
	if _, err := stale.SetBackendVersion(id, "1.0.0", 100); err != nil {
		t.Fatalf("set version: %v", err)
	}

	// A second writer that touches a DIFFERENT field via UpdateLocked must
	// preserve the version (because it re-reads), not wipe it.
	if err := UpdateLocked(func(cfg *MultiProfileConfig) error {
		cfg.FindByOlaresID(id).Name = "x"
		return nil
	}); err != nil {
		t.Fatalf("UpdateLocked: %v", err)
	}

	reloaded, _ := LoadMultiProfileConfig()
	p := reloaded.FindByOlaresID(id)
	if p.BackendVersion != "1.0.0" {
		t.Errorf("version was clobbered: %q, want 1.0.0", p.BackendVersion)
	}
	if p.Name != "x" {
		t.Errorf("Name = %q, want x", p.Name)
	}
}

func TestUpdateLockedNoChangeSkipsSave(t *testing.T) {
	t.Setenv(homeEnv, t.TempDir())
	const id = "alice@olares.com"
	seedProfile(t, id)

	err := UpdateLocked(func(cfg *MultiProfileConfig) error {
		return errNoConfigChange
	})
	if err != nil {
		t.Errorf("errNoConfigChange should surface as nil, got %v", err)
	}
}

func TestSetDetectResults(t *testing.T) {
	t.Setenv(homeEnv, t.TempDir())
	const id = "alice@olares.com"
	seedProfile(t, id)

	// Full pass: location + role + version all persisted in one call.
	if err := newCfg(t).SetDetectResults(id, "host", 10, "owner", 20, "1.12.0", 30); err != nil {
		t.Fatalf("SetDetectResults full: %v", err)
	}
	p := reload(t).FindByOlaresID(id)
	if p.Location != "host" || p.LocationProbedAt != 10 {
		t.Errorf("location not persisted: %+v", p)
	}
	if p.OwnerRole != "owner" || p.WhoamiRefreshedAt != 20 {
		t.Errorf("role not persisted: %+v", p)
	}
	if p.BackendVersion != "1.12.0" || p.BackendVersionRefreshedAt != 30 {
		t.Errorf("version not persisted: %+v", p)
	}

	// Partial pass: empty role/version leave the previously-cached values
	// untouched, only location updates.
	if err := newCfg(t).SetDetectResults(id, "lan", 40, "", 0, "", 0); err != nil {
		t.Fatalf("SetDetectResults partial: %v", err)
	}
	p = reload(t).FindByOlaresID(id)
	if p.Location != "lan" || p.LocationProbedAt != 40 {
		t.Errorf("location not updated on partial pass: %+v", p)
	}
	if p.OwnerRole != "owner" {
		t.Errorf("role should be preserved on partial pass, got %q", p.OwnerRole)
	}
	if p.BackendVersion != "1.12.0" {
		t.Errorf("version should be preserved on partial pass, got %q", p.BackendVersion)
	}
}

func newCfg(t *testing.T) *MultiProfileConfig {
	t.Helper()
	return &MultiProfileConfig{}
}

func reload(t *testing.T) *MultiProfileConfig {
	t.Helper()
	c, err := LoadMultiProfileConfig()
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	return c
}
