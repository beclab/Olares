package pipelines

import (
	"fmt"
	"path"

	"github.com/pkg/errors"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/phase"
	"github.com/beclab/Olares/cli/pkg/phase/cluster"
	"github.com/spf13/viper"
)

func CliInstallTerminusPipeline() error {
	var terminusVersion, _ = phase.GetOlaresVersion()
	if terminusVersion != "" {
		return errors.New("Olares is already installed, please uninstall it first.")
	}

	arg := common.NewArgument()
	arg.SetKubeVersion(viper.GetString(common.FlagKubeType))
	arg.SetOlaresVersion(viper.GetString(common.FlagVersion))
	arg.SetMinikubeProfile(viper.GetString(common.FlagMiniKubeProfile))
	arg.SetStorage(getStorageConfig())
	arg.SetSwapConfig(common.SwapConfig{
		EnablePodSwap:    viper.GetBool(common.FlagEnablePodSwap),
		Swappiness:       viper.GetInt(common.FlagSwappiness),
		EnableZRAM:       viper.GetBool(common.FlagEnableZRAM),
		ZRAMSize:         viper.GetString(common.FlagZRAMSize),
		ZRAMSwapPriority: viper.GetInt(common.FlagZRAMSwapPriority),
	})
	if err := arg.SwapConfig.Validate(); err != nil {
		return err
	}
	arg.WithJuiceFS = viper.GetBool(common.FlagEnableJuiceFS)
	if viper.IsSet(common.FlagEnableReverseProxy) {
		val := viper.GetBool(common.FlagEnableReverseProxy)
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
