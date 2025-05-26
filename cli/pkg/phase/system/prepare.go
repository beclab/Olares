package system

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
	_ "bytetrade.io/web3os/installer/pkg/gpu"
	"bytetrade.io/web3os/installer/pkg/manifest"
)

func PrepareSystemPhase(runtime *common.KubeRuntime) *pipeline.Pipeline {
	manifestMap, err := manifest.ReadAll(runtime.Arg.Manifest)
	if err != nil {
		logger.Fatal(err)
	}

	var m []module.Module
	si := runtime.GetSystemInfo()
	switch {
	case si.IsWsl():
		m = (&wslPhaseBuilder{runtime: runtime, manifestMap: manifestMap}).build()
	case si.IsDarwin():
		m = (&macOsPhaseBuilder{runtime: runtime, manifestMap: manifestMap}).build()
	default:
		m = (&linuxPhaseBuilder{runtime: runtime, manifestMap: manifestMap}).build()
	}

	return &pipeline.Pipeline{
		Name:    "Prepare the System Environment",
		Modules: m,
		Runtime: runtime,
	}
}

type phaseBuilder interface {
	build() []module.Module
}
