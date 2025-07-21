package upgrade

import (
	"context"
	"errors"
	"fmt"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
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
	target, ok := p.(state.UpgradeTarget)
	if !ok {
		return nil, errors.New("invalid param")
	}

	preDownloadedPath := filepath.Join(commands.TERMINUS_BASE_DIR, "pkg", "components", fmt.Sprintf("olares-cli-v%s", target.Version.Original()))
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
