package system

import (
	"bytetrade.io/web3os/installer/pkg/gpu"
	"strings"

	"bytetrade.io/web3os/installer/pkg/bootstrap/os"
	"bytetrade.io/web3os/installer/pkg/bootstrap/patch"
	"bytetrade.io/web3os/installer/pkg/bootstrap/precheck"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/container"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/daemon"
	"bytetrade.io/web3os/installer/pkg/images"
	"bytetrade.io/web3os/installer/pkg/k3s"
	"bytetrade.io/web3os/installer/pkg/manifest"
	"bytetrade.io/web3os/installer/pkg/storage"
	"bytetrade.io/web3os/installer/pkg/terminus"
)

var _ phaseBuilder = &linuxPhaseBuilder{}

type linuxPhaseBuilder struct {
	runtime     *common.KubeRuntime
	manifestMap manifest.InstallationManifest
}

func (l *linuxPhaseBuilder) base() phase {
	m := []module.Module{
		&os.PvePatchModule{Skip: !l.runtime.GetSystemInfo().IsPveOrPveLxc()},
		&precheck.RunPrechecksModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: l.manifestMap,
				BaseDir:  l.runtime.GetBaseDir(), // l.runtime.Arg.BaseDir,
			},
		},
		&patch.InstallDepsModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: l.manifestMap,
				BaseDir:  l.runtime.GetBaseDir(), // l.runtime.Arg.BaseDir,
			},
		},
		&os.ConfigSystemModule{},
	}

	return m
}

func (l *linuxPhaseBuilder) installContainerModule() []module.Module {
	var isK3s = strings.Contains(l.runtime.Arg.KubernetesVersion, "k3s")
	if isK3s {
		return []module.Module{
			&k3s.InstallContainerModule{
				ManifestModule: manifest.ManifestModule{
					Manifest: l.manifestMap,
					BaseDir:  l.runtime.GetBaseDir(), // l.runtime.Arg.BaseDir,
				},
			},
		}
	} else {
		return []module.Module{
			&container.InstallContainerModule{
				ManifestModule: manifest.ManifestModule{
					Manifest: l.manifestMap,
					BaseDir:  l.runtime.GetBaseDir(), // l.runtime.Arg.BaseDir,
				},
				NoneCluster: true,
			}, //
		}
	}
}

func (l *linuxPhaseBuilder) build() []module.Module {
	return l.base().
		addModule(cloudModuleBuilder(func() []module.Module {
			return []module.Module{
				&storage.InitStorageModule{Skip: !l.runtime.Arg.IsCloudInstance},
			}
		}).withCloud(l.runtime)...).
		addModule(cloudModuleBuilder(l.installContainerModule).withoutCloud(l.runtime)...).
		addModule(&terminus.WriteReleaseFileModule{}).
		addModule(gpuModuleBuilder(func() []module.Module {
			return []module.Module{
				&gpu.InstallDriversModule{
					ManifestModule: manifest.ManifestModule{
						Manifest: l.manifestMap,
						BaseDir:  l.runtime.GetBaseDir(), // l.runtime.Arg.BaseDir,
					},
				},
				&gpu.InstallContainerToolkitModule{
					ManifestModule: manifest.ManifestModule{
						Manifest: l.manifestMap,
						BaseDir:  l.runtime.GetBaseDir(), // l.runtime.Arg.BaseDir,
					},
				},
				&gpu.RestartContainerdModule{},
			}

		}).withGPU(l.runtime)...).
		addModule(cloudModuleBuilder(func() []module.Module {
			// unitl now, system ready
			return []module.Module{
				&images.PreloadImagesModule{
					ManifestModule: manifest.ManifestModule{
						Manifest: l.manifestMap,
						BaseDir:  l.runtime.GetBaseDir(), // l.runtime.Arg.BaseDir,
					},
				}, //
			}
		}).withoutCloud(l.runtime)...).
		addModule(terminusBoxModuleBuilder(func() []module.Module {
			return []module.Module{
				&daemon.InstallTerminusdBinaryModule{
					ManifestModule: manifest.ManifestModule{
						Manifest: l.manifestMap,
						BaseDir:  l.runtime.GetBaseDir(), // l.runtime.Arg.BaseDir,
					},
				},
			}
		}).inBox(l.runtime)...).
		addModule(&terminus.PreparedModule{})
}
