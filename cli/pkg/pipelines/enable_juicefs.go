package pipelines

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/manifest"
	"github.com/beclab/Olares/cli/pkg/phase"
	"github.com/beclab/Olares/cli/pkg/phase/cluster"
	"github.com/beclab/Olares/cli/pkg/storage"
	"github.com/beclab/Olares/cli/version"
)

func EnableJuiceFSPipeline(ctx context.Context) error {
	arg := common.NewArgument()
	if !arg.SystemInfo.IsLinux() {
		fmt.Println("error: enabling JuiceFS is only supported on Linux nodes")
		os.Exit(1)
	}

	// If JuiceFS has already been migrated/enabled on this node, there is
	// nothing left to do. We deliberately do NOT start Olares here: the user
	// manages the Olares lifecycle themselves now. This check runs before the
	// console logger is initialized (SetConsoleLog below), so print directly.
	if storage.IsJuiceFSEnabled() {
		fmt.Println("JuiceFS is already enabled on this node, the rootfs migration is already complete, nothing to do")
		return nil
	}

	kubeType := phase.GetKubeType()
	// The migration requires Olares to be stopped, so we can't rely on
	// GetOlaresVersion() (which queries the cluster via kubectl). Read the
	// installed version from the on-disk marker first; only fall back to the
	// cluster query and then the binary's build version.
	sysVersion := installedOlaresVersion(arg.BaseDir)
	if sysVersion == "" {
		sysVersion, _ = phase.GetOlaresVersion()
	}
	if sysVersion == "" {
		sysVersion = version.VERSION
	}
	arg.SetOlaresVersion(sysVersion)
	arg.SetKubeVersion(kubeType)
	arg.SetStorage(getStorageConfig())
	arg.SetConsoleLog("enablejuicefs.log", true)

	runtime, err := common.NewKubeRuntime(*arg)
	if err != nil {
		return fmt.Errorf("error creating runtime: %v", err)
	}

	manifestPath := path.Join(runtime.GetInstallerDir(), "installation.manifest")
	runtime.Arg.SetManifest(manifestPath)
	manifestMap, err := manifest.ReadAll(runtime.Arg.Manifest)
	if err != nil {
		return fmt.Errorf("error reading installation manifest: %v", err)
	}

	p := cluster.EnableJuiceFS(runtime, manifestMap)
	if err := p.Start(ctx); err != nil {
		logger.Errorf("failed to enable JuiceFS on the master node: %v", err)
		return err
	}
	return nil
}

// installedOlaresVersion reads the Olares version from the on-disk install
// marker (<baseDir>/.installed), whose content is "<version> <kubetype>".
// It returns an empty string if the marker is missing or unreadable.
func installedOlaresVersion(baseDir string) string {
	if baseDir == "" {
		return ""
	}
	data, err := os.ReadFile(path.Join(baseDir, common.TerminusStateFileInstalled))
	if err != nil {
		return ""
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}
