package cliconfig

import "testing"

func TestSetLocation(t *testing.T) {
	t.Setenv(homeEnv, t.TempDir())

	const id = "alice@olares.com"
	seed := &MultiProfileConfig{}
	seed.Upsert(ProfileConfig{OlaresID: id, LocationUnreachableAt: 999})
	if err := SaveMultiProfileConfig(seed); err != nil {
		t.Fatalf("seed save: %v", err)
	}

	cfg, err := LoadMultiProfileConfig()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if err := cfg.SetLocation(id, "lan", 1000); err != nil {
		t.Fatalf("set: %v", err)
	}

	reloaded, err := LoadMultiProfileConfig()
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	p := reloaded.FindByOlaresID(id)
	if p == nil || p.Location != "lan" || p.LocationProbedAt != 1000 {
		t.Fatalf("persisted = %+v, want location lan @ 1000", p)
	}
	// A successful probe must clear any prior outage cooldown.
	if p.LocationUnreachableAt != 0 {
		t.Errorf("LocationUnreachableAt = %d, want 0 (SetLocation clears cooldown)", p.LocationUnreachableAt)
	}

	if err := reloaded.SetLocation("bob@olares.com", "lan", 1); err == nil {
		t.Error("setting location for an unknown profile should error")
	}
}

func TestSetAndClearLocationUnreachable(t *testing.T) {
	t.Setenv(homeEnv, t.TempDir())

	const id = "alice@olares.com"
	seed := &MultiProfileConfig{}
	seed.Upsert(ProfileConfig{OlaresID: id, Location: "host"})
	if err := SaveMultiProfileConfig(seed); err != nil {
		t.Fatalf("seed save: %v", err)
	}

	cfg, _ := LoadMultiProfileConfig()
	if err := cfg.SetLocationUnreachable(id, 5000); err != nil {
		t.Fatalf("mark unreachable: %v", err)
	}

	reloaded, _ := LoadMultiProfileConfig()
	p := reloaded.FindByOlaresID(id)
	if p.LocationUnreachableAt != 5000 {
		t.Fatalf("LocationUnreachableAt = %d, want 5000", p.LocationUnreachableAt)
	}
	// Last-known-good location is preserved across an outage.
	if p.Location != "host" {
		t.Errorf("Location = %q, want host preserved across outage", p.Location)
	}

	if err := reloaded.ClearLocationUnreachable(id); err != nil {
		t.Fatalf("clear: %v", err)
	}
	final, _ := LoadMultiProfileConfig()
	if final.FindByOlaresID(id).LocationUnreachableAt != 0 {
		t.Error("ClearLocationUnreachable should reset the stamp to 0")
	}
}

func TestClearLocationUnreachableNoOpWhenZero(t *testing.T) {
	t.Setenv(homeEnv, t.TempDir())

	const id = "alice@olares.com"
	cfg := &MultiProfileConfig{}
	cfg.Upsert(ProfileConfig{OlaresID: id})
	if err := SaveMultiProfileConfig(cfg); err != nil {
		t.Fatalf("seed: %v", err)
	}
	// Already 0 → returns nil without error (and without needing a write).
	if err := cfg.ClearLocationUnreachable(id); err != nil {
		t.Errorf("clear on already-zero should be a no-op nil, got %v", err)
	}
}
