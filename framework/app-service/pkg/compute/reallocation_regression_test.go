package compute

import (
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
)

// TestPickAllocationsExcludesOwnExistingClaim is a regression test for the
// re-allocation bug: when an app already held an exclusive GPU, attaching its
// own allocation as a binding made deviceAvailableMemory report zero free
// memory, so PickAllocations could not re-place the app on the card it already
// owned. A re-install or a manual re-bind therefore failed with "no available
// compute resource" even though replaceAppAllocations is designed to support
// re-allocation. AllocateForInstall and ApplyBindingSelection now attach the
// node view with the app's own allocation excluded (withoutAppAllocations);
// this test pins both the buggy precondition (own binding counted) and the
// fixed behavior (own binding excluded) on the pure PickAllocations path.
func TestPickAllocationsExcludesOwnExistingClaim(t *testing.T) {
	const appName, owner = "stable-diffusion", "alice"
	app := &appcfg.ApplicationConfig{
		AppName:         appName,
		OwnerName:       owner,
		SelectedGpuType: utils.NvidiaCardType,
		Accelerator: []appcfg.ResourceMode{
			resourceMode(utils.NvidiaCardType, "8Gi", "1Gi"),
		},
	}
	req, ok := SelectedRequirement(app)
	if !ok {
		t.Fatalf("expected a selected requirement for the nvidia app")
	}

	own := Allocation{
		AppName:  appName,
		Owner:    owner,
		Mode:     utils.NvidiaCardType,
		NodeName: "nvidia-a",
		DeviceID: "gpu0",
	}
	newNodes := func() []Node {
		return []Node{nvidiaNode("nvidia-a", Device{
			ID:          "gpu0",
			Memory:      16 * gi,
			Health:      deviceHealthYes,
			SupportType: SupportTypeExclusive,
		})}
	}

	// Precondition (the bug): counting the app's own claim makes the single
	// exclusive card look fully occupied, so no placement is found.
	withOwn := newNodes()
	attachBindings(withOwn, []Allocation{own})
	if _, ok := PickAllocations(app, req, withOwn, PressureSnapshot{}); ok {
		t.Fatalf("with the app's own binding attached the exclusive card should look occupied")
	}

	// The fix: excluding the app's own claim lets it re-take its own card.
	withoutOwn := newNodes()
	attachBindings(withoutOwn, withoutAppAllocations([]Allocation{own}, appName, owner))
	picked, ok := PickAllocations(app, req, withoutOwn, PressureSnapshot{})
	if !ok {
		t.Fatalf("re-allocation must succeed once the app's own claim is excluded")
	}
	if len(picked) != 1 || picked[0].NodeName != "nvidia-a" || picked[0].DeviceID != "gpu0" {
		t.Fatalf("expected the app to be re-placed on gpu0, got %#v", picked)
	}

	// A *different* app's exclusive binding must still block the card: the fix
	// only drops the requesting app's own rows, not everyone else's.
	otherNodes := newNodes()
	other := Allocation{AppName: "other", Owner: "bob", NodeName: "nvidia-a", DeviceID: "gpu0"}
	attachBindings(otherNodes, withoutAppAllocations([]Allocation{other}, appName, owner))
	if _, ok := PickAllocations(app, req, otherNodes, PressureSnapshot{}); ok {
		t.Fatalf("a different app's exclusive binding must keep the card occupied")
	}
}

// TestWithoutAppAllocationsDropsOnlyMatchingOwner verifies the helper matches
// on the full (appName, owner) key rather than appName alone, so it never
// frees an unrelated app's or user's claim.
func TestWithoutAppAllocationsDropsOnlyMatchingOwner(t *testing.T) {
	const appName, owner = "stable-diffusion", "alice"
	all := []Allocation{
		{AppName: appName, Owner: owner, DeviceID: "gpu0"},
		{AppName: appName, Owner: "someone-else", DeviceID: "gpu1"},
		{AppName: "other", Owner: owner, DeviceID: "gpu2"},
	}
	kept := withoutAppAllocations(all, appName, owner)
	if len(kept) != 2 {
		t.Fatalf("expected to drop only rows owned by (%s,%s), kept=%#v", appName, owner, kept)
	}
	for _, allocation := range kept {
		if allocation.AppName == appName && allocation.Owner == owner {
			t.Fatalf("withoutAppAllocations leaked an owned row: %#v", allocation)
		}
	}
}
