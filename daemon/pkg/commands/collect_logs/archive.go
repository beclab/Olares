package collectlogs

import (
	"archive/tar"
	"compress/gzip"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/fanout"
	"github.com/beclab/Olares/daemon/pkg/commands"
	"k8s.io/klog/v2"
)

// archiveRetention is the number of most-recent aggregated archives kept per
// user in podLogsDir; older ones are pruned after each successful run.
const archiveRetention = 5

func podLogsDir(home string) string {
	return filepath.Join(home, commands.EXPORT_POD_LOGS_DIR)
}

func stagingRunDir(home, runID string) string {
	return filepath.Join(podLogsDir(home), ".staging", sanitizeComponent(runID))
}

func nodeStagingDir(home, runID, node string) string {
	return filepath.Join(stagingRunDir(home, runID), sanitizeComponent(node))
}

func finalArchivePath(home, runID string) string {
	return filepath.Join(podLogsDir(home), archiveFileName(runID))
}

func archiveFileName(runID string) string {
	return fmt.Sprintf("olares-logs-%s.tar.gz", sanitizeComponent(runID))
}

// finalArchiveRelPath is the archive path relative to the caller's Files root
// (Home), e.g. "Home/pod_logs/olares-logs-<runID>.tar.gz", returned to the
// caller so the frontend can locate the file in Files once collection finishes.
func finalArchiveRelPath(runID string) string {
	return filepath.Join(commands.EXPORT_POD_LOGS_DIR, archiveFileName(runID))
}

func newRunID() string {
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		return time.Now().Format("20060102-150405")
	}
	return time.Now().Format("20060102-150405") + "-" + hex.EncodeToString(buf)
}

// sanitizeComponent keeps a value usable as a single path segment.
func sanitizeComponent(s string) string {
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	s = strings.ReplaceAll(s, "..", "_")
	return s
}

// pruneOldArchives keeps the keep most-recent olares-logs-*.tar.gz in dir and
// removes the rest. Failures are logged, never returned: retention must not
// fail an otherwise successful collection.
func pruneOldArchives(dir string, keep int) {
	if keep <= 0 {
		return
	}
	matches, err := filepath.Glob(filepath.Join(dir, "olares-logs-*.tar.gz"))
	if err != nil {
		klog.Warningf("prune archives: glob %s error: %v", dir, err)
		return
	}
	if len(matches) <= keep {
		return
	}

	type entry struct {
		path    string
		modTime time.Time
	}
	entries := make([]entry, 0, len(matches))
	for _, p := range matches {
		fi, err := os.Stat(p)
		if err != nil {
			klog.Warningf("prune archives: stat %s error: %v", p, err)
			continue
		}
		entries = append(entries, entry{path: p, modTime: fi.ModTime()})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].modTime.After(entries[j].modTime)
	})
	for _, e := range entries[min(keep, len(entries)):] {
		if err := os.Remove(e.path); err != nil {
			klog.Warningf("prune archives: remove %s error: %v", e.path, err)
		}
	}
}

type collectReport struct {
	RunID      string            `json:"runID"`
	Caller     string            `json:"caller"`
	StartedAt  time.Time         `json:"startedAt"`
	FinishedAt time.Time         `json:"finishedAt"`
	Nodes      []nodeReportEntry `json:"nodes"`
}

type nodeReportEntry struct {
	Name     string `json:"name"`
	IP       string `json:"ip"`
	IsSelf   bool   `json:"isSelf"`
	IsMaster bool   `json:"isMaster"`
	Status   string `json:"status"`
	Err      string `json:"err,omitempty"`
}

