package upgrade

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/commands"
)

func TestDownloadCLISkipsExactInstalledVersion(t *testing.T) {
	version := semver.MustParse("1.12.7-20260922")
	i := NewDownloadCLI().(*downloadCLI)
	i.installedVersion = func() (*semver.Version, error) { return version, nil }
	res, err := i.Execute(context.Background(), state.UpgradeTarget{
		Version: *version, CliURL: "invalid-url-that-must-not-be-used",
	})
	if err != nil || !res.(ExecutionRes).Finished() {
		t.Fatalf("skip already installed CLI: result=%v error=%v", res, err)
	}
}

func TestDownloadCLIRetriesTruncatedResponseWithoutLeavingPartialArchive(t *testing.T) {
	var archive bytes.Buffer
	gzipWriter := gzip.NewWriter(&archive)
	tarWriter := tar.NewWriter(gzipWriter)
	binary := []byte("cli-binary")
	if err := tarWriter.WriteHeader(&tar.Header{Name: "olares-cli", Mode: 0755, Size: int64(len(binary))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tarWriter.Write(binary); err != nil {
		t.Fatal(err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.Header().Set("Content-Length", "1000")
			_, _ = w.Write([]byte("truncated"))
			return
		}
		_, _ = w.Write(archive.Bytes())
	}))
	defer server.Close()

	oldBaseDir := commands.TERMINUS_BASE_DIR
	commands.TERMINUS_BASE_DIR = t.TempDir()
	t.Cleanup(func() { commands.TERMINUS_BASE_DIR = oldBaseDir })
	version := semver.MustParse("1.12.7-20260922")
	i := NewDownloadCLI().(*downloadCLI)
	i.installedVersion = func() (*semver.Version, error) { return semver.MustParse("1.12.7-20260916"), nil }
	res, err := i.Execute(context.Background(), state.UpgradeTarget{Version: *version, CliURL: server.URL})
	if err != nil || !res.(ExecutionRes).Finished() {
		t.Fatalf("download CLI: result=%v error=%v", res, err)
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("download attempts = %d, want 2", got)
	}
	destDir := filepath.Join(commands.TERMINUS_BASE_DIR, "pkg", "components")
	got, err := os.ReadFile(filepath.Join(destDir, "olares-cli-v"+version.Original()))
	if err != nil || !bytes.Equal(got, binary) {
		t.Fatalf("downloaded binary = %q, error = %v", got, err)
	}
	partials, err := filepath.Glob(filepath.Join(destDir, ".olares-cli-download-*"))
	if err != nil || len(partials) != 0 {
		t.Fatalf("partial archives = %v, error = %v", partials, err)
	}
}
