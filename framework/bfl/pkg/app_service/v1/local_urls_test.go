package app_service

import (
	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"testing"
)

func TestLocalURLGeneratorsWithoutOwnerOrPublicDomain(t *testing.T) {
	single, multi, err := localURLGenerators("machine")
	if err != nil {
		t.Fatal(err)
	}
	if got := single("app", "abc123"); got != "abc123.machine.olares.local" {
		t.Fatal(got)
	}
	entrances := []*appv1.Entrance{{Name: "settings"}, {Name: "files"}}
	domains := []appv1.DefaultThirdLevelDomainConfig{{AppName: "olares-app", EntranceName: "settings", ThirdLevelDomain: "settings"}}
	if got := multi("olares-app", "abc123", 0, entrances, domains); got != "settings.machine.olares.local" {
		t.Fatal(got)
	}
	expected := appv1.ResolveEntranceIDWithDefaultThirdLevelDomainOverride(entrances, 1, "abc123", domains) + ".machine.olares.local"
	if got := multi("olares-app", "abc123", 1, entrances, domains); got != expected {
		t.Fatal(got)
	}
}

func TestLocalURLGeneratorsRejectMissingMachineName(t *testing.T) {
	for _, name := range []string{"", " ", "a.b", "../../machine"} {
		if _, _, err := localURLGenerators(name); err == nil {
			t.Fatalf("expected error for %q", name)
		}
	}
}
