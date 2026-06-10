package collectlogs

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/fanout"
)

type tarEntry struct {
	name    string
	mode    int64
	modTime time.Time
	content string
}

// writeNodeArchive creates an olares-logs-<ts>.tar.gz inside nodeDir mimicking
// what olares-cli produces on a node.
func writeNodeArchive(t *testing.T, nodeDir string, entries []tarEntry) {
	t.Helper()
	if err := os.MkdirAll(nodeDir, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(nodeDir, "olares-logs-20240101-000000.tar.gz")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gw := gzip.NewWriter(f)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()
	for _, e := range entries {
		hdr := &tar.Header{
			Name:     e.name,
			Mode:     e.mode,
			Size:     int64(len(e.content)),
			ModTime:  e.modTime,
			Typeflag: tar.TypeReg,
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(e.content)); err != nil {
			t.Fatal(err)
		}
	}
}

// readArchive returns the headers (keyed by name) and contents of a tar.gz.
func readArchive(t *testing.T, path string) (map[string]*tar.Header, map[string]string) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	headers := map[string]*tar.Header{}
	contents := map[string]string{}
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		h := *hdr
		headers[hdr.Name] = &h
		data, err := io.ReadAll(tr)
		if err != nil {
			t.Fatal(err)
		}
		contents[hdr.Name] = string(data)
	}
	return headers, contents
}

func TestBuildArchiveFlattensAndPreservesMetadata(t *testing.T) {
	tmp := t.TempDir()
	runDir := filepath.Join(tmp, "staging")
	mtime := time.Unix(1700000000, 0)
	writeNodeArchive(t, filepath.Join(runDir, "master"), []tarEntry{
		{name: "k3s.log", mode: 0600, modTime: mtime, content: "k3s log line"},
		{name: "pods/ns/app.log", mode: 0644, modTime: mtime, content: "pod log"},
	})

	archivePath := filepath.Join(tmp, "final.tar.gz")
	results := []fanout.NodeResult{
		{Node: fanout.NodeTarget{Name: "master", IsMaster: true, IsSelf: true}, Status: fanout.StatusOK},
	}
	if err := buildArchive(archivePath, runDir, "run-1", "alice", time.Now(), results); err != nil {
		t.Fatalf("buildArchive error: %v", err)
	}

	headers, contents := readArchive(t, archivePath)

	if _, ok := headers["collect-report.json"]; !ok {
		t.Errorf("missing collect-report.json")
	}
	k3s, ok := headers["nodes/master/k3s.log"]
	if !ok {
		t.Fatalf("missing flattened nodes/master/k3s.log; got %v", keys(headers))
	}
	if contents["nodes/master/k3s.log"] != "k3s log line" {
		t.Errorf("k3s.log content mismatch: %q", contents["nodes/master/k3s.log"])
	}
	if k3s.Mode&0777 != 0600 {
		t.Errorf("mode not preserved: got %o want 0600", k3s.Mode&0777)
	}
	if k3s.ModTime.Unix() != mtime.Unix() {
		t.Errorf("mtime not preserved: got %v want %v", k3s.ModTime, mtime)
	}
	if _, ok := headers["nodes/master/pods/ns/app.log"]; !ok {
		t.Errorf("missing nested pod log; got %v", keys(headers))
	}
	// No nested tar.gz should leak into the final archive.
	for name := range headers {
		if filepath.Ext(name) == ".gz" {
			t.Errorf("found nested archive in output: %s", name)
		}
	}
}

func TestBuildArchiveRejectsZipSlip(t *testing.T) {
	tmp := t.TempDir()
	runDir := filepath.Join(tmp, "staging")
	writeNodeArchive(t, filepath.Join(runDir, "master"), []tarEntry{
		{name: "../evil.txt", mode: 0644, modTime: time.Unix(1700000000, 0), content: "pwned"},
	})

	results := []fanout.NodeResult{
		{Node: fanout.NodeTarget{Name: "master"}, Status: fanout.StatusOK},
	}
	err := buildArchive(filepath.Join(tmp, "final.tar.gz"), runDir, "run-1", "alice", time.Now(), results)
	if err == nil {
		t.Fatalf("expected zip-slip to be rejected")
	}
}

func TestBuildArchiveEmptyAndFailedNodes(t *testing.T) {
	tmp := t.TempDir()
	runDir := filepath.Join(tmp, "staging")
	// "ok-empty" reported ok but produced no archive.
	if err := os.MkdirAll(filepath.Join(runDir, "ok-empty"), 0755); err != nil {
		t.Fatal(err)
	}

	archivePath := filepath.Join(tmp, "final.tar.gz")
	results := []fanout.NodeResult{
		{Node: fanout.NodeTarget{Name: "ok-empty"}, Status: fanout.StatusOK},
		{Node: fanout.NodeTarget{Name: "down"}, Status: fanout.StatusUnreachable, Err: "connection refused"},
	}
	if err := buildArchive(archivePath, runDir, "run-1", "alice", time.Now(), results); err != nil {
		t.Fatalf("buildArchive error: %v", err)
	}

	headers, contents := readArchive(t, archivePath)
	if _, ok := headers["nodes/ok-empty/empty.txt"]; !ok {
		t.Errorf("missing empty.txt for ok-but-empty node; got %v", keys(headers))
	}
	errTxt, ok := contents["nodes/down/error.txt"]
	if !ok {
		t.Fatalf("missing error.txt for failed node; got %v", keys(headers))
	}
	if want := "connection refused"; !strings.Contains(errTxt, want) {
		t.Errorf("error.txt missing detail %q: %q", want, errTxt)
	}
}

func TestPruneOldArchives(t *testing.T) {
	dir := t.TempDir()
	base := time.Now()
	var paths []string
	for i := 0; i < 8; i++ {
		p := filepath.Join(dir, fmt.Sprintf("olares-logs-2024010%d-000000.tar.gz", i))
		if err := os.WriteFile(p, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
		// Make mtimes strictly increasing so newest are deterministic.
		mt := base.Add(time.Duration(i) * time.Minute)
		if err := os.Chtimes(p, mt, mt); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, p)
	}
	// An unrelated file must never be touched.
	keepFile := filepath.Join(dir, "notes.txt")
	if err := os.WriteFile(keepFile, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	pruneOldArchives(dir, 5)

	remaining, err := filepath.Glob(filepath.Join(dir, "olares-logs-*.tar.gz"))
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining) != 5 {
		t.Fatalf("want 5 archives kept, got %d: %v", len(remaining), remaining)
	}
	// The 5 newest (indices 3..7) must survive; 0..2 pruned.
	for i := 0; i < 3; i++ {
		if _, err := os.Stat(paths[i]); !os.IsNotExist(err) {
			t.Errorf("expected %s pruned", paths[i])
		}
	}
	for i := 3; i < 8; i++ {
		if _, err := os.Stat(paths[i]); err != nil {
			t.Errorf("expected %s kept: %v", paths[i], err)
		}
	}
	if _, err := os.Stat(keepFile); err != nil {
		t.Errorf("unrelated file was removed: %v", err)
	}
}

func keys(m map[string]*tar.Header) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
