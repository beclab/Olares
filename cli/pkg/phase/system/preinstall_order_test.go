package system

import (
	"testing"

	olarescommon "github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/gpu"
	"github.com/beclab/Olares/cli/pkg/images"
	"github.com/beclab/Olares/cli/pkg/manifest"
	"github.com/beclab/Olares/cli/pkg/preinstall"
	"github.com/beclab/Olares/cli/pkg/storage"
)

func TestMacosPhaseBuilderSkipsOfflinePreinstallWiring(t *testing.T) {
	builder := &macOsPhaseBuilder{runtime: &olarescommon.KubeRuntime{}}
	modules := builder.build()
	for _, module := range modules {
		switch module.(type) {
		case *images.PreloadImagesModule:
			t.Fatalf("macOS phase must not include PreloadImagesModule: offline preinstall only supports Linux/WSL")
		case *preinstall.MaterializeModule:
			t.Fatalf("macOS phase must not include MaterializeModule: offline preinstall only supports Linux/WSL")
		}
	}
}

func TestMarketPreinstallModulesFollowImagePreload(t *testing.T) {
	selections := preinstall.ProfileSelections{
		HardwareProfile: gpu.IntelType,
		DetectedGPUType: gpu.IntelType,
	}
	modules := marketPreinstallModules(manifest.InstallationManifest{}, "/installer", "/base", selections)
	if len(modules) != 2 {
		t.Fatalf("module count = %d", len(modules))
	}
	if _, ok := modules[0].(*images.PreloadImagesModule); !ok {
		t.Fatalf("first module = %T", modules[0])
	}
	materialize, ok := modules[1].(*preinstall.MaterializeModule)
	if !ok {
		t.Fatalf("second module = %T", modules[1])
	}
	if materialize.InstallerDir != "/installer" || materialize.RootDir != storage.OlaresRootDir {
		t.Fatalf("materialize paths = %#v", materialize)
	}
	if materialize.ProfileSelections.HardwareProfile != gpu.IntelType ||
		materialize.ProfileSelections.DetectedGPUType != gpu.IntelType {
		t.Fatalf("production selections = %#v", materialize.ProfileSelections)
	}
}

func TestDetectPreinstallGPUTypeUsesMostSpecificHardware(t *testing.T) {
	systemInfo := &connector.SystemInfo{
		HostInfo: &connector.HostInfo{OsType: common.Darwin, OsArch: "arm64"},
		CpuInfo: &connector.CpuInfo{
			IsGB10Chip:      true,
			IsRyzenAIMax:    true,
			IsIntelGPU:      true,
			IsMThreadsM1000: true,
		},
	}

	if got := detectPreinstallGPUType(systemInfo, true); got != gpu.GB10ChipType {
		t.Fatalf("detectPreinstallGPUType() = %q, want %q", got, gpu.GB10ChipType)
	}

	systemInfo.CpuInfo.IsGB10Chip = false
	if got := detectPreinstallGPUType(systemInfo, true); got != gpu.AMDType {
		t.Fatalf("detectPreinstallGPUType() = %q, want %q", got, gpu.AMDType)
	}

	systemInfo.CpuInfo.IsRyzenAIMax = false
	if got := detectPreinstallGPUType(systemInfo, true); got != gpu.IntelType {
		t.Fatalf("detectPreinstallGPUType() = %q, want %q", got, gpu.IntelType)
	}

	systemInfo.CpuInfo.IsIntelGPU = false
	if got := detectPreinstallGPUType(systemInfo, true); got != gpu.MooreSocType {
		t.Fatalf("detectPreinstallGPUType() = %q, want %q", got, gpu.MooreSocType)
	}

	systemInfo.CpuInfo.IsMThreadsM1000 = false
	if got := detectPreinstallGPUType(systemInfo, true); got != gpu.AppleMChipType {
		t.Fatalf("detectPreinstallGPUType() = %q, want %q", got, gpu.AppleMChipType)
	}
}

func TestDetectPreinstallGPUTypeFallsBackToConfiguredNvidiaOrCPU(t *testing.T) {
	systemInfo := &connector.SystemInfo{
		HostInfo: &connector.HostInfo{OsType: common.Linux, OsArch: "amd64"},
		CpuInfo:  &connector.CpuInfo{},
	}

	if got := detectPreinstallGPUType(systemInfo, true); got != gpu.NvidiaCardType {
		t.Fatalf("detectPreinstallGPUType() = %q, want %q", got, gpu.NvidiaCardType)
	}
	if got := detectPreinstallGPUType(systemInfo, false); got != gpu.CPUType {
		t.Fatalf("detectPreinstallGPUType() = %q, want %q", got, gpu.CPUType)
	}
}
