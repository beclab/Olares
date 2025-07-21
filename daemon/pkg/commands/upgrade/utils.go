package upgrade

import (
	"encoding/json"
	"fmt"
	"github.com/beclab/Olares/daemon/cmd/terminusd/version"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

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
