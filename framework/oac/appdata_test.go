package oac

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"path/filepath"
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
