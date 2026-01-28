package pipelines

import (
	"fmt"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/phase"
	"github.com/beclab/Olares/cli/pkg/phase/cluster"
	"github.com/spf13/viper"
)

func UninstallTerminusPipeline() error {
	version := viper.GetString(common.FlagVersion)
	kubeType := phase.GetKubeType()

	if version == "" {
		version, _ = phase.GetOlaresVersion()
	}

	var arg = common.NewArgument()
	arg.SetOlaresVersion(version)
	arg.SetBaseDir(viper.GetString(common.FlagBaseDir))
	arg.SetConsoleLog("uninstall.log", true)
	arg.SetKubeVersion(kubeType)
	arg.SetStorage(getStorageConfig())

	phase := viper.GetString(common.FlagUninstallPhase)
	all := viper.GetBool(common.FlagUninstallAll)

	if err := checkPhase(phase, all, arg.SystemInfo.GetOsType()); err != nil {
		return err
	}

	runtime, err := common.NewKubeRuntime(*arg)
	if err != nil {
		return err
	}

	if all {
		phase = cluster.PhaseDownload.String()
	}

	var p = cluster.UninstallTerminus(phase, runtime)
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
