package pipelines

import (
	"fmt"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/phase"
	"github.com/beclab/Olares/cli/pkg/phase/cluster"
	"github.com/beclab/Olares/cli/version"
)

// UninstallTerminusOptions carries the user-supplied configuration for
// UninstallTerminusPipeline. All flag / env / viper resolution happens
// in the cmd/ctl layer.
type UninstallTerminusOptions struct {
	// Phase is the uninstall phase requested via --phase. Ignored
	// when All is true.
	Phase string
	// All requests a complete uninstall (--all). When true, Phase is
	// internally overridden to "download" so the pipeline tears
	// everything down.
	All bool
	// Storage carries the storage backend parameters used to remove
	// remote artifacts during uninstall.
	Storage *common.Storage
}

func UninstallTerminusPipeline(opts UninstallTerminusOptions) error {
	kubeType := phase.GetKubeType()

	sysversion, _ := phase.GetOlaresVersion()
	if sysversion == "" {
		sysversion = version.VERSION
	}

	var arg = common.NewArgument()
	arg.SetOlaresVersion(sysversion)
	arg.SetConsoleLog("uninstall.log", true)
	arg.SetKubeVersion(kubeType)
	arg.SetStorage(opts.Storage)
	arg.ClearMasterHostConfig()

	uninstallPhase := opts.Phase
	if err := checkPhase(uninstallPhase, opts.All, arg.SystemInfo.GetOsType()); err != nil {
		return err
	}

	runtime, err := common.NewKubeRuntime(*arg)
	if err != nil {
		return err
	}

	if opts.All {
		uninstallPhase = cluster.PhaseDownload.String()
	}

	var p = cluster.UninstallTerminus(uninstallPhase, runtime)
	if err := p.Start(); err != nil {
		logger.Errorf("uninstall Olares failed: %v", err)
		return err
	}

	return nil

}

func checkPhase(phase string, all bool, osType string) error {
	if osType == common.Linux && !all {
		if cluster.UninstallPhaseString(phase).Type() == cluster.PhaseInvalid {
			return fmt.Errorf("Please specify the phase to uninstall, such as --phase install. Supported: install, prepare, download.")
		}
	}
	return nil
}
