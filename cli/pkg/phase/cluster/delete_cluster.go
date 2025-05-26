package cluster

import (
	"bytetrade.io/web3os/installer/pkg/bootstrap/os"
	"bytetrade.io/web3os/installer/pkg/bootstrap/precheck"
	"bytetrade.io/web3os/installer/pkg/certs"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/container"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
	"bytetrade.io/web3os/installer/pkg/daemon"
	"bytetrade.io/web3os/installer/pkg/k3s"
	"bytetrade.io/web3os/installer/pkg/kubernetes"
	"bytetrade.io/web3os/installer/pkg/kubesphere"
	"bytetrade.io/web3os/installer/pkg/storage"
	"bytetrade.io/web3os/installer/pkg/terminus"
)

type UninstallPhaseType int
type UninstallPhaseString string

const (
	PhaseInvalid UninstallPhaseType = iota
	PhaseInstall
	PhaseStorage
	PhasePrepare
	PhaseDownload
)

func (p UninstallPhaseType) String() string {
	switch p {
	case PhaseInvalid:
		return "invalid"
	case PhaseInstall:
		return "install"
	case PhaseStorage:
		return "storage"
	case PhasePrepare:
		return "prepare"
	case PhaseDownload:
		return "download"
	}
	return ""
}

func (s UninstallPhaseString) String() string {
	return string(s)
}

func (s UninstallPhaseString) Type() UninstallPhaseType {
	switch s.String() {
	case PhaseInstall.String():
		return PhaseInstall
	case PhaseStorage.String():
		return PhaseStorage
	case PhasePrepare.String():
		return PhasePrepare
	case PhaseDownload.String():
		return PhaseDownload
	}
	return PhaseInvalid

}

type phaseBuilder struct {
	phase   string
	modules []module.Module
	runtime *common.KubeRuntime
}

func (p *phaseBuilder) convert() UninstallPhaseType {
	return UninstallPhaseString(p.phase).Type()
}

func (p *phaseBuilder) phaseInstall() *phaseBuilder {
	if p.convert() >= PhaseInstall {
		// _ = (&kubesphere.GetKubeType{}).Execute(p.runtime)

		p.modules = []module.Module{
			&precheck.GreetingsModule{},
		}

		if p.runtime.Arg.SystemInfo.IsWsl() {
			p.modules = append(p.modules, &precheck.RemoveChattrModule{})
		}

		if p.runtime.Arg.Storage.StorageType == common.S3 || p.runtime.Arg.Storage.StorageType == common.OSS || p.runtime.Arg.Storage.StorageType == common.COS {
			p.modules = append(p.modules,
				&precheck.GetStorageKeyModule{},
				&storage.RemoveMountModule{},
			)
		}

		switch p.runtime.Cluster.Kubernetes.Type {
		case common.K3s:
			p.modules = append(p.modules, &k3s.DeleteClusterModule{})
		default:
			p.modules = append(p.modules, &kubernetes.ResetClusterModule{}, &kubernetes.UmountKubeModule{})
		}

		p.modules = append(p.modules,
			&certs.UninstallAutoRenewCertsModule{},
			&container.KillContainerdProcessModule{},
			&os.ClearOSEnvironmentModule{},
			&certs.UninstallCertsFilesModule{},
			&storage.DeleteUserDataModule{},
			&terminus.DeleteWizardFilesModule{},
			&storage.RemoveJuiceFSModule{},
			&storage.DeletePhaseFlagModule{
				PhaseFile: common.TerminusStateFileInstalled,
				BaseDir:   p.runtime.GetBaseDir(),
			},
		)
	}
	return p
}

func (p *phaseBuilder) phaseStorage() *phaseBuilder {
	if p.convert() >= PhaseStorage {
		p.modules = append(p.modules, &storage.RemoveStorageModule{})
	}
	return p
}

func (p *phaseBuilder) phasePrepare() *phaseBuilder {
	if p.convert() >= PhasePrepare {
		p.modules = append(p.modules,
			&container.DeleteZfsMountModule{},
			&container.UninstallContainerModule{Skip: p.runtime.Arg.IsCloudInstance},
			&storage.DeleteTerminusDataModule{},
			&storage.DeletePhaseFlagModule{
				PhaseFile: common.TerminusStateFilePrepared,
				BaseDir:   p.runtime.GetBaseDir(),
			},
			&terminus.RemoveReleaseFileModule{},
		)
	}
	return p
}

func (p *phaseBuilder) phaseDownload() *phaseBuilder {
	terminusdAction := &daemon.CheckTerminusdService{}
	err := terminusdAction.Execute()

	if p.convert() >= PhaseDownload {
		if err == nil {
			p.modules = append(p.modules, &daemon.UninstallTerminusdModule{})
		}
		p.modules = append(p.modules,
			&kubesphere.DeleteCacheModule{},
		)

		if p.runtime.Arg.DeleteCache {
			p.modules = append(p.modules, &storage.DeleteCacheModule{
				BaseDir: p.runtime.GetBaseDir(),
			})
		}
	}
	return p
}

func (p *phaseBuilder) phaseMacos() {
	p.modules = []module.Module{
		&precheck.GreetingsModule{},
	}
	if p.convert() >= PhasePrepare {
		p.modules = append(p.modules, &kubesphere.DeleteMinikubeModule{}, &certs.UninstallCertsFilesModule{})
	}
	if p.convert() >= PhaseDownload {
		p.modules = append(p.modules, &kubesphere.DeleteKubeSphereCachesModule{})
		if p.runtime.Arg.DeleteCache {
			p.modules = append(p.modules, &kubesphere.DeleteCacheModule{})
		}
	}
}

func UninstallTerminus(phase string, runtime *common.KubeRuntime) pipeline.Pipeline {
	var builder = &phaseBuilder{
		phase:   phase,
		runtime: runtime,
	}

	var systemInfo = runtime.GetSystemInfo()
	if systemInfo.IsDarwin() {
		builder.phaseMacos()
	} else if systemInfo.IsWindows() {
		builder.modules = (&windowsUninstallPhaseBuilder{runtime: runtime}).build()
	} else {
		builder.
			phaseInstall().
			phaseStorage().
			phasePrepare().
			phaseDownload()
	}

	return pipeline.Pipeline{
		Name:    "Uninstall Olares",
		Runtime: builder.runtime,
		Modules: builder.modules,
	}
}
