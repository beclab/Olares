package upgrade

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"bytetrade.io/web3os/terminusd/pkg/commands"
	"github.com/Masterminds/semver/v3"
	"k8s.io/klog/v2"
)

type upgradeCli struct {
	commands.Operation
}

var _ commands.Interface = &upgradeCli{}

func NewUpgradeCli() commands.Interface {
	return &upgradeCli{
		Operation: commands.Operation{
			Name: commands.UpgradeCli,
		},
	}
}

func (i *upgradeCli) Execute(ctx context.Context, p any) (res any, err error) {
	version, ok := p.(string)
	if !ok {
		return nil, errors.New("invalid param")
	}

	targetVersion, err := semver.NewVersion(version)
	if err != nil {
		return nil, fmt.Errorf("invalid target version %s: %v", version, err)
	}

	currentVersion, err := getCurrentCliVersion()
	if err != nil {
		// if we can't get the current version, assume we need to upgrade
		klog.Warningf("Failed to get current olares-cli version: %v, proceeding with upgrade", err)
	} else {
		if !currentVersion.LessThan(targetVersion) {
			return newExecutionRes(true, nil), nil
		}
	}

	arch := "amd64"
	if runtime.GOARCH == "arm" {
		arch = "arm64"
	}

	tmpDir, err := os.MkdirTemp("", "olares-cli-upgrade-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	downloadURL := fmt.Sprintf("%s/olares-cli-v%s_linux_%s.tar.gz", commands.CDN_URL, version, arch)

	tarFile := filepath.Join(tmpDir, "olares-cli.tar.gz")
	if err := downloadFile(downloadURL, tarFile); err != nil {
		return nil, fmt.Errorf("failed to download olares-cli: %v", err)
	}

	if err := extractTarGz(tarFile, tmpDir); err != nil {
		return nil, fmt.Errorf("failed to extract olares-cli: %v", err)
	}

	binaryPath := filepath.Join(tmpDir, "olares-cli")
	if err := os.Rename(binaryPath, "/usr/local/bin/olares-cli"); err != nil {
		return nil, fmt.Errorf("failed to move olares-cli to /usr/local/bin: %v", err)
	}

	if err := os.Chmod("/usr/local/bin/olares-cli", 0755); err != nil {
		return nil, fmt.Errorf("failed to make olares-cli executable: %v", err)
	}

	return newExecutionRes(true, nil), nil
}

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

	version, err := semver.NewVersion(parts[2])
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
