package system

import (
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/gpu"
	"github.com/beclab/Olares/cli/pkg/manifest"
	"github.com/beclab/Olares/cli/pkg/preinstall"
	"github.com/beclab/Olares/cli/pkg/storage"
)

func marketPreinstallModules(
	manifestMap manifest.InstallationManifest,
	installerDir, baseDir, osVersion string,
	selections preinstall.ProfileSelections,
) []module.Module {
	return []module.Module{
		&preinstall.PublishDeclarationModule{
			InstallerDir:      installerDir,
			RootDir:           storage.OlaresRootDir,
			OSVersion:         osVersion,
			ProfileSelections: selections,
		},
	}
}

func productionPreinstallSelections(runtime *common.KubeRuntime) preinstall.ProfileSelections {
	nvidiaEnabled := runtime.Arg != nil && runtime.Arg.GPU != nil && runtime.Arg.GPU.Enable
	detectedGPUType := detectPreinstallGPUType(runtime.GetSystemInfo(), nvidiaEnabled)
	return preinstall.ProfileSelections{
		HardwareProfile: detectedGPUType,
		DetectedGPUType: detectedGPUType,
	}
}

func detectPreinstallGPUType(systemInfo connector.Systems, nvidiaEnabled bool) string {
	switch {
	case systemInfo.IsGB10Chip():
		return gpu.GB10ChipType
	case systemInfo.IsRyzenAIMax():
		return gpu.AMDType
	case systemInfo.IsAmdGPU():
		return gpu.AmdGpuType
	case systemInfo.IsIntelGPU():
		return gpu.IntelType
	case systemInfo.IsMThreadsM1000():
		return gpu.MooreSocType
	case systemInfo.IsDarwin() && systemInfo.GetOsArch() == "arm64":
		return gpu.AppleMChipType
	case nvidiaEnabled:
		return gpu.NvidiaCardType
	default:
		return gpu.CPUType
	}
}
