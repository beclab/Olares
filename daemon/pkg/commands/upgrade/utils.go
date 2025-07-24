package upgrade

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/beclab/Olares/daemon/cmd/terminusd/version"
	"github.com/dustin/go-humanize"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/Masterminds/semver/v3"
)

func getCurrentCliVersion() (*semver.Version, error) {
	cmd := exec.Command("olares-cli", "-v")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute olares-cli -v: %v", err)
	}

	// parse version from output
	// expected format: "olares-cli version ${VERSION}"
	parts := strings.Split(string(output), " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("unexpected version output format: %s", string(output))
	}

	version, err := semver.NewVersion(strings.TrimSpace(parts[2]))
	if err != nil {
		return nil, fmt.Errorf("invalid version format: %v", err)
	}

	return version, nil
}

func getCurrentDaemonVersion() (*semver.Version, error) {
	v, err := semver.NewVersion(*version.RawVersion())
	if err != nil {
		return nil, fmt.Errorf("invalid version of olaresd: %v", err)
	}

	return v, nil
}

func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func extractTarGz(tarFile, destDir string) error {
	cmd := exec.Command("tar", "-xzf", tarFile, "-C", destDir)
	return cmd.Run()
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func unmarshalComponentManifestFile(path string) (map[string]manifestComponent, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(f)
	ret := make(map[string]manifestComponent)
	if err := decoder.Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
}

// checks:
// 1. whether available space is enough
// 2. if the path is at the same partition with K8s root, whether disk space remains more than 10% after use
func tryToUseDiskSpace(path string, size uint64) error {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return fmt.Errorf("failed to statfs path %s: %v", path, err)
	}

	kfs := syscall.Statfs_t{}
	err = syscall.Statfs("/var/lib", &fs)
	if err != nil {
		return fmt.Errorf("failed to statfs K8s root path %s: %v", path, err)
	}

	total := fs.Blocks * uint64(fs.Bsize)
	available := fs.Bavail * uint64(fs.Bsize)

	var ksize uint64
	if fs.Fsid == kfs.Fsid {
		ksize = uint64(float64(total) * 0.11)
	}

	if available > size+ksize {
		return nil
	}

	errStr := fmt.Sprintf("insufficient disk space, available: %s, required: %s", humanize.Bytes(available), humanize.Bytes(size))
	if ksize > 0 {
		errStr += fmt.Sprintf(", reserved for K8s: %s", humanize.Bytes(ksize))
	}

	return errors.New(errStr)

}
