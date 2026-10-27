package v1

import (
	"encoding/json"
	"testing"

	"bytetrade.io/web3os/bfl/pkg/constants"
)

func TestClusterConnectionURLs(t *testing.T) {
	for _, tc := range []struct{ name, vaultURL string }{
		{"owner", "http://vault.owner.olares.local/server"},
		{"cluster-2", "http://vault.cluster-2.olares.local/server"},
	} {
		name, url, err := clusterConnectionURLs(tc.name, constants.WaitActivateVault)
		if err != nil || name != tc.name || url != tc.vaultURL {
			t.Errorf("clusterConnectionURLs(%q) = %q %q %v", tc.name, name, url, err)
		}
	}
}

func TestClusterConnectionURLsRejectInvalidSettings(t *testing.T) {
	for _, name := range []string{
		"", "invalid/path", "owner@olares.local", "bad.local", "https://example.org", "bad name", ".",
	} {
		if _, _, err := clusterConnectionURLs(name, constants.WaitActivateVault); err == nil {
			t.Errorf("invalid cluster local name %q accepted", name)
		}
	}
}

func TestOlaresInfoUsesClusterNameJSON(t *testing.T) {
	data, err := json.Marshal(OlaresInfo{OlaresID: "member", ClusterName: "owner", VaultURL: "http://vault.owner.olares.local/server"})
	if err != nil {
		t.Fatal(err)
	}
	var info map[string]any
	if err := json.Unmarshal(data, &info); err != nil {
		t.Fatal(err)
	}
	if info["clusterName"] != "owner" {
		t.Fatalf("missing cluster name: %s", data)
	}
	if _, ok := info["localName"]; ok {
		t.Fatalf("old field still exposed: %s", data)
	}
}

func TestOlaresInfoVaultURLOnlyWhileWaitingForVault(t *testing.T) {
	statuses := []constants.WizardStatus{
		constants.WaitActivateVault, constants.WaitActivateSystem, constants.SystemActivating,
		constants.WaitActivateNetwork, constants.SystemActivateFailed, constants.NetworkActivating,
		constants.NetworkActivateFailed, constants.WaitResetPassword, constants.Completed, "", "unknown",
	}
	for _, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			name, url, err := clusterConnectionURLs("owner", status)
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
			if info["clusterName"] != "owner" || name != "owner" {
				t.Fatalf("clusterName must remain available: %s", data)
			}
		})
	}
}
