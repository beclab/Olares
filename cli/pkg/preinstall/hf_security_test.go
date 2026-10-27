package preinstall

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLockHFCacheRootNeverChangesOwner(t *testing.T) {
	rootPath := filepath.Join(t.TempDir(), "huggingface")
	if err := os.Mkdir(rootPath, 0o755); err != nil {
		t.Fatal(err)
	}
	beforeInfo, err := os.Stat(rootPath)
	if err != nil {
		t.Fatal(err)
	}
	beforeUID, uidKnown := fileOwnerUID(beforeInfo)

	root, err := os.OpenRoot(rootPath)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()

	restore, err := lockHFCacheRoot(root)
	if err != nil {
		t.Fatalf("lockHFCacheRoot() error = %v", err)
	}

	lockedInfo, statErr := os.Stat(rootPath)
	if statErr != nil {
		t.Fatal(statErr)
	}
	if lockedInfo.Mode().Perm()&0o222 != 0 {
		t.Fatalf("locked mode = %o, want no write bits for anyone", lockedInfo.Mode().Perm())
	}
	if uidKnown {
		if lockedUID, ok := fileOwnerUID(lockedInfo); !ok || lockedUID != beforeUID {
			t.Fatalf("owner changed while locked: before=%d locked=%d ok=%v", beforeUID, lockedUID, ok)
		}
	}

	if err := restore(); err != nil {
		t.Fatalf("restore() error = %v", err)
	}
	afterInfo, statErr := os.Stat(rootPath)
	if statErr != nil {
		t.Fatal(statErr)
	}
	if afterInfo.Mode().Perm() != 0o755 {
		t.Fatalf("restored mode = %o, want 755", afterInfo.Mode().Perm())
	}
	if uidKnown {
		if afterUID, ok := fileOwnerUID(afterInfo); !ok || afterUID != beforeUID {
			t.Fatalf("owner changed after restore: before=%d after=%d ok=%v", beforeUID, afterUID, ok)
		}
	}
}

func TestLockHFCacheRootWithOpsUsesChmodOnlyNoChown(t *testing.T) {
	rootPath := filepath.Join(t.TempDir(), "huggingface")
	if err := os.Mkdir(rootPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(rootPath, 0o777); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	var operations []string
	ops := hfRootLockOps{
		chmod: func(file *os.File, mode os.FileMode) error {
			operations = append(operations, fmt.Sprintf("chmod:%o", mode))
			return file.Chmod(mode)
		},
		sync: func(file *os.File) error {
			operations = append(operations, "sync")
			return file.Sync()
		},
	}

	restore, err := lockHFCacheRootWithOps(root, ops)
	if err != nil {
		t.Fatalf("lockHFCacheRootWithOps() error = %v", err)
	}
	wantLock := []string{"chmod:555", "sync"}
	if got := operations[:len(wantLock)]; !reflect.DeepEqual(got, wantLock) {
		t.Fatalf("lock operations = %v, want %v", got, wantLock)
	}
	if err := restore(); err != nil {
		t.Fatalf("restore() error = %v", err)
	}
	wantAll := []string{"chmod:555", "sync", "chmod:755", "sync"}
	if !reflect.DeepEqual(operations, wantAll) {
		t.Fatalf("operations = %v, want %v", operations, wantAll)
	}
}

func TestLockHFCacheRootWithOpsRestoresSafelyWhenLockFails(t *testing.T) {
	rootPath := filepath.Join(t.TempDir(), "huggingface")
	if err := os.Mkdir(rootPath, 0o755); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	var operations []string
	ops := hfRootLockOps{
		chmod: func(file *os.File, mode os.FileMode) error {
			operations = append(operations, fmt.Sprintf("chmod:%o", mode))
			if mode == 0o555 {
				return fmt.Errorf("injected chmod failure")
			}
			return file.Chmod(mode)
		},
		sync: func(file *os.File) error {
			operations = append(operations, "sync")
			return file.Sync()
		},
	}

	restore, err := lockHFCacheRootWithOps(root, ops)

	if err == nil || !strings.Contains(err.Error(), "injected chmod failure") {
		t.Fatalf("lockHFCacheRootWithOps() error = %v", err)
	}
	if restore != nil {
		t.Fatal("restore callback returned after failed lock")
	}
	want := []string{"chmod:555", "chmod:755", "sync"}
	if !reflect.DeepEqual(operations, want) {
		t.Fatalf("operations = %v, want %v", operations, want)
	}
	info, statErr := os.Stat(rootPath)
	if statErr != nil {
		t.Fatal(statErr)
	}
	if info.Mode().Perm() != 0o755 {
		t.Fatalf("restored mode = %o, want 755", info.Mode().Perm())
	}
}

func TestCreateHFStagingSyncsDirectoryAfterMarker(t *testing.T) {
	parentPath := t.TempDir()
	parent, err := os.OpenRoot(parentPath)
	if err != nil {
		t.Fatal(err)
	}
	defer parent.Close()
	stageSynced := false
	parentSynced := false

	name, stage, info, err := createHFStaging(parent, "models--acme--tiny", "acme/tiny", func(root *os.Root) error {
		if _, err := root.Lstat(hfStageMarkerFileName); err != nil {
			return fmt.Errorf("marker missing before directory sync: %w", err)
		}
		stageSynced = true
		return syncRootDirectory(root, ".")
	}, func(root *os.Root) error {
		if !stageSynced {
			return fmt.Errorf("parent synced before staging directory")
		}
		entries, err := fs.ReadDir(root.FS(), ".")
		if err != nil {
			return err
		}
		if len(entries) != 1 {
			return fmt.Errorf("parent entries = %d, want 1", len(entries))
		}
		staging, err := root.OpenRoot(entries[0].Name())
		if err != nil {
			return err
		}
		defer staging.Close()
		if _, err := staging.Lstat(hfStageMarkerFileName); err != nil {
			return fmt.Errorf("marker missing before parent sync: %w", err)
		}
		parentSynced = true
		return syncRootDirectory(root, ".")
	})
	if err != nil {
		t.Fatalf("createHFStaging() error = %v", err)
	}
	if !stageSynced {
		t.Fatal("staging directory was not synced")
	}
	if !parentSynced {
		t.Fatal("parent directory was not synced")
	}
	if err := stage.Close(); err != nil {
		t.Fatal(err)
	}
	if err := removeKnownHFStaging(parent, name, info); err != nil {
		t.Fatal(err)
	}
}
