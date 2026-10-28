package upgrade

import (
	"context"
	"errors"
	"fmt"
	"io"
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

	// Only the exact target may be used to derive the upgrade plan. The
	// adopt sequence is deliberately reentrant — it runs again after the
	// olaresd it installs restarts this process. Both binaries must describe
	// the selected release, even when a failed later upgrade left a newer
	// binary installed.
	if current, err := installedCLIVersion(); err != nil {
		klog.Warningf("Failed to read the installed olares-cli version: %v, installing anyway", err)
	} else if current.Equal(&target.Version) {
		return newExecutionRes(true, nil), nil
	}

	preDownloadedPath := filepath.Join(commands.TERMINUS_BASE_DIR, "pkg", "components", fmt.Sprintf("olares-cli-v%s", target.Version.Original()))
	if _, err := os.Stat(preDownloadedPath); err != nil {
		klog.Warningf("Failed to find pre-downloaded binary path %s: %v", preDownloadedPath, err)
		return newExecutionRes(false, nil), err
	}

	if err := installCLIFile(ctx, preDownloadedPath, cli.TERMINUS_CLI, &target.Version); err != nil {
		return nil, fmt.Errorf("failed to install olares-cli: %w", err)
	}

	return newExecutionRes(true, nil), nil
}

// Stage and validate on the destination filesystem before replacing the running
// binary. A short copy or wrong archive must leave the old CLI executable.
func installCLIFile(ctx context.Context, source, destination string, target *semver.Version) error {
	src, err := os.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()
	staged, err := os.CreateTemp(filepath.Dir(destination), ".olares-cli-*")
	if err != nil {
		return err
	}
	defer os.Remove(staged.Name())
	defer staged.Close()
	if _, err := io.Copy(staged, src); err != nil {
		return err
	}
	if err := staged.Chmod(0755); err != nil {
		return err
	}
	if err := staged.Sync(); err != nil {
		return err
	}
	if err := staged.Close(); err != nil {
		return err
	}
	v, err := cliVersionAt(ctx, staged.Name())
	if err != nil {
		return fmt.Errorf("validate downloaded CLI: %w", err)
	}
	if !v.Equal(target) {
		return fmt.Errorf("downloaded CLI is %s, expected %s", v, target)
	}
	return os.Rename(staged.Name(), destination)
}

// installedCLIVersion is what `olares-cli -v` on this machine reports, whose
// first line is "olares-cli version <v>".
//
// Only that line is read, and the version is looked for anywhere in it rather
// than at a fixed position, the same way the node-status probe reads it. A
// binary that reworded its own version banner should not silently start
// reinstalling itself on every upgrade.
func installedCLIVersion() (*semver.Version, error) {
	return cliVersionAt(context.Background(), cli.TERMINUS_CLI)
}

func cliVersionAt(ctx context.Context, binary string) (*semver.Version, error) {
	out, err := exec.CommandContext(ctx, binary, "-v").Output()
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
