//go:build darwin || linux

package preinstall

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestMaterializeHFArtifactsRejectsSpecialFile(t *testing.T) {
	installerDir, targetRoot, _, _ := writeHFArtifactFixture(t)
	source := filepath.Join(installerDir, StaticRelativeDir, "artifacts", "tiny", "blobs", "weights")
	if err := os.Remove(source); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(source, 0o600); err != nil {
		t.Skipf("mkfifo unavailable: %v", err)
	}

	err := materializeHFArtifacts(installerDir, targetRoot, nil)

	if err == nil || !strings.Contains(err.Error(), "regular file") {
		t.Fatalf("materializeHFArtifacts() error = %v", err)
	}
	assertNoHFStaging(t, targetRoot)
}

func TestCopyVerifiedRegularFileRejectsHardlink(t *testing.T) {
	sourcePath := t.TempDir()
	original := filepath.Join(sourcePath, "source")
	if err := os.WriteFile(original, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(original, filepath.Join(sourcePath, "alias")); err != nil {
		t.Fatal(err)
	}
	sourceRoot, err := os.OpenRoot(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	defer sourceRoot.Close()
	targetRoot, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer targetRoot.Close()

	_, err = copyVerifiedRegularFile(sourceRoot, targetRoot, verifiedCopy{
		Source: "source", Target: "target", Size: 1, MaxSize: 1, OutputMode: 0o644,
	})
	if err == nil || !strings.Contains(err.Error(), "hardlink") {
		t.Fatalf("copyVerifiedRegularFile() error = %v, want hardlink rejection", err)
	}
}
