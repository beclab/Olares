package upgrade

import (
	"runtime"
	"syscall"
	"testing"
)

func TestReleaseArchIsGOARCH(t *testing.T) {
	if got, want := releaseArch(), runtime.GOARCH; got != want {
		t.Fatalf("releaseArch() = %q, want %q (the old == \"arm\" check never matched arm64)", got, want)
	}
}

func TestDiskSpaceKeepsDownloadVolumeStats(t *testing.T) {
	err := checkDiskSpace("/packages", 100, func(path string, fs *syscall.Statfs_t) error {
		fs.Blocks, fs.Bsize = 1000, 1
		if path == "/packages" {
			fs.Bavail = 900
		} else {
			fs.Bavail = 1
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Kubernetes volume overwrote download volume stats: %v", err)
	}
}
