package system

import (
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/core/pipeline"
	_ "github.com/beclab/Olares/cli/pkg/gpu"
	"github.com/beclab/Olares/cli/pkg/manifest"
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
