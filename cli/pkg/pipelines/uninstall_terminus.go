package pipelines

import (
	"fmt"
	"os"

	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/phase"
	"bytetrade.io/web3os/installer/pkg/phase/cluster"
)

func UninstallTerminusPipeline(opt *options.CliTerminusUninstallOptions) error {
	terminusVersion := opt.Version
	kubeType := phase.GetKubeType()

	if terminusVersion == "" {
		terminusVersion, _ = phase.GetOlaresVersion()
	}

	var arg = common.NewArgument()
	arg.SetOlaresVersion(terminusVersion)
	arg.SetBaseDir(opt.BaseDir)
	arg.SetConsoleLog("uninstall.log", true)
	arg.SetKubeVersion(kubeType)
	arg.SetDeleteCRI(opt.All || (opt.Phase == cluster.PhasePrepare.String() || opt.Phase == cluster.PhaseDownload.String()))
	arg.SetStorage(&common.Storage{
		StorageVendor: os.Getenv(common.ENV_TERMINUS_IS_CLOUD_VERSION),
		StorageType:   os.Getenv(common.ENV_STORAGE),
		StorageBucket: os.Getenv(common.ENV_S3_BUCKET),
	})

	if err := checkPhase(opt.Phase, opt.All, arg.SystemInfo.GetOsType()); err != nil {
		return err
	}

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return err
	}

	phaseName := opt.Phase
	if opt.All {
		phaseName = cluster.PhaseDownload.String()
	}

	var p = cluster.UninstallTerminus(phaseName, runtime)
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
