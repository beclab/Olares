package upgrade

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	semver "github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/daemon/pkg/cli"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/commands"
	"k8s.io/klog/v2"
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

	// Already at or past the target, so there is nothing to install. The
	// adopt sequence is deliberately reentrant — it runs again after the
	// olaresd it installs restarts this process — and olaresd's own install
	// step has always asked this question. Asking it here too keeps the two
	// halves of the same sequence behaving the same way.
	if current, err := installedCLIVersion(); err != nil {
		klog.Warningf("Failed to read the installed olares-cli version: %v, installing anyway", err)
	} else if !current.LessThan(&target.Version) {
		return newExecutionRes(true, nil), nil
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

// installedCLIVersion is what `olares-cli -v` on this machine reports, whose
// first line is "olares-cli version <v>".
//
// Only that line is read, and the version is looked for anywhere in it rather
// than at a fixed position, the same way the node-status probe reads it. A
// binary that reworded its own version banner should not silently start
// reinstalling itself on every upgrade.
func installedCLIVersion() (*semver.Version, error) {
	out, err := exec.Command(cli.TERMINUS_CLI, "-v").Output()
	if err != nil {
		return nil, err
	}
	first := strings.SplitN(string(out), "\n", 2)[0]
	for _, field := range strings.Fields(first) {
		if v, err := semver.NewVersion(strings.TrimPrefix(field, "v")); err == nil {
			return v, nil
		}
	}
	return nil, fmt.Errorf("no version in %q", strings.TrimSpace(first))
}
