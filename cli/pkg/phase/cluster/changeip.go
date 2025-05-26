package cluster

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
	"bytetrade.io/web3os/installer/pkg/terminus"
)

func ChangeIP(runtime *common.KubeRuntime) *pipeline.Pipeline {
	var modules []module.Module
	si := runtime.GetSystemInfo()
	if si.IsDarwin() || si.IsWindows() {
		runtime.Arg.HostIP = si.GetLocalIp()
		modules = []module.Module{&terminus.ChangeHostIPModule{}}
	} else {
		logger.Infof("changing the Olares OS IP to %s ...", si.GetLocalIp())
		modules = []module.Module{
			&terminus.CheckPreparedModule{},
			&terminus.CheckInstalledModule{},
			&terminus.ChangeIPModule{},
		}
	}

	return &pipeline.Pipeline{
		Name:    "Change the IP address of Olares OS components",
		Modules: modules,
		Runtime: runtime,
	}
}
