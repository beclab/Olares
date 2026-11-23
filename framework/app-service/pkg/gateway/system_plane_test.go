package gateway

import (
	"reflect"
	"testing"
)

func TestSystemPlaneHostPatternsOwner(t *testing.T) {
	auth := SystemPlaneServices[0]
	got := SystemPlaneHostPatterns(auth, "alice.olares.com", "alice", false)
	want := []string{"auth.alice.olares.com", "auth.alice.olares.local"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("auth hosts=%v want %v", got, want)
	}

	profile := SystemPlaneServices[3]
	got = SystemPlaneHostPatterns(profile, "alice.olares.com", "alice", false)
	want = []string{"alice.olares.com", "alice.olares.local"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("profile hosts=%v want %v", got, want)
	}
}

func TestSystemPlaneHostPatternsEphemeralWizardOnly(t *testing.T) {
	wizard := SystemPlaneServices[2]
	got := SystemPlaneHostPatterns(wizard, "alice.olares.com", "guest1", true)
	want := []string{"wizard-guest1.alice.olares.com", "wizard-guest1.alice.olares.local"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ephemeral wizard hosts=%v want %v", got, want)
	}
	if h := SystemPlaneHostPatterns(SystemPlaneServices[0], "alice.olares.com", "guest1", true); len(h) != 0 {
		t.Fatalf("ephemeral auth should be empty, got %v", h)
	}
	if h := SystemPlaneHostPatterns(SystemPlaneServices[3], "alice.olares.com", "guest1", true); len(h) != 0 {
		t.Fatalf("ephemeral profile should be empty, got %v", h)
	}
}
