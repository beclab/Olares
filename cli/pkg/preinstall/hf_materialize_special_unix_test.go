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
