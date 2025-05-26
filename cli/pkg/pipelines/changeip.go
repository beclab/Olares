package pipelines

import (
	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/phase"
	"bytetrade.io/web3os/installer/pkg/phase/cluster"
	"fmt"
	"github.com/pkg/errors"
	"net"
)

func ChangeIPPipeline(opt *options.ChangeIPOptions) error {
	terminusVersion := opt.Version
	kubeType := phase.GetKubeType()
	if terminusVersion == "" {
		terminusVersion, _ = phase.GetOlaresVersion()
	}

	var arg = common.NewArgument()
	arg.SetOlaresVersion(terminusVersion)
	arg.SetBaseDir(opt.BaseDir)
	arg.SetConsoleLog("changeip.log", true)
	arg.SetKubeVersion(kubeType)
	arg.SetMinikubeProfile(opt.MinikubeProfile)
	arg.SetWSLDistribution(opt.WSLDistribution)
	if err := arg.LoadMasterHostConfigIfAny(); err != nil {
		return errors.Wrap(err, "failed to load master host config")
	}
	if opt.NewMasterHost != "" {
		if ip := net.ParseIP(opt.NewMasterHost); !util.IsValidIPv4Addr(ip) {
			return fmt.Errorf("master host %s is not a valid IPv4 address", opt.NewMasterHost)
		} else {
			arg.MasterHost = opt.NewMasterHost
		}
	}
	//only run validation if it's a worker node with master host config set
	if arg.MasterHost != "" {
		if err := arg.MasterHostConfig.Validate(); err != nil {
			return fmt.Errorf("invalid master host config: %w", err)
		}
	}

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return err
	}

	var p = cluster.ChangeIP(runtime)
	if err := p.Start(); err != nil {
		logger.Errorf("failed to run change ip pipeline: %v", err)
		return err
	}

	return nil

}
