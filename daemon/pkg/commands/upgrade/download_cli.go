package upgrade

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/beclab/Olares/daemon/pkg/cluster/state"

	"github.com/beclab/Olares/daemon/pkg/commands"
)

type downloadCLI struct {
	commands.Operation
}

var _ commands.Interface = &downloadCLI{}

func NewDownloadCLI() commands.Interface {
	return &downloadCLI{
		Operation: commands.Operation{
			Name: commands.DownloadCLI,
		},
	}
}

func (i *downloadCLI) Execute(ctx context.Context, p any) (res any, err error) {
	target, ok := p.(state.UpgradeTarget)
	if !ok {
		return nil, errors.New("invalid param")
	}

	arch := "amd64"
	if runtime.GOARCH == "arm" {
		arch = "arm64"
	}

	destDir := filepath.Join(commands.TERMINUS_BASE_DIR, "pkg", "components")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create components directory: %v", err)
	}

	downloadURL := target.CliURL
	if downloadURL == "" {
		downloadURL = fmt.Sprintf("%s/olares-cli-v%s_linux_%s.tar.gz", commands.OLARES_CDN_SERVICE, target.Version.Original(), arch)
	}
	tarFile := filepath.Join(destDir, fmt.Sprintf("olares-cli-v%s.tar.gz", target.Version.Original()))

	if err := downloadFile(downloadURL, tarFile); err != nil {
		return nil, fmt.Errorf("failed to download olares-cli: %v", err)
	}

	if err := extractTarGz(tarFile, destDir); err != nil {
		return nil, fmt.Errorf("failed to extract olares-cli: %v", err)
	}

	binaryPath := filepath.Join(destDir, "olares-cli")
	versionedPath := filepath.Join(destDir, fmt.Sprintf("olares-cli-v%s", target.Version.Original()))
	if err := os.Rename(binaryPath, versionedPath); err != nil {
		return nil, fmt.Errorf("failed to rename olares-cli binary: %v", err)
	}

	if err := os.Chmod(versionedPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to make olares-cli executable: %v", err)
	}

	os.Remove(tarFile)

	return newExecutionRes(true, nil), nil
}
