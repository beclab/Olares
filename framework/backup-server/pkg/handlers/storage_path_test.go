package handlers

import (
	"os"
	"path/filepath"
	"testing"

	"olares.com/backup-server/pkg/constant"
)

func TestParseStoragePathRejectsEscapes(t *testing.T) {
	originalNodeName := constant.NodeName
	constant.NodeName = "node-a"
	t.Cleanup(func() { constant.NodeName = originalNodeName })

	for _, input := range []string{
		"/Files/../../var/run/secrets/token",
		"/Files/a/../b",
		"/Files//Home",
		"/Files\\..\\secret",
		"/tmp/backup",
	} {
		if _, _, err := ParseStoragePath(input); err == nil {
			t.Errorf("ParseStoragePath(%q) succeeded, want rejection", input)
		}
	}
}

func TestResolveStoragePathStaysWithinSelectedRoot(t *testing.T) {
	root := t.TempDir()
	userspace := filepath.Join(root, "userspace")
	appcache := filepath.Join(root, "appcache")
	outside := filepath.Join(root, "outside")
	for _, dir := range []string{userspace, appcache, outside} {
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(userspace, "Home"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(userspace, "escape")); err != nil {
		t.Fatal(err)
	}

	resolved, err := ResolveStoragePath("/Files/Home", userspace, appcache)
	if err != nil {
		t.Fatalf("ResolveStoragePath valid path: %v", err)
	}
	if resolved != filepath.Join(userspace, "Home") {
		t.Fatalf("resolved %q, want %q", resolved, filepath.Join(userspace, "Home"))
	}
	if _, err := ResolveStoragePath("/Files/escape/secret", userspace, appcache); err == nil {
		t.Fatal("ResolveStoragePath followed a symlink outside the userspace root")
	}
}

func TestFormatEndpointRejectsFilesystemTraversal(t *testing.T) {
	root := t.TempDir()
	userspace := filepath.Join(root, "userspace")
	appcache := filepath.Join(root, "appcache")
	for _, dir := range []string{userspace, appcache, filepath.Join(userspace, "Home")} {
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := FormatEndpoint(constant.BackupLocationFileSystem.String(), "", "", "/Files/../../tmp/olares-backups/repo", userspace, appcache); err == nil {
		t.Fatal("FormatEndpoint accepted a filesystem traversal path")
	}

	repository, err := FormatEndpoint(constant.BackupLocationFileSystem.String(), "", "", "/Files/Home/olares-backups/repo", userspace, appcache)
	if err != nil {
		t.Fatalf("FormatEndpoint valid path: %v", err)
	}
	if repository.Endpoint != filepath.Join(userspace, "Home") {
		t.Fatalf("endpoint %q is outside expected userspace path", repository.Endpoint)
	}
}
