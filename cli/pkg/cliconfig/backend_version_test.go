package cliconfig

import (
	"testing"
)

func TestSetBackendVersion(t *testing.T) {
	// Isolate the config dir to a temp location.
	t.Setenv(homeEnv, t.TempDir())

	const id = "alice@olares.com"
	seed := &MultiProfileConfig{}
	seed.Upsert(ProfileConfig{OlaresID: id})
	if err := SaveMultiProfileConfig(seed); err != nil {
		t.Fatalf("seed save: %v", err)
	}

	// First write: empty -> version reports changed.
	cfg, err := LoadMultiProfileConfig()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	changed, err := cfg.SetBackendVersion(id, "1.12.5", 1000)
	if err != nil {
		t.Fatalf("set: %v", err)
	}
	if !changed {
		t.Error("first write empty->1.12.5 should report changed=true")
	}

	// Persisted across reloads.
	reloaded, err := LoadMultiProfileConfig()
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	p := reloaded.FindByOlaresID(id)
	if p == nil || p.BackendVersion != "1.12.5" || p.BackendVersionRefreshedAt != 1000 {
		t.Fatalf("persisted profile = %+v, want version 1.12.5 @ 1000", p)
	}

	// Same version again: not changed (but timestamp still updated).
	changed, err = reloaded.SetBackendVersion(id, "1.12.5", 2000)
	if err != nil {
		t.Fatalf("set same: %v", err)
	}
	if changed {
		t.Error("rewriting the same version should report changed=false")
	}
	if reloaded.FindByOlaresID(id).BackendVersionRefreshedAt != 2000 {
		t.Error("refreshedAt should update even when version is unchanged")
	}

	// Upgrade detected: 1.12.5 -> 1.12.6 reports changed.
	changed, err = reloaded.SetBackendVersion(id, "1.12.6", 3000)
	if err != nil {
		t.Fatalf("set upgrade: %v", err)
	}
	if !changed {
		t.Error("1.12.5 -> 1.12.6 should report changed=true")
	}

	// Unknown profile errors.
	if _, err := reloaded.SetBackendVersion("bob@olares.com", "1.12.6", 4000); err == nil {
		t.Error("setting version for an unknown profile should error")
	}
}
