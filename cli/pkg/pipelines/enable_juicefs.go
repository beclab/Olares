package pipelines

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/manifest"
	"github.com/beclab/Olares/cli/pkg/phase"
	"github.com/beclab/Olares/cli/pkg/phase/cluster"
	"github.com/beclab/Olares/cli/version"
)

func EnableJuiceFSPipeline(ctx context.Context, stopTimeout, stopCheckInterval time.Duration) error {
	arg := common.NewArgument()
	if !arg.SystemInfo.IsLinux() {
		fmt.Println("error: enabling JuiceFS is only supported on Linux nodes")
		os.Exit(1)
	}

	kubeType := phase.GetKubeType()
	sysVersion, _ := phase.GetOlaresVersion()
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

	p := cluster.EnableJuiceFS(runtime, manifestMap, stopTimeout, stopCheckInterval)
	if err := p.Start(ctx); err != nil {
		logger.Errorf("failed to enable JuiceFS on the master node: %v", err)
		return err
	}
	return nil
}
