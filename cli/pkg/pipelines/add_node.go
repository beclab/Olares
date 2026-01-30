package pipelines

import (
	"fmt"
	"os"
	"path"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/phase/cluster"
	"github.com/spf13/viper"
)

func AddNodePipeline() error {
	arg := common.NewArgument()
	if !arg.SystemInfo.IsLinux() {
		fmt.Println("error: Only Linux nodes can be added to an Olares cluster!")
		os.Exit(1)
	}

	arg.SetOlaresVersion(viper.GetString(common.FlagVersion))
	arg.SetBaseDir(viper.GetString(common.FlagBaseDir))
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
	if err := p.Start(); err != nil {
		return err
	}
	return nil
}
