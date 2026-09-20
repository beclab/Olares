package cliconfig

import "testing"

// The managed flag has to survive a write/read cycle: startup import decides
// whether to take over an existing entry by reading it back out of config.json.
func TestManagedFieldsRoundTrip(t *testing.T) {
	t.Setenv(homeEnv, t.TempDir())

	const id = "alice@olares.com"
	seed := &MultiProfileConfig{}
	seed.Upsert(ProfileConfig{OlaresID: id, Managed: true, AppName: "lares"})
	if err := SaveMultiProfileConfig(seed); err != nil {
		t.Fatalf("save: %v", err)
	}

	reloaded, err := LoadMultiProfileConfig()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	p := reloaded.FindByOlaresID(id)
	if p == nil {
		t.Fatal("profile disappeared")
	}
	if !p.Managed || p.AppName != "lares" {
		t.Fatalf("profile = %+v, want managed=true appName=lares", p)
	}
}

// A profile written before these fields existed must keep loading, and must
// not read as managed.
func TestLegacyProfileIsNotManaged(t *testing.T) {
	t.Setenv(homeEnv, t.TempDir())

	seed := &MultiProfileConfig{}
	seed.Upsert(ProfileConfig{OlaresID: "bob@olares.com"})
	if err := SaveMultiProfileConfig(seed); err != nil {
		t.Fatalf("save: %v", err)
	}
	reloaded, err := LoadMultiProfileConfig()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if p := reloaded.FindByOlaresID("bob@olares.com"); p == nil || p.Managed || p.AppName != "" {
		t.Fatalf("profile = %+v, want a plain non-managed entry", p)
	}
}
