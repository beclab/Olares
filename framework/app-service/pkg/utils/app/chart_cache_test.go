package app

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
)

type tarEntry struct {
	name     string
	body     string
	typeflag byte
	linkname string
}

func makeTgz(t *testing.T, entries ...tarEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, e := range entries {
		tf := e.typeflag
		if tf == 0 {
			tf = tar.TypeReg
		}
		hdr := &tar.Header{Name: e.name, Mode: 0o644, Typeflag: tf, Linkname: e.linkname, Size: int64(len(e.body))}
		if tf != tar.TypeReg {
			hdr.Size = 0
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if tf == tar.TypeReg {
			if _, err := tw.Write([]byte(e.body)); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func chartTgz(t *testing.T, app, version string) []byte {
	return makeTgz(t,
		tarEntry{name: app + "/Chart.yaml", body: fmt.Sprintf("apiVersion: v2\nname: %s\nversion: %s\n", app, version)},
		tarEntry{name: app + "/OlaresManifest.yaml", body: "olaresManifest.version: '0.10.0'\n"},
		tarEntry{name: app + "/templates/cm.yaml", body: "kind: ConfigMap\n"},
	)
}

func sumHex(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

func opts(app string) *ConfigOptions {
	return &ConfigOptions{App: app, RawAppName: app, Owner: "alice", MarketSource: "market.olares"}
}

func TestStoreChartPackageUsesPerVersionDirectories(t *testing.T) {
	root := t.TempDir()
	v1, v2 := chartTgz(t, "notes", "1.0.0"), chartTgz(t, "notes", "2.0.0")

	p1, err := storeChartPackage(root, opts("notes"), "1.0.0", sumHex(v1), v1)
	if err != nil {
		t.Fatal(err)
	}
	p2, err := storeChartPackage(root, opts("notes"), "2.0.0", sumHex(v2), v2)
	if err != nil {
		t.Fatal(err)
	}
	if p1 == p2 {
		t.Fatalf("two versions share one directory %s", p1)
	}
	want := filepath.Join(root, "market.olares", "alice", "notes", "1.0.0-"+sumHex(v1)[:12], "notes")
	if p1 != want {
		t.Fatalf("chart root = %s, want %s", p1, want)
	}
	// The first download is untouched by the second (the old layout rewrote it).
	if b, err := os.ReadFile(filepath.Join(p1, "Chart.yaml")); err != nil || !strings.Contains(string(b), "1.0.0") {
		t.Fatalf("v1 chart changed or missing: %q %v", b, err)
	}
	// Same bytes again reuse the complete directory.
	again, err := storeChartPackage(root, opts("notes"), "1.0.0", sumHex(v1), v1)
	if err != nil || again != p1 {
		t.Fatalf("reuse = %s, %v; want %s", again, err, p1)
	}
	if entries, _ := os.ReadDir(filepath.Join(root, chartTmpDir)); len(entries) != 0 {
		t.Fatalf("temporary directories left behind: %d", len(entries))
	}
}

func TestStoreChartPackageSeparatesOwnersAndSources(t *testing.T) {
	root := t.TempDir()
	pkg := chartTgz(t, "notes", "1.0.0")
	a := opts("notes")
	b := opts("notes")
	b.Owner = "bob"
	c := opts("notes")
	c.MarketSource = "upload"
	seen := map[string]bool{}
	for _, o := range []*ConfigOptions{a, b, c} {
		p, err := storeChartPackage(root, o, "1.0.0", sumHex(pkg), pkg)
		if err != nil {
			t.Fatal(err)
		}
		if seen[p] {
			t.Fatalf("directory %s shared across owner/source", p)
		}
		seen[p] = true
	}
}

func TestPruneChartVersionsNeedsBothRankAndAge(t *testing.T) {
	appDir := t.TempDir()
	now := time.Now()
	mk := func(name string, age time.Duration) string {
		d := filepath.Join(appDir, name)
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
		m := filepath.Join(d, chartCompleteMarker)
		if err := os.WriteFile(m, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		ts := now.Add(-age)
		_ = os.Chtimes(m, ts, ts)
		return d
	}
	day := 24 * time.Hour
	current := mk("9.0.0-cur", 0)
	recent := []string{mk("8.0.0-a", 1*day), mk("7.0.0-b", 2*day), mk("6.0.0-c", 3*day), mk("5.0.0-d", 4*day)}
	oldButRanked := mk("4.0.0-e", 40*day) // 6th newest, old: pruned
	youngButOutranked := mk("3.0.0-f", 5*day)
	_ = youngButOutranked
	veryOld := mk("1.0.0-g", 90*day)
	incomplete := filepath.Join(appDir, "0.0.1-h")
	_ = os.MkdirAll(incomplete, 0o755)

	pruneChartVersions(appDir, current, 5, 30*day, now)

	for _, d := range append([]string{current, incomplete}, recent...) {
		if _, err := os.Stat(d); err != nil {
			t.Errorf("%s removed, want kept", filepath.Base(d))
		}
	}
	// 3.0.0-f is outside the newest five but used 5 days ago: kept.
	if _, err := os.Stat(youngButOutranked); err != nil {
		t.Errorf("recently used directory pruned")
	}
	for _, d := range []string{oldButRanked, veryOld} {
		if _, err := os.Stat(d); !os.IsNotExist(err) {
			t.Errorf("%s kept, want pruned", filepath.Base(d))
		}
	}
}

func TestStoreChartPackageReuseRefreshesAge(t *testing.T) {
	root := t.TempDir()
	pkg := chartTgz(t, "notes", "1.0.0")
	p, err := storeChartPackage(root, opts("notes"), "1.0.0", sumHex(pkg), pkg)
	if err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(filepath.Dir(p), chartCompleteMarker)
	old := time.Now().Add(-60 * 24 * time.Hour)
	_ = os.Chtimes(marker, old, old)
	if _, err := storeChartPackage(root, opts("notes"), "1.0.0", sumHex(pkg), pkg); err != nil {
		t.Fatal(err)
	}
	fi, _ := os.Stat(marker)
	if time.Since(fi.ModTime()) > time.Minute {
		t.Fatalf("reuse did not refresh the marker age: %v", fi.ModTime())
	}
}

func TestExtractChartArchiveRefusesUnsafeEntries(t *testing.T) {
	cases := map[string][]byte{
		"parent traversal": makeTgz(t, tarEntry{name: "../evil", body: "x"}),
		"absolute path":    makeTgz(t, tarEntry{name: "/etc/evil", body: "x"}),
		"symlink":          makeTgz(t, tarEntry{name: "notes/link", typeflag: tar.TypeSymlink, linkname: "/etc/passwd"}),
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			dst := t.TempDir()
			if err := extractChartArchive(data, dst); err == nil {
				t.Fatal("unsafe archive was extracted")
			}
		})
	}
}

func TestChartRootInFallsBackToSingleTopDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "other-name", "templates"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := chartRootIn(dir, "notes")
	if err != nil || got != filepath.Join(dir, "other-name") {
		t.Fatalf("chartRootIn = %s, %v", got, err)
	}
}

func TestResolveChartURLStaysInRepo(t *testing.T) {
	got, err := resolveChartURL("http://appstore-service:82/", "charts/notes-1.0.0.tgz")
	if err != nil || got != "http://appstore-service:82/charts/notes-1.0.0.tgz" {
		t.Fatalf("resolve = %s, %v", got, err)
	}
	if _, err := resolveChartURL("http://appstore-service:82", "http://evil.example/x.tgz"); err == nil {
		t.Fatal("cross-host chart url accepted")
	}
}

// chartRepoServer serves a static-index.yaml and one chart package, and
// records which paths were requested and with which market headers.
func chartRepoServer(t *testing.T, pkg []byte, app, version string) (*httptest.Server, *atomic.Int32, *atomic.Value) {
	t.Helper()
	var indexHits atomic.Int32
	var lastUser atomic.Value
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastUser.Store(r.Header.Get(constants.MarketUser))
		switch r.URL.Path {
		case "/static-index.yaml":
			indexHits.Add(1)
			fmt.Fprintf(w, "apiVersion: v1\nentries:\n  %s:\n  - apiVersion: v2\n    name: %s\n    version: %s\n    urls:\n    - /charts/%s-%s.tgz\n",
				app, app, version, app, version)
		case fmt.Sprintf("/charts/%s-%s.tgz", app, version):
			_, _ = w.Write(pkg)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, &indexHits, &lastUser
}

func TestGetIndexAndDownloadChartExactReference(t *testing.T) {
	t.Chdir(t.TempDir())
	pkg := chartTgz(t, "notes", "1.2.0")
	srv, indexHits, lastUser := chartRepoServer(t, pkg, "notes", "1.2.0")

	o := opts("notes")
	o.RepoURL = srv.URL + "/"
	o.Version = "1.2.0"
	o.ChartURL = "charts/notes-1.2.0.tgz"
	o.ChartDigest = "sha256:" + sumHex(pkg)

	p, err := GetIndexAndDownloadChart(context.Background(), o)
	if err != nil {
		t.Fatal(err)
	}
	if indexHits.Load() != 0 {
		t.Fatalf("exact reference still fetched static-index.yaml %d time(s)", indexHits.Load())
	}
	if lastUser.Load() != "alice" {
		t.Fatalf("X-Market-User = %v, want alice", lastUser.Load())
	}
	if _, err := os.Stat(filepath.Join(p, "OlaresManifest.yaml")); err != nil {
		t.Fatalf("manifest missing in %s: %v", p, err)
	}
	if !strings.HasPrefix(p, filepath.Join("charts", "market.olares", "alice", "notes")) {
		t.Fatalf("chart root %s is not in the per-version cache", p)
	}
}

func TestGetIndexAndDownloadChartRejectsDigestMismatch(t *testing.T) {
	t.Chdir(t.TempDir())
	pkg := chartTgz(t, "notes", "1.2.0")
	srv, _, _ := chartRepoServer(t, pkg, "notes", "1.2.0")

	o := opts("notes")
	o.RepoURL = srv.URL
	o.Version = "1.2.0"
	o.ChartURL = "charts/notes-1.2.0.tgz"
	o.ChartDigest = "sha256:" + strings.Repeat("0", 64)

	_, err := GetIndexAndDownloadChart(context.Background(), o)
	if !errors.Is(err, ErrChartDigestMismatch) {
		t.Fatalf("err = %v, want digest mismatch", err)
	}
	if _, statErr := os.Stat(filepath.Join("charts", "market.olares")); !os.IsNotExist(statErr) {
		t.Fatal("a mismatched package was written to the cache")
	}
}

func TestGetIndexAndDownloadChartLegacyIndexPath(t *testing.T) {
	t.Chdir(t.TempDir())
	pkg := chartTgz(t, "notes", "1.2.0")
	srv, indexHits, _ := chartRepoServer(t, pkg, "notes", "1.2.0")

	o := opts("notes")
	o.RepoURL = srv.URL
	p, err := GetIndexAndDownloadChart(context.Background(), o)
	if err != nil {
		t.Fatal(err)
	}
	if indexHits.Load() != 1 {
		t.Fatalf("index fetched %d time(s), want 1", indexHits.Load())
	}
	want := filepath.Join("charts", "market.olares", "alice", "notes", "1.2.0-"+sumHex(pkg)[:12], "notes")
	if p != want {
		t.Fatalf("chart root = %s, want %s", p, want)
	}
	// The legacy shared directory is never written any more.
	if _, err := os.Stat(filepath.Join("charts", "notes")); !os.IsNotExist(err) {
		t.Fatal("download still wrote ./charts/<rawAppName>")
	}
}

func TestGetAppConfigReadsChartPathFirst(t *testing.T) {
	o := opts("notes")
	o.ChartPath = "/some/downloaded/root"
	o.NeedDownloadChart = true
	// getAppConfigFromRepo must not download when ChartPath is set; reading
	// a missing path fails on the manifest, not on the network.
	_, gotPath, err := getAppConfigFromRepo(context.Background(), o)
	if gotPath != o.ChartPath {
		t.Fatalf("chart path = %q, want %q (err %v)", gotPath, o.ChartPath, err)
	}
}

func TestGetIndexAndDownloadChartIgnoresMismatchedReference(t *testing.T) {
	t.Chdir(t.TempDir())
	pkg := chartTgz(t, "notes", "1.2.0")
	srv, indexHits, _ := chartRepoServer(t, pkg, "notes", "1.2.0")

	o := opts("notes")
	o.RepoURL = srv.URL
	o.Version = "1.2.0"
	o.ChartURL = "charts/notes-0.9.0.tgz" // names another version
	o.ChartDigest = "sha256:" + strings.Repeat("0", 64)

	if _, err := GetIndexAndDownloadChart(context.Background(), o); err != nil {
		t.Fatalf("mismatched reference should fall back to the index: %v", err)
	}
	if indexHits.Load() != 1 {
		t.Fatalf("index fetched %d time(s), want 1", indexHits.Load())
	}
}

func TestGetIndexAndDownloadChartFallsBackWhenReferenceMissing(t *testing.T) {
	t.Chdir(t.TempDir())
	pkg := chartTgz(t, "notes", "1.2.0")
	var indexHits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/static-index.yaml":
			indexHits.Add(1)
			fmt.Fprint(w, "apiVersion: v1\nentries:\n  notes:\n  - apiVersion: v2\n    name: notes\n    version: 1.2.0\n    urls:\n    - /alt/notes-1.2.0.tgz\n")
		case "/alt/notes-1.2.0.tgz":
			_, _ = w.Write(pkg)
		default:
			http.NotFound(w, r) // the referenced charts/... path is gone
		}
	}))
	t.Cleanup(srv.Close)

	o := opts("notes")
	o.RepoURL = srv.URL
	o.Version = "1.2.0"
	o.ChartURL = "charts/notes-1.2.0.tgz"
	o.ChartDigest = "sha256:" + sumHex(pkg)

	if _, err := GetIndexAndDownloadChart(context.Background(), o); err != nil {
		t.Fatalf("404 on the reference should fall back to the index: %v", err)
	}
	if indexHits.Load() != 1 {
		t.Fatalf("index fetched %d time(s), want 1", indexHits.Load())
	}
}
