package upgrade

import (
	"fmt"
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
	cmd := exec.Command("olaresd", "--version")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute olaresd --version: %v", err)
	}

	// parse version from output
	// expected format: "olaresd version: v${VERSION}"
	parts := strings.Split(string(output), " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("unexpected version output format: %s", string(output))
	}

	version, err := semver.NewVersion(strings.TrimPrefix(strings.TrimSpace(parts[2]), "v"))
	if err != nil {
		return nil, fmt.Errorf("invalid version format: %v", err)
	}

	return version, nil
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
