package system

import (
	"strings"

	"github.com/beclab/Olares/cli/pkg/amdgpu"
	"github.com/beclab/Olares/cli/pkg/bootstrap/os"
	"github.com/beclab/Olares/cli/pkg/bootstrap/patch"
	"github.com/beclab/Olares/cli/pkg/bootstrap/precheck"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/container"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/daemon"
	"github.com/beclab/Olares/cli/pkg/gpu"
	"github.com/beclab/Olares/cli/pkg/images"
	"github.com/beclab/Olares/cli/pkg/k3s"
	"github.com/beclab/Olares/cli/pkg/manifest"
	"github.com/beclab/Olares/cli/pkg/storage"
	"github.com/beclab/Olares/cli/pkg/terminus"
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
				&amdgpu.InstallAmdRocmModule{},
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
