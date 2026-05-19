package terminus

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallLinkerdViz_requiresVendor(t *testing.T) {
	err := InstallLinkerdViz(t.Context(), nil, nil, "", "")
	if err != ErrLinkerdVizVendorNotFound {
		t.Fatalf("got %v want %v", err, ErrLinkerdVizVendorNotFound)
	}
}

func TestLinkerdPrometheusRBACManifest_vendor(t *testing.T) {
	vendor := t.TempDir()
	manifest := filepath.Join(vendor, linkerdPrometheusProxyRBACFile)
	if err := os.MkdirAll(filepath.Dir(manifest), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifest, []byte("apiVersion: v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := linkerdPrometheusRBACManifest(vendor)
	if got != manifest {
		t.Fatalf("got %q want %q", got, manifest)
	}
}
