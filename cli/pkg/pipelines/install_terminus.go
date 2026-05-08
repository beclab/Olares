package pipelines

import (
	"fmt"
	"path"

	"github.com/pkg/errors"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/phase"
	"github.com/beclab/Olares/cli/pkg/phase/cluster"
)

// InstallTerminusOptions carries the user-supplied configuration for
// CliInstallTerminusPipeline. All flag / env / viper resolution happens
// in the cmd/ctl layer; the pipeline itself receives a fully populated
// options value and never reads the global viper registry.
type InstallTerminusOptions struct {
	KubeType        string
	OlaresVersion   string
	MinikubeProfile string
	Storage         *common.Storage
	Swap            common.SwapConfig
	WithJuiceFS     bool

	// EnableReverseProxy is tri-state: nil means "not set, decide at
	// runtime"; non-nil means the user explicitly chose true/false.
	EnableReverseProxy *bool
}

func CliInstallTerminusPipeline(opts InstallTerminusOptions) error {
	var terminusVersion, _ = phase.GetOlaresVersion()
	if terminusVersion != "" {
		return errors.New("Olares is already installed, please uninstall it first.")
	}

	arg := common.NewArgument()
	arg.SetKubeVersion(opts.KubeType)
	arg.SetOlaresVersion(opts.OlaresVersion)
	arg.SetMinikubeProfile(opts.MinikubeProfile)
	arg.SetStorage(opts.Storage)
	arg.SetSwapConfig(opts.Swap)
	if err := arg.SwapConfig.Validate(); err != nil {
		return err
	}
	arg.WithJuiceFS = opts.WithJuiceFS
	if opts.EnableReverseProxy != nil {
		val := *opts.EnableReverseProxy
		arg.NetworkSettings.EnableReverseProxy = &val
	}
	runtime, err := common.NewKubeRuntime(*arg)
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
