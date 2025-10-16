package pipelines

import (
	"fmt"
	"path"

	"github.com/pkg/errors"

	"github.com/beclab/Olares/cli/cmd/ctl/options"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/phase"
	"github.com/beclab/Olares/cli/pkg/phase/cluster"
)

func CliInstallTerminusPipeline(opts *options.CliTerminusInstallOptions) error {
	var terminusVersion, _ = phase.GetOlaresVersion()
	if terminusVersion != "" {
		return errors.New("Olares is already installed, please uninstall it first.")
	}

	arg := common.NewArgument()
	arg.SetBaseDir(opts.BaseDir)
	arg.SetKubeVersion(opts.KubeType)
	arg.SetOlaresVersion(opts.Version)
	arg.SetMinikubeProfile(opts.MiniKubeProfile)
	arg.SetStorage(getStorageValueFromEnv())
	arg.SetTokenMaxAge()
	arg.SetSwapConfig(opts.SwapConfig)
	if err := arg.SwapConfig.Validate(); err != nil {
		return err
	}
	if opts.WithJuiceFS {
		arg.WithJuiceFS = true
	}
	if opts.EnableReverseProxy != nil {
		arg.NetworkSettings.EnableReverseProxy = opts.EnableReverseProxy
	}
	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return fmt.Errorf("error creating runtime: %v", err)
	}

	manifest := path.Join(runtime.GetInstallerDir(), "installation.manifest")

	runtime.Arg.SetManifest(manifest)

	var p = cluster.InstallSystemPhase(runtime)
	logger.InfoInstallationProgress("Start to Install Olares ...")
	if err := p.Start(); err != nil {
		return err
	}

	return nil
}
