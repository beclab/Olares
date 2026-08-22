package preinstall

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPublishEnsureAppsIsIdempotent(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, filepath.FromSlash(EnsureRuntimeRelativeDir))
	t.Cleanup(func() { _ = makeWritable(target) })
	for range 2 {
		if err := PublishEnsureApps(root); err != nil {
			t.Fatal(err)
		}
	}
	data, err := os.ReadFile(filepath.Join(target, EnsureAppsFileName))
	if err != nil {
		t.Fatal(err)
	}
	var contract struct {
		SchemaVersion string `json:"schemaVersion"`
		Apps          []struct {
			AppID        string       `json:"appId"`
			AppName      string       `json:"appName"`
			InstallScope InstallScope `json:"installScope"`
			InstallOrder int          `json:"installOrder"`
		} `json:"apps"`
	}
	if err := json.Unmarshal(data, &contract); err != nil {
		t.Fatal(err)
	}
	if contract.SchemaVersion != SupportedSchemaVersion || len(contract.Apps) != 1 {
		t.Fatalf("unexpected ensure apps contract: %+v", contract)
	}
	if app := contract.Apps[0]; app.AppID != "f3395cd5" || app.AppName != "router" ||
		app.InstallScope != InstallScopeShared || app.InstallOrder != 20 {
		t.Fatalf("unexpected ensured app: %+v", app)
	}
	if !EnsureAppsPublished(root) {
		t.Fatal("EnsureAppsPublished returned false")
	}
}
