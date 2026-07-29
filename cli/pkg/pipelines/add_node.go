package pipelines

import (
	"context"
	"fmt"
	"path"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/phase/cluster"
	"github.com/beclab/Olares/cli/version"
)

func AddNodePipeline(ctx context.Context) error {
	arg := common.NewArgument()
	if !arg.SystemInfo.IsLinux() {
		return fmt.Errorf("only Linux nodes can be added to an Olares cluster")
	}

	arg.SetOlaresVersion(version.VERSION)
	arg.SetConsoleLog("addnode.log", true)

	if err := arg.MasterHostConfig.Validate(); err != nil {
		return fmt.Errorf("invalid master host config: %w", err)
	}

	runtime, err := common.NewKubeRuntime(*arg)
	if err != nil {
		return fmt.Errorf("error creating runtime: %v", err)
	}

	manifest := path.Join(runtime.GetInstallerDir(), "installation.manifest")
	runtime.Arg.SetManifest(manifest)

	var p = cluster.AddNodePhase(runtime)
	if err := p.Start(ctx); err != nil {
		return err
	}
	return nil
}
