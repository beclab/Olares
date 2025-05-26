package cluster

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/gpu"
	"bytetrade.io/web3os/installer/pkg/kubesphere/plugins"
	"bytetrade.io/web3os/installer/pkg/manifest"
	"bytetrade.io/web3os/installer/pkg/storage"
	"bytetrade.io/web3os/installer/pkg/terminus"
)

type linuxInstallPhaseBuilder struct {
	runtime     *common.KubeRuntime
	manifestMap manifest.InstallationManifest
}

func (l *linuxInstallPhaseBuilder) base() phase {
	m := []module.Module{
		&plugins.CopyEmbed{},
		&terminus.CheckPreparedModule{Force: true},
	}

	return m
}

func (l *linuxInstallPhaseBuilder) storage() phase {
	return []module.Module{
		&storage.InstallRedisModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: l.manifestMap,
				BaseDir:  l.runtime.GetBaseDir(),
			},
		},
		&storage.InstallJuiceFsModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: l.manifestMap,
				BaseDir:  l.runtime.GetBaseDir(),
			},
		},
	}
}

func (l *linuxInstallPhaseBuilder) installCluster() phase {
	kubeType := l.runtime.Arg.Kubetype
	if kubeType == common.K3s {
		return NewK3sCreateClusterPhase(l.runtime, l.manifestMap)
	} else {
		return NewCreateClusterPhase(l.runtime, l.manifestMap)
	}
}

func (l *linuxInstallPhaseBuilder) installGpuPlugin() phase {
	var skipGpuPlugin = !l.runtime.Arg.GPU.Enable
	if l.runtime.GetSystemInfo().IsWsl() {
		skipGpuPlugin = false
	}
	return []module.Module{
		&gpu.RestartK3sServiceModule{Skip: !(l.runtime.Arg.Kubetype == common.K3s)},
		&gpu.InstallPluginModule{Skip: skipGpuPlugin},
		&gpu.GetCudaVersionModule{},
	}
}

func (l *linuxInstallPhaseBuilder) installTerminus() phase {
	return []module.Module{
		&terminus.GetNATGatewayIPModule{},
		&terminus.InstallAccountModule{},
		&terminus.InstallSettingsModule{},
		&terminus.InstallOsSystemModule{},
		&terminus.InstallLauncherModule{},
		&terminus.InstallAppsModule{},
	}
}

func (l *linuxInstallPhaseBuilder) installBackup() phase {
	return []module.Module{
		&terminus.InstallVeleroModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: l.manifestMap,
				BaseDir:  l.runtime.GetBaseDir(),
			},
		},
	}
}

func (l *linuxInstallPhaseBuilder) build() []module.Module {
	return l.base().
		addModule(fsModuleBuilder(func() []module.Module {
			return l.storage()
		}).withJuiceFS(l.runtime)...).
		addModule(l.installCluster()...).
		addModule(gpuModuleBuilder(func() []module.Module {
			return l.installGpuPlugin()
		}).withGPU(l.runtime)...).
		addModule(l.installTerminus()...).
		addModule(backupModuleBuilder(func() []module.Module {
			return l.installBackup()
		}).withBackup(l.runtime)...).
		addModule(&terminus.InstalledModule{}).
		addModule(&terminus.WriteReleaseFileModule{}).
		addModule(&terminus.WelcomeModule{})
}