// buildArchive merges each node's staging output into a single tar.gz. Per-node
// content goes under nodes/<node>/; failed nodes get an error.txt so "this node
// had no data / was unreachable" is visible after extraction. A
// collect-report.json at the root records the authoritative per-node status.
func buildArchive(archivePath, runDir, runID, caller string, startedAt time.Time, results []fanout.NodeResult) error {
	if err := os.MkdirAll(filepath.Dir(archivePath), 0755); err != nil {
		return fmt.Errorf("mkdir archive dir: %w", err)
	}
	f, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("create archive: %w", err)
	}
	defer f.Close()
	gw := gzip.NewWriter(f)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	report := collectReport{
		RunID:     runID,
		Caller:    caller,
		StartedAt: startedAt,
	}

	for _, r := range results {
		report.Nodes = append(report.Nodes, nodeReportEntry{
			Name:     r.Node.Name,
			IP:       r.Node.IP,
			IsSelf:   r.Node.IsSelf,
			IsMaster: r.Node.IsMaster,
			Status:   string(r.Status),
			Err:      r.Err,
		})

		nodePrefix := filepath.Join("nodes", sanitizeComponent(r.Node.Name))
		if r.Status != fanout.StatusOK {
			msg := fmt.Sprintf("status: %s\nerror: %s\n", r.Status, r.Err)
			if err := writeTarBytes(tw, filepath.Join(nodePrefix, "error.txt"), []byte(msg)); err != nil {
				return err
			}
			continue
		}
		if err := mergeNodeArchive(tw, filepath.Join(runDir, sanitizeComponent(r.Node.Name)), nodePrefix); err != nil {
			return err
		}
	}

	report.FinishedAt = time.Now()
	reportBytes, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}
	if err := writeTarBytes(tw, "collect-report.json", reportBytes); err != nil {
		return err
	}
	return nil
}

// mergeNodeArchive streams the single olares-logs-*.tar.gz that olares-cli
// produced in nodeDir into tw under destPrefix, keeping the aggregated archive
// a single layer deep (nodes/<node>/<files>) instead of nesting a tar.gz per
// node. Staging stays compressed on disk; only the in-memory tar stream is
// rewritten.
func mergeNodeArchive(tw *tar.Writer, nodeDir, destPrefix string) error {
	matches, err := filepath.Glob(filepath.Join(nodeDir, "olares-logs-*.tar.gz"))
	if err != nil {
		return fmt.Errorf("glob archive in %s: %w", nodeDir, err)
	}
	if len(matches) == 0 {
		// Node reported ok but produced nothing; leave a marker.
		return writeTarBytes(tw, filepath.Join(destPrefix, "empty.txt"), []byte("node reported ok but produced no output\n"))
	}
	for _, archivePath := range matches {
		if err := streamTarGzInto(tw, archivePath, destPrefix); err != nil {
			return err
		}
	}
	return nil
}

// streamTarGzInto copies every entry of the gzipped tar at archivePath into tw,
// re-rooting names under destPrefix while preserving the original header (mode,
// mtime). Entry names are sanitized to guard against zip-slip.
func streamTarGzInto(tw *tar.Writer, archivePath, destPrefix string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("open archive %s: %w", archivePath, err)
	}
	defer f.Close()
	gr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("gzip reader %s: %w", archivePath, err)
	}
	defer gr.Close()
	tr := tar.NewReader(gr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar %s: %w", archivePath, err)
		}

		clean := filepath.Clean(header.Name)
		if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
			return fmt.Errorf("unsafe tar entry %q in %s", header.Name, archivePath)
		}

		header.Name = filepath.Join(destPrefix, clean)
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("write header %s: %w", header.Name, err)
		}
		if header.Typeflag == tar.TypeReg {
			if _, err := io.CopyN(tw, tr, header.Size); err != nil {
				return fmt.Errorf("write data %s: %w", header.Name, err)
			}
		}
	}
	return nil
}

func writeTarBytes(tw *tar.Writer, name string, data []byte) error {
	header := &tar.Header{
		Name:    name,
		Mode:    0644,
		Size:    int64(len(data)),
		ModTime: time.Now(),
	}
	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("write header %s: %w", name, err)
	}
	if _, err := tw.Write(data); err != nil {
		return fmt.Errorf("write data %s: %w", name, err)
	}
	return nil
}
