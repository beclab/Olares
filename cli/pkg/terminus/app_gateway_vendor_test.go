package terminus

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveAppGatewayVendorDir_directVendor(t *testing.T) {
	vendor := t.TempDir()
	marker := "linkerd-viz-values.yaml"
	if err := os.WriteFile(filepath.Join(vendor, marker), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := ResolveAppGatewayVendorDir(vendor, marker)
	if got != vendor {
		t.Fatalf("got %q want %q", got, vendor)
	}
}

func TestResolveAppGatewayVendorDir_installerRoot(t *testing.T) {
	root := t.TempDir()
	marker := "linkerd-viz-values.yaml"
	vals := filepath.Join(root, "wizard", "config", appGatewayVendorDirName, marker)
	if err := os.MkdirAll(filepath.Dir(vals), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(vals, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := ResolveAppGatewayVendorDir(root, marker)
	want := filepath.Join(root, "wizard", "config", appGatewayVendorDirName)
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestResolveAppGatewayVendorDir_installerEnv(t *testing.T) {
	root := t.TempDir()
	marker := "linkerd-values.yaml"
	vals := filepath.Join(root, "wizard", "config", appGatewayVendorDirName, marker)
	if err := os.MkdirAll(filepath.Dir(vals), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(vals, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OLARES_INSTALLER_DIR", root)
	got := ResolveAppGatewayVendorDir("", marker)
	want := filepath.Join(root, "wizard", "config", appGatewayVendorDirName)
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
