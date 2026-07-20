package cluster

import (
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/core/pipeline"
	"github.com/beclab/Olares/cli/pkg/terminus"
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
