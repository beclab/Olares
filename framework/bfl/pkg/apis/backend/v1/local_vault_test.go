package v1

import (
	"encoding/json"
	"testing"

	"bytetrade.io/web3os/bfl/pkg/constants"
)

func TestClusterConnectionURLs(t *testing.T) {
	for _, tc := range []struct{ name, zone, clusterName, vaultURL string }{
		{"owner", "", "owner", "http://vault.owner.olares.local/server"},
		{"owner", "olares.com", "owner.olares.com", "https://vault.owner.olares.com/server"},
		{"owner", "example.org", "owner.example.org", "https://vault.owner.example.org/server"},
		{"owner", "example.org.", "owner.example.org", "https://vault.owner.example.org/server"},
		{"owner", "olares.local", "owner.olares.local", "http://vault.owner.olares.local/server"},
	} {
		name, url, err := clusterConnectionURLs(tc.name, tc.zone, constants.WaitActivateVault)
		if err != nil || name != tc.clusterName || url != tc.vaultURL {
			t.Errorf("clusterConnectionURLs(%q, %q) = %q %q %v", tc.name, tc.zone, name, url, err)
		}
	}
}

func TestClusterConnectionURLsRejectInvalidSettings(t *testing.T) {
	for _, tc := range []struct{ name, zone string }{
		{"", ""}, {"invalid/path", ""}, {"owner@olares.local", ""}, {"bad.local", ""},
		{"owner", "https://example.org"}, {"owner", "example.org/path"},
		{"owner", ".example.org"}, {"owner", "bad zone"}, {"owner", "."},
	} {
		if _, _, err := clusterConnectionURLs(tc.name, tc.zone, constants.WaitActivateVault); err == nil {
			t.Errorf("invalid cluster endpoint input %#v accepted", tc)
		}
	}
}

func TestOlaresInfoUsesClusterNameJSON(t *testing.T) {
	data, err := json.Marshal(OlaresInfo{ClusterName: "owner.example.org", VaultURL: "https://vault.owner.example.org/server"})
	if err != nil {
		t.Fatal(err)
	}
	var info map[string]any
	if err := json.Unmarshal(data, &info); err != nil {
		t.Fatal(err)
	}
	if info["clusterName"] != "owner.example.org" {
		t.Fatalf("missing cluster name: %s", data)
	}
	if _, ok := info["localName"]; ok {
		t.Fatalf("old field still exposed: %s", data)
	}
}

func TestNormalizeLegacyLocalOlaresID(t *testing.T) {
	for _, tc := range []struct{ id, want string }{
		{"owner", "owner"}, {"owner@", "owner"},
		{"member@olares.local", "member"},
		{"member@olares.com", "member@olares.com"},
		{"member@@olares.local", "member@@olares.local"},
		{"@", "@"}, {"", ""},
	} {
		if got := normalizeLocalOlaresID(tc.id); got != tc.want {
			t.Errorf("normalizeLocalOlaresID(%q) = %q, want %q", tc.id, got, tc.want)
		}
	}
}

func TestOlaresInfoVaultURLOnlyWhileWaitingForVault(t *testing.T) {
	statuses := []constants.WizardStatus{
		constants.WaitActivateVault, constants.WaitActivateSystem, constants.SystemActivating,
		constants.WaitActivateNetwork, constants.SystemActivateFailed, constants.NetworkActivating,
		constants.NetworkActivateFailed, constants.WaitResetPassword, constants.Completed, "", "unknown",
	}
	for _, zone := range []string{"", "example.org"} {
		for _, status := range statuses {
			t.Run(zone+"/"+string(status), func(t *testing.T) {
				name, url, err := clusterConnectionURLs("owner", zone, status)
				if err != nil {
					t.Fatal(err)
				}
				data, err := json.Marshal(OlaresInfo{ClusterName: name, VaultURL: url, WizardStatus: status})
				if err != nil {
					t.Fatal(err)
				}
				var info map[string]any
				if err := json.Unmarshal(data, &info); err != nil {
					t.Fatal(err)
				}
				_, hasURL := info["vaultUrl"]
				if hasURL != (status == constants.WaitActivateVault) {
					t.Fatalf("unexpected vaultUrl presence: %s", data)
				}
				if info["clusterName"] != name || name == "" {
					t.Fatalf("clusterName must remain available: %s", data)
				}
			})
		}
	}
}
