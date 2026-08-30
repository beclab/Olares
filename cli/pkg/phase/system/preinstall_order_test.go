package system

import (
	"os"
	"strings"
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
			t.Fatalf("macOS phase must not include PreloadImagesModule: minikube installs load no images locally")
		case *preinstall.PublishDeclarationModule:
			t.Fatalf("macOS phase must not include PublishDeclarationModule: market preinstall only supports Linux")
		}
	}
}

// The two phase builders below are asserted against their source rather than
// their module lists. Neither can be built here: linuxPhaseBuilder.build()
// reads the platform through BaseRuntime.GetSystemInfo(), whose backing field
// is private and reachable only via NewBaseRuntime -- which creates
// directories and a log file -- and wslPhaseBuilder.build() probes the real
// GPU on its first line.

func TestLinuxPhasePreloadsImagesBeforePublishingTheDeclaration(t *testing.T) {
	source := phaseSource(t, "linux.go")
	preload := strings.Index(source, "&images.PreloadImagesModule{")
	publish := strings.Index(source, "addModule(marketPreinstallModules(")
	if preload < 0 || publish < 0 {
		t.Fatalf("build markers: preload=%d publish=%d", preload, publish)
	}
	// Market reads the declaration as soon as it is published, and installing
	// a local chart needs the images already in containerd.
	if preload >= publish {
		t.Fatalf("declaration is published before images are preloaded: preload=%d publish=%d", preload, publish)
	}
}

func TestWslPhasePublishesNoDeclarationButStillPreloadsImages(t *testing.T) {
	source := phaseSource(t, "wsl.go")
	if strings.Contains(source, "marketPreinstallModules(") {
		t.Fatalf("WSL phase must not publish a preinstall declaration: market preinstall only supports Linux")
	}
	// Preloading is not part of preinstall: an offline WSL install needs the
	// system's own images in containerd whether or not anything is
	// preinstalled.
	if !strings.Contains(source, "&images.PreloadImagesModule{") {
		t.Fatalf("WSL phase must still preload images: offline installs have no registry to pull from")
	}
}

func phaseSource(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
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
	materialize, ok := modules[0].(*preinstall.PublishDeclarationModule)
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
