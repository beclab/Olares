package upgrade

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/commands"
	"k8s.io/klog/v2"
)

type downloadCLI struct {
	commands.Operation
	installedVersion func() (*semver.Version, error)
}

var _ commands.Interface = &downloadCLI{}

func NewDownloadCLI() commands.Interface {
	return &downloadCLI{
		Operation: commands.Operation{
			Name: commands.DownloadCLI,
		},
		installedVersion: installedCLIVersion,
	}
}

func (i *downloadCLI) Execute(ctx context.Context, p any) (res any, err error) {
	target, ok := p.(state.UpgradeTarget)
	if !ok {
		return nil, errors.New("invalid param")
	}
	// Node preparation is reentrant, and the selected CLI may already have
	// been installed before olaresd restarted. Do not make a fresh CDN request
	// for a binary that installCLI would deliberately leave untouched.
	readVersion := i.installedVersion
	if readVersion == nil {
		readVersion = installedCLIVersion
	}
	if current, err := readVersion(); err == nil && current.Equal(&target.Version) {
		return newExecutionRes(true, nil), nil
	} else if err != nil {
		klog.Warningf("Failed to read the installed olares-cli version: %v, downloading anyway", err)
	}

	arch := releaseArch()

	destDir := filepath.Join(commands.TERMINUS_BASE_DIR, "pkg", "components")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create components directory: %v", err)
	}

	downloadURL := target.CliURL
	if downloadURL == "" {
		downloadURL = fmt.Sprintf("%s/olares-cli-v%s_linux_%s.tar.gz", commands.OLARES_CDN_SERVICE, target.Version.Original(), arch)
	}
	tarFile := filepath.Join(destDir, fmt.Sprintf("olares-cli-v%s.tar.gz", target.Version.Original()))

	var downloadErr error
	for attempt := 0; attempt < 3; attempt++ {
		if downloadErr = downloadFile(ctx, downloadURL, tarFile); downloadErr == nil {
			break
		}
		if attempt == 2 {
			return nil, fmt.Errorf("failed to download olares-cli after 3 attempts: %w", downloadErr)
		}
		timer := time.NewTimer(time.Duration(attempt+1) * time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
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
