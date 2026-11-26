package terminus

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateUserChartSourcesBeforeCopy(t *testing.T) {
	root := t.TempDir()
	createChart := func(dir string) {
		t.Helper()
		dir = filepath.Join(root, dir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "Chart.yaml"), []byte("apiVersion: v2\nname: fixture\nversion: 0.1.0\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	createChart("launcher")
	if err := validateUserChartSources(root); err == nil {
		t.Fatal("missing apps bundle was accepted")
	}
	createChart("apps/wizard")
	createChart("apps/auth")
	if err := os.Mkdir(filepath.Join(root, "apps", "system-apps"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := validateUserChartSources(root); err == nil {
		t.Fatal("incomplete chart directory was accepted")
	}
	createChart("apps/system-apps")
	if err := validateUserChartSources(root); err != nil {
		t.Fatal(err)
	}
}
