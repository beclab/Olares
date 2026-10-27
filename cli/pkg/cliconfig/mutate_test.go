package cliconfig

import (
	"context"
	"sync"
	"testing"
)

// TestMutateProfile_PreservesConcurrentProfileSwitch is the regression test
// for the bug that made `profile use` report success and then silently
// revert: a command holding a MultiProfileConfig loaded before the switch
// wrote its own field and saved the whole struct, restoring the active
// profile as it was at load time.
//
// The shape here mirrors what happened on a real machine — `profile use B`
// lands while a `cluster context --refresh` started under A is in flight —
// and asserts that the switch survives and the refreshed field still lands.
func TestMutateProfile_PreservesConcurrentProfileSwitch(t *testing.T) {
	t.Setenv(homeEnv, t.TempDir())
	ctx := context.Background()

	const (
		idA = "a@olares.com"
		idB = "b@olares.com"
	)
	seed := &MultiProfileConfig{CurrentProfile: idA}
	seed.Upsert(ProfileConfig{OlaresID: idA})
	seed.Upsert(ProfileConfig{OlaresID: idB})
	if err := SaveMultiProfileConfig(seed); err != nil {
		t.Fatalf("seed save: %v", err)
	}

	// The in-flight command's view of the world, captured before the
	// switch happens. Writing through it must not resurrect idA.
	stale, err := LoadMultiProfileConfig()
	if err != nil {
		t.Fatalf("load stale: %v", err)
	}

	// The switch, performed the way `profile use` does it.
	switched, err := LoadMultiProfileConfig()
	if err != nil {
		t.Fatalf("load for switch: %v", err)
	}
	if _, err := switched.SetCurrent(idB); err != nil {
		t.Fatalf("set current: %v", err)
	}
	if err := SaveMultiProfileConfig(switched); err != nil {
		t.Fatalf("save switch: %v", err)
	}

	if _, err := SetOwnerRole(ctx, stale.Current().OlaresID, "owner", 1000); err != nil {
		t.Fatalf("set owner role: %v", err)
	}

	got, err := LoadMultiProfileConfig()
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if got.CurrentProfile != idB {
		t.Errorf("currentProfile = %q, want %q — the cache refresh reverted the switch", got.CurrentProfile, idB)
	}
	if p := got.FindByOlaresID(idA); p == nil || p.OwnerRole != "owner" {
		t.Errorf("profile %s = %+v, want ownerRole=owner — the refresh should still land", idA, p)
	}
}

// TestMutateProfile_ConcurrentWritersDoNotLoseFields checks that parallel
// mutations of different fields on the same profile all survive. Without the
// lock + re-read, the last writer would overwrite the others with the state
// it loaded before they ran.
func TestMutateProfile_ConcurrentWritersDoNotLoseFields(t *testing.T) {
	t.Setenv(homeEnv, t.TempDir())
	ctx := context.Background()

	const id = "alice@olares.com"
	seed := &MultiProfileConfig{CurrentProfile: id}
	seed.Upsert(ProfileConfig{OlaresID: id})
	if err := SaveMultiProfileConfig(seed); err != nil {
		t.Fatalf("seed save: %v", err)
	}

	var wg sync.WaitGroup
	errs := make(chan error, 3)
	for _, write := range []func() error{
		func() error { _, err := SetOwnerRole(ctx, id, "owner", 1000); return err },
		func() error { _, err := SetBackendVersion(ctx, id, "1.12.7", 2000); return err },
		func() error {
			_, err := SetClusterContext(ctx, id, &ClusterContextCache{GlobalRole: "platform-admin"}, 3000)
			return err
		},
	} {
		wg.Add(1)
		go func(fn func() error) {
			defer wg.Done()
			if err := fn(); err != nil {
				errs <- err
			}
		}(write)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent write: %v", err)
	}

	got, err := LoadMultiProfileConfig()
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	p := got.FindByOlaresID(id)
	if p == nil {
		t.Fatal("profile disappeared")
	}
	if p.OwnerRole != "owner" {
		t.Errorf("ownerRole = %q, want owner", p.OwnerRole)
	}
	if p.BackendVersion != "1.12.7" {
		t.Errorf("backendVersion = %q, want 1.12.7", p.BackendVersion)
	}
	if p.ClusterContext == nil || p.ClusterContext.GlobalRole != "platform-admin" {
		t.Errorf("clusterContext = %+v, want globalRole=platform-admin", p.ClusterContext)
	}
	if got.CurrentProfile != id {
		t.Errorf("currentProfile = %q, want %q", got.CurrentProfile, id)
	}
}
