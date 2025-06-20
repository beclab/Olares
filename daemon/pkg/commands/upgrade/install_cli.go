package upgrade

import (
	"context"
	"errors"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/daemon/pkg/commands"
	"k8s.io/klog/v2"
	"os"
	"os/exec"
	"path/filepath"
)

type installCLI struct {
	commands.Operation
}

var _ commands.Interface = &installCLI{}

func NewInstallCLI() commands.Interface {
	return &installCLI{
		Operation: commands.Operation{
			Name: commands.InstallCLI,
		},
	}
}

func (i *installCLI) Execute(ctx context.Context, p any) (res any, err error) {
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
		klog.Warningf("Failed to get current olares-cli version: %v, proceeding with installation", err)
	} else {
		if !currentVersion.LessThan(targetVersion) {
			return newExecutionRes(true, nil), nil
		}
	}

	preDownloadedPath := filepath.Join(commands.TERMINUS_BASE_DIR, "pkg", "components", fmt.Sprintf("olares-cli-v%s", version))
	if _, err := os.Stat(preDownloadedPath); err != nil {
		klog.Warningf("Failed to find pre-downloaded binary path %s: %v", preDownloadedPath, err)
		return newExecutionRes(false, nil), err
	}

	cmd := exec.Command("cp", "-f", preDownloadedPath, "/usr/local/bin/olares-cli")
	err = cmd.Run()
	if err != nil {
		klog.Warningf("Failed to install olares-cli: %v", err)
		return newExecutionRes(false, nil), err
	}

	if err := os.Chmod("/usr/local/bin/olares-cli", 0755); err != nil {
		return nil, fmt.Errorf("failed to make olares-cli executable: %v", err)
	}

	return newExecutionRes(true, nil), nil
}
