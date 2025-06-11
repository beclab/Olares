package pipelines

import (
	"fmt"
	"os"
	"path"

	"github.com/beclab/Olares/cli/cmd/ctl/options"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/phase/cluster"
	"github.com/pkg/errors"
)

func AddNodePipeline(opts *options.AddNodeOptions) error {
	arg := common.NewArgument()
	if !arg.SystemInfo.IsLinux() {
		fmt.Println("error: Only Linux nodes can be added to an Olares cluster!")
		os.Exit(1)
	}
	arg.SetBaseDir(opts.BaseDir)
	if opts.Version == "" {
		return errors.New("Olares version must be specified")
	}
	arg.SetOlaresVersion(opts.Version)
	if err := arg.LoadMasterHostConfigIfAny(); err != nil {
		return errors.Wrap(err, "failed to load master host config")
	}
	arg.SetMasterHostOverride(opts.MasterHostConfig)
	if err := arg.MasterHostConfig.Validate(); err != nil {
		return fmt.Errorf("invalid master host config: %w", err)
	}
	arg.SetConsoleLog("addnode.log", true)
	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
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
