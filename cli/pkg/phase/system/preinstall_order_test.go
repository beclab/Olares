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

func TestMarketPreinstallModulesCarryWhatThePublishNeeds(t *testing.T) {
	selections := preinstall.ProfileSelections{
		HardwareProfile: gpu.IntelType,
		DetectedGPUType: gpu.IntelType,
	}
	modules := marketPreinstallModules(manifest.InstallationManifest{}, "/installer", "/base", "1.12.7-rc.1", selections)
	if len(modules) != 1 {
		t.Fatalf("module count = %d", len(modules))
	}
	materialize, ok := modules[0].(*preinstall.MaterializeModule)
	if !ok {
		t.Fatalf("module = %T", modules[0])
	}
	// The declaration is published under the Olares root the market chart
	// mounts, not under the installer's base directory.
	if materialize.InstallerDir != "/installer" || materialize.RootDir != storage.OlaresRootDir {
		t.Fatalf("materialize paths = %#v", materialize)
	}
	// The declaration is named after the version being installed, so that
	// version has to reach the module that writes it.
	if materialize.OSVersion != "1.12.7-rc.1" {
		t.Fatalf("materialize osVersion = %q", materialize.OSVersion)
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
	systemInfo.HasAmdGPU = true
	if got := detectPreinstallGPUType(systemInfo, true); got != gpu.AmdGpuType {
		t.Fatalf("detectPreinstallGPUType() = %q, want %q", got, gpu.AmdGpuType)
	}

	systemInfo.HasAmdGPU = false
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
