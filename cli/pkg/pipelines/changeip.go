package pipelines

import (
	"context"
	"fmt"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/phase"
	"github.com/beclab/Olares/cli/pkg/phase/cluster"
	"github.com/beclab/Olares/cli/version"
	"github.com/spf13/viper"
)

func ChangeIPPipeline() error {
	kubeType := phase.GetKubeType()
	sysversion, _ := phase.GetOlaresVersion()
	if sysversion == "" {
		sysversion = version.VERSION
	}

	var arg = common.NewArgument()
	arg.SetOlaresVersion(sysversion)
	arg.SetConsoleLog("changeip.log", true)
	arg.SetKubeVersion(kubeType)
	arg.SetMinikubeProfile(viper.GetString(common.FlagMiniKubeProfile))
	arg.SetWSLDistribution(viper.GetString(common.FlagWSLDistribution))

	// Validate master host config only if it's a worker node with master host set
	if arg.MasterHost != "" {
		if err := arg.MasterHostConfig.Validate(); err != nil {
			return fmt.Errorf("invalid master host config: %w", err)
		}
	}

	runtime, err := common.NewKubeRuntime(*arg)
	if err != nil {
		return err
	}

	var p = cluster.ChangeIP(runtime)
	// TODO(ctx): plumb ctx in a follow-up; this entry point is not yet ctx-aware.
	if err := p.Start(context.Background()); err != nil {
		logger.Errorf("failed to run change ip pipeline: %v", err)
		return err
	}

	return nil
}
