package system

import (
	"path"
	"strings"

	cc "github.com/beclab/Olares/cli/pkg/core/common"

	"github.com/beclab/Olares/cli/pkg/daemon"

	"github.com/beclab/Olares/cli/pkg/bootstrap/os"
	"github.com/beclab/Olares/cli/pkg/bootstrap/patch"
	"github.com/beclab/Olares/cli/pkg/bootstrap/precheck"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/container"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/gpu"
	"github.com/beclab/Olares/cli/pkg/images"
	"github.com/beclab/Olares/cli/pkg/k3s"
	"github.com/beclab/Olares/cli/pkg/manifest"
	"github.com/beclab/Olares/cli/pkg/terminus"
)

var _ phaseBuilder = &linuxPhaseBuilder{}

type wslPhaseBuilder struct {
	runtime     *common.KubeRuntime
	manifestMap manifest.InstallationManifest
	baseDir     string
}

func (l *wslPhaseBuilder) base() phase {
	return []module.Module{
		&precheck.RunPrechecksModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: l.manifestMap,
				BaseDir:  l.baseDir,
			},
		},
		&patch.InstallDepsModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: l.manifestMap,
				BaseDir:  l.baseDir,
			},
		},
		&os.ConfigSystemModule{},
	}
}

func (l *wslPhaseBuilder) installContainerModule() []module.Module {
	var isK3s = strings.Contains(l.runtime.Arg.KubernetesVersion, "k3s")
	if isK3s {
		return []module.Module{
			&k3s.InstallContainerModule{
				ManifestModule: manifest.ManifestModule{
					Manifest: l.manifestMap,
					BaseDir:  l.baseDir,
				},
			},
		}
	} else {
		return []module.Module{
			&container.InstallContainerModule{
				ManifestModule: manifest.ManifestModule{
					Manifest: l.manifestMap,
					BaseDir:  l.baseDir,
				},
				NoneCluster: true,
			}, //
		}
	}
}

func (l *wslPhaseBuilder) build() []module.Module {
	var baseDir = l.runtime.GetBaseDir()
	var systemInfo = l.runtime.GetSystemInfo()

	if systemInfo.IsWsl() {
		var wslPackageDir = l.runtime.Arg.GetWslUserPath()
		if wslPackageDir != "" {
			baseDir = path.Join(wslPackageDir, cc.DefaultBaseDir)
		}
	}

	l.baseDir = baseDir

	(&gpu.CheckWslGPU{}).Execute(l.runtime)
	return l.base().
		addModule(l.installContainerModule()...).
		addModule(&terminus.WriteReleaseFileModule{}).
		addModule(gpuModuleBuilder(func() []module.Module {
			return []module.Module{
				// on wsl, only install container toolkit. cuda driver is already installed in windows
				&gpu.InstallContainerToolkitModule{
					ManifestModule: manifest.ManifestModule{
						Manifest: l.manifestMap,
						BaseDir:  l.baseDir,
					},
				},
				&gpu.RestartContainerdModule{},
			}

		}).withGPU(l.runtime)...).
		addModule(&images.PreloadImagesModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: l.manifestMap,
				BaseDir:  l.baseDir,
			},
		}).
		addModule(terminusBoxModuleBuilder(func() []module.Module {
			return []module.Module{
				&daemon.InstallTerminusdBinaryModule{
					ManifestModule: manifest.ManifestModule{
						Manifest: l.manifestMap,
						BaseDir:  l.baseDir,
					},
				},
			}
		}).inBox(l.runtime)...).
		addModule(&terminus.PreparedModule{})
}
