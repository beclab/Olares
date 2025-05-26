package system

import (
	"path"
	"strings"

	cc "bytetrade.io/web3os/installer/pkg/core/common"

	"bytetrade.io/web3os/installer/pkg/daemon"

	"bytetrade.io/web3os/installer/pkg/bootstrap/os"
	"bytetrade.io/web3os/installer/pkg/bootstrap/patch"
	"bytetrade.io/web3os/installer/pkg/bootstrap/precheck"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/container"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/gpu"
	"bytetrade.io/web3os/installer/pkg/images"
	"bytetrade.io/web3os/installer/pkg/k3s"
	"bytetrade.io/web3os/installer/pkg/manifest"
	"bytetrade.io/web3os/installer/pkg/terminus"
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
