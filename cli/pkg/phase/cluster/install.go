package cluster

import (
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/core/pipeline"
	"github.com/beclab/Olares/cli/pkg/manifest"
)

func InstallSystemPhase(runtime *common.KubeRuntime) *pipeline.Pipeline {
	var err error
	var manifestMap manifest.InstallationManifest
	si := runtime.GetSystemInfo()
	if !si.IsWindows() {
		manifestMap, err = manifest.ReadAll(runtime.Arg.Manifest)
		if err != nil {
			logger.Fatal(err)
		}
	}

	var m []module.Module

	switch {
	case si.IsWindows():
		m = (&windowsInstallPhaseBuilder{runtime: runtime}).build()
	case si.IsDarwin():
		m = (&macosInstallPhaseBuilder{runtime: runtime, manifestMap: manifestMap}).build()
	default:
		m = (&linuxInstallPhaseBuilder{runtime: runtime, manifestMap: manifestMap}).build()
	}

	return &pipeline.Pipeline{
		Name:    "Install the System",
		Modules: m,
		Runtime: runtime,
	}
}
