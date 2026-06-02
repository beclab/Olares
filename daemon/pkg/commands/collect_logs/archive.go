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
	"strings"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/fanout"
	"github.com/beclab/Olares/daemon/pkg/commands"
)

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
	return filepath.Join(podLogsDir(home), fmt.Sprintf("olares-logs-%s.tar.gz", sanitizeComponent(runID)))
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
		if err := addDirToTar(tw, filepath.Join(runDir, sanitizeComponent(r.Node.Name)), nodePrefix); err != nil {
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

func addDirToTar(tw *tar.Writer, srcDir, destPrefix string) error {
	info, err := os.Stat(srcDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Node reported ok but produced nothing; leave a marker.
			return writeTarBytes(tw, filepath.Join(destPrefix, "empty.txt"), []byte("node reported ok but produced no output\n"))
		}
		return fmt.Errorf("stat %s: %w", srcDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", srcDir)
	}

	return filepath.Walk(srcDir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		src, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open %s: %w", path, err)
		}
		defer src.Close()
		header := &tar.Header{
			Name:    filepath.Join(destPrefix, rel),
			Mode:    0644,
			Size:    fi.Size(),
			ModTime: fi.ModTime(),
		}
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("write header %s: %w", path, err)
		}
		if _, err := io.CopyN(tw, src, header.Size); err != nil {
			return fmt.Errorf("write data %s: %w", path, err)
		}
		return nil
	})
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
