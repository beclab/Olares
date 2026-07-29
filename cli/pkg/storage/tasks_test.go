package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDeleteTerminusUserDataPreservesMarketPreinstall(t *testing.T) {
	root := t.TempDir()
	oldUserDataDir := OlaresUserDataDir
	oldJuiceFsCacheDir := JuiceFsCacheDir
	oldJuiceFsRootDir := OlaresJuiceFSRootDir
	oldRedisServiceFile := RedisServiceFile
	t.Cleanup(func() {
		OlaresUserDataDir = oldUserDataDir
		JuiceFsCacheDir = oldJuiceFsCacheDir
		OlaresJuiceFSRootDir = oldJuiceFsRootDir
		RedisServiceFile = oldRedisServiceFile
	})

	OlaresUserDataDir = filepath.Join(root, "userdata")
	JuiceFsCacheDir = filepath.Join(root, "jfscache")
	OlaresJuiceFSRootDir = filepath.Join(root, "rootfs")
	RedisServiceFile = filepath.Join(root, "redis-server.service")

	preinstallFile := filepath.Join(OlaresUserDataDir, "Cache", "market-preinstall", "bundle.json")
	removedFile := filepath.Join(OlaresUserDataDir, "Home", "user-data")
	for _, file := range []string{preinstallFile, removedFile, filepath.Join(JuiceFsCacheDir, "cache")} {
		if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte("test"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if err := (&DeleteTerminusUserData{}).Execute(nil); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if _, err := os.Stat(preinstallFile); err != nil {
		t.Errorf("market preinstall file was not preserved: %v", err)
	}
	if _, err := os.Stat(removedFile); !os.IsNotExist(err) {
		t.Errorf("regular userdata still exists, stat error = %v", err)
	}
	if _, err := os.Stat(JuiceFsCacheDir); !os.IsNotExist(err) {
		t.Errorf("juicefs cache still exists, stat error = %v", err)
	}
}

func TestDeleteTerminusDataRemovesMarketPreinstall(t *testing.T) {
	root := t.TempDir()
	oldOlaresRootDir := OlaresRootDir
	oldOlaresSharedLibDir := OlaresSharedLibDir
	oldStorageDataDir := StorageDataDir
	t.Cleanup(func() {
		OlaresRootDir = oldOlaresRootDir
		OlaresSharedLibDir = oldOlaresSharedLibDir
		StorageDataDir = oldStorageDataDir
	})

	OlaresRootDir = filepath.Join(root, "olares")
	OlaresSharedLibDir = filepath.Join(OlaresRootDir, "share")
	StorageDataDir = filepath.Join(root, "osdata")
	preinstallFile := filepath.Join(OlaresRootDir, "userdata", "Cache", "market-preinstall", "bundle.json")
	if err := os.MkdirAll(filepath.Dir(preinstallFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(preinstallFile, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := (&DeleteTerminusData{}).Execute(nil); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if _, err := os.Stat(OlaresRootDir); !os.IsNotExist(err) {
		t.Errorf("Olares root still exists after full deletion, stat error = %v", err)
	}
}
