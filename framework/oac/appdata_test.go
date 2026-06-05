package oac

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCheckAppDataUsage_PropagatesScannerError guards against silently
// reporting "no reference found" when bufio.Scanner aborts mid-file. The
// previous implementation only consulted scanner.Scan(); a failed read or a
// line longer than bufio.MaxScanTokenSize would exit the loop without an
// error and could let a chart that does reference .Values.userspace.appdata
// past the error point through lint.
func TestCheckAppDataUsage_PropagatesScannerError(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "templates"), 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}
	// A single line larger than bufio.MaxScanTokenSize (64 KiB) reliably
	// produces bufio.ErrTooLong from scanner.Err(). The line itself does
	// not contain the appdata marker, so a correct implementation must
	// neither find a hit nor swallow the scanner error.
	big := bytes.Repeat([]byte("a"), bufio.MaxScanTokenSize+1024)
	if err := os.WriteFile(filepath.Join(dir, "templates", "big.yaml"), big, 0o644); err != nil {
		t.Fatalf("write big.yaml: %v", err)
	}

	err := checkAppDataUsage(dir, stubManifest{})
	if err == nil {
		t.Fatal("expected scanner error to be surfaced, got nil")
	}
	if !errors.Is(err, bufio.ErrTooLong) {
		t.Fatalf("expected bufio.ErrTooLong in chain, got %v", err)
	}
}

// TestCheckAppDataUsage_FindsReference exercises the happy "match found"
// path so the regression test above is paired with a positive assertion of
// the function's intended detection behaviour.
func TestCheckAppDataUsage_FindsReference(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "templates"), 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}
	body := []byte("foo: bar\nmount: {{ .Values.userspace.appdata }}/x\n")
	if err := os.WriteFile(filepath.Join(dir, "templates", "deploy.yaml"), body, 0o644); err != nil {
		t.Fatalf("write deploy.yaml: %v", err)
	}

	err := checkAppDataUsage(dir, stubManifest{})
	if err == nil {
		t.Fatal("expected error reporting missing permission.appData, got nil")
	}
}

// TestCheckAppDataUsage_NoReference makes sure a clean chart returns nil so
// the scanner-error path above isn't trivially passing on every input.
func TestCheckAppDataUsage_NoReference(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "templates"), 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}
	body := []byte("foo: bar\nbaz: qux\n")
	if err := os.WriteFile(filepath.Join(dir, "templates", "ok.yaml"), body, 0o644); err != nil {
		t.Fatalf("write ok.yaml: %v", err)
	}

	if err := checkAppDataUsage(dir, stubManifest{}); err != nil {
		t.Fatalf("expected nil for chart with no appdata reference, got %v", err)
	}
}

// permManifest is a Manifest stub with configurable permission flags for
// template-vs-permission cross-check tests.
type permManifest struct {
	configVersion string
	appData       bool
	appCommon     bool
	externalData  bool
}

func (p permManifest) APIVersion() string            { return "v1" }
func (p permManifest) ConfigVersion() string         { return p.configVersion }
func (p permManifest) ConfigType() string            { return "app" }
func (p permManifest) AppName() string               { return "stub" }
func (p permManifest) AppVersion() string            { return "0.0.0" }
func (p permManifest) Entrances() []EntranceInfo     { return nil }
func (p permManifest) OptionsImages() []string       { return nil }
func (p permManifest) PermissionAppData() bool       { return p.appData }
func (p permManifest) PermissionAppCommon() bool     { return p.appCommon }
func (p permManifest) PermissionExternalData() bool { return p.externalData }
func (p permManifest) Raw() any                      { return nil }

func writeTemplate(t *testing.T, dir, name string, body []byte) {
	t.Helper()
	tmpl := filepath.Join(dir, "templates")
	if err := os.Mkdir(tmpl, 0o755); err != nil && !os.IsExist(err) {
		t.Fatalf("mkdir templates: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpl, name), body, 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func TestCheckAppCommonUsage_DeniedWithoutPermission(t *testing.T) {
	dir := t.TempDir()
	writeTemplate(t, dir, "deploy.yaml", []byte("path: {{ .Values.userspace.appCommon }}/x\n"))

	err := checkAppCommonUsage(dir, permManifest{})
	if err == nil {
		t.Fatal("expected error when appCommon template ref is used without permission.appCommon")
	}
	if !strings.Contains(err.Error(), "permission.appCommon") {
		t.Fatalf("error should mention permission.appCommon, got: %v", err)
	}
}

func TestCheckAppCommonUsage_AllowedWithPermission(t *testing.T) {
	dir := t.TempDir()
	writeTemplate(t, dir, "deploy.yaml", []byte("path: {{ .Values.userspace.appCommon }}/x\n"))

	if err := checkAppCommonUsage(dir, permManifest{appCommon: true}); err != nil {
		t.Fatalf("expected nil when permission.appCommon is true, got %v", err)
	}
}

func TestCheckSharedLibUsage_DeniedOnModernManifest(t *testing.T) {
	dir := t.TempDir()
	writeTemplate(t, dir, "deploy.yaml", []byte("path: {{ .Values.sharedlib }}/x\n"))

	err := checkSharedLibUsage(dir, permManifest{configVersion: "0.12.0"})
	if err == nil {
		t.Fatal("expected error when sharedlib is used without permission.externalData on >= 0.12.0")
	}
	if !strings.Contains(err.Error(), "permission.externalData") {
		t.Fatalf("error should mention permission.externalData, got: %v", err)
	}
}

func TestCheckSharedLibUsage_SkippedOnLegacyManifest(t *testing.T) {
	dir := t.TempDir()
	writeTemplate(t, dir, "deploy.yaml", []byte("path: {{ .Values.sharedlib }}/x\n"))

	if err := checkSharedLibUsage(dir, permManifest{configVersion: "0.11.0"}); err != nil {
		t.Fatalf("sharedlib check must not run below 0.12.0, got: %v", err)
	}
}

func TestCheckSharedLibUsage_AllowedWithPermission(t *testing.T) {
	dir := t.TempDir()
	writeTemplate(t, dir, "deploy.yaml", []byte("path: {{ .Values.sharedlib }}/x\n"))

	if err := checkSharedLibUsage(dir, permManifest{configVersion: "0.12.0", externalData: true}); err != nil {
		t.Fatalf("expected nil when permission.externalData is true, got %v", err)
	}
}
