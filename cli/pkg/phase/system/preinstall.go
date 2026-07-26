package system

import (
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/gpu"
	"github.com/beclab/Olares/cli/pkg/images"
	"github.com/beclab/Olares/cli/pkg/manifest"
	"github.com/beclab/Olares/cli/pkg/preinstall"
	"github.com/beclab/Olares/cli/pkg/storage"
)

func marketPreinstallModules(
	manifestMap manifest.InstallationManifest,
	installerDir, baseDir string,
	selections preinstall.ProfileSelections,
) []module.Module {
	return []module.Module{
		&images.PreloadImagesModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: manifestMap,
				BaseDir:  baseDir,
			},
		},
		&preinstall.MaterializeModule{
			InstallerDir:      installerDir,
			RootDir:           storage.OlaresRootDir,
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
