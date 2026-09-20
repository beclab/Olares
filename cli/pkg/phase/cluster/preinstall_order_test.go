package cluster

import (
	"os"
	"strings"
	"testing"

	olarescommon "github.com/beclab/Olares/cli/pkg/common"
	corecommon "github.com/beclab/Olares/cli/pkg/core/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/preinstall"
	"github.com/beclab/Olares/cli/pkg/terminus"
)

// installTerminus reads the platform to decide whether the HF cache is
// materialized, so a builder without one cannot be constructed here.
func runtimeOn(osPlatform string) *olarescommon.KubeRuntime {
	return &olarescommon.KubeRuntime{
		Arg: &olarescommon.Argument{
			SystemInfo: &connector.SystemInfo{
				HostInfo: &connector.HostInfo{
					OsType:     corecommon.Linux,
					OsPlatform: osPlatform,
				},
			},
		},
	}
}

func TestLinuxInstallMaterializesHFCacheAfterCommonBeforeApps(t *testing.T) {
	modules := (&linuxInstallPhaseBuilder{runtime: runtimeOn("")}).installTerminus()
	index := func(match func(any) bool) int {
		for i, module := range modules {
			if match(module) {
				return i
			}
		}
		return -1
	}
	osSystem := index(func(value any) bool {
		_, ok := value.(*terminus.InstallOsSystemModule)
		return ok
	})
	hfCache := index(func(value any) bool {
		_, ok := value.(*preinstall.HFCacheMaterializeModule)
		return ok
	})
	launcher := index(func(value any) bool {
		_, ok := value.(*terminus.InstallLauncherModule)
		return ok
	})
	apps := index(func(value any) bool {
		_, ok := value.(*terminus.InstallAppsModule)
		return ok
	})

	if osSystem < 0 || hfCache < 0 || launcher < 0 || apps < 0 {
		t.Fatalf("module indexes: osSystem=%d hfCache=%d launcher=%d apps=%d", osSystem, hfCache, launcher, apps)
	}
	if !(osSystem < hfCache && hfCache < launcher && hfCache < apps) {
		t.Fatalf("unexpected module order: osSystem=%d hfCache=%d launcher=%d apps=%d", osSystem, hfCache, launcher, apps)
	}
	if modules[hfCache].IsSkip() {
		t.Fatalf("HF cache module is skipped on a plain Linux install: %#v", modules[hfCache])
	}
}

// WSL shares the Linux builder, so the module is wired either way and the
// platform difference is only visible through IsSkip.
func TestWslInstallSkipsHFCacheMaterialize(t *testing.T) {
	modules := (&linuxInstallPhaseBuilder{runtime: runtimeOn(corecommon.WSL)}).installTerminus()
	found := false
	for _, module := range modules {
		hfCache, ok := module.(*preinstall.HFCacheMaterializeModule)
		if !ok {
			continue
		}
		found = true
		if !hfCache.IsSkip() {
			t.Fatalf("WSL install must skip HFCacheMaterializeModule: the system phase publishes no declaration, so the cache has nothing to serve")
		}
	}
	if !found {
		t.Fatalf("HFCacheMaterializeModule missing from WSL install: modules=%#v", modules)
	}
}

func TestMacosInstallTerminusNeverMaterializesHFCache(t *testing.T) {
	modules := (&macosInstallPhaseBuilder{}).installTerminus()
	for _, module := range modules {
		if _, ok := module.(*preinstall.HFCacheMaterializeModule); ok {
			t.Fatalf("macOS installTerminus must not wire HFCacheMaterializeModule: offline preinstall only supports Linux, modules=%#v", modules)
		}
	}
}

func TestLinuxBuildPlacesHFCacheBeforeInstalledModule(t *testing.T) {
	data, err := os.ReadFile("linux.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	hfCache := strings.Index(source, "&preinstall.HFCacheMaterializeModule{")
	installed := strings.Index(source, "addModule(&terminus.InstalledModule{})")
	if hfCache < 0 || installed < 0 {
		t.Fatalf("build markers: hfCache=%d installed=%d", hfCache, installed)
	}
	if hfCache >= installed {
		t.Fatalf("HF cache module appears after InstalledModule: hfCache=%d installed=%d", hfCache, installed)
	}
}
