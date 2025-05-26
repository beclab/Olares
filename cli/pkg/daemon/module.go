package daemon

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/manifest"
	"bytetrade.io/web3os/installer/pkg/terminus"
)

type UninstallTerminusdModule struct {
	common.KubeModule
}

func (u *UninstallTerminusdModule) Init() {
	u.Name = "UninstallOlaresdModule"
	u.Desc = "Uninstall olaresd"

	disableService := &task.RemoteTask{
		Name:     "DisableOlaresdService",
		Desc:     "disable olaresd service",
		Hosts:    u.Runtime.GetHostsByRole(common.K8s),
		Action:   new(DisableTerminusdService),
		Parallel: false,
		Retry:    1,
	}

	uninstall := &task.RemoteTask{
		Name:     "UninstallOlaresd",
		Desc:     "Uninstall olaresd",
		Hosts:    u.Runtime.GetHostsByRole(common.K8s),
		Action:   &UninstallTerminusd{},
		Parallel: false,
		Retry:    1,
	}

	u.Tasks = []task.Interface{
		disableService,
		uninstall,
	}
}

type ReplaceOlaresdBinaryModule struct {
	common.KubeModule
	manifest.ManifestModule
}

func (m *ReplaceOlaresdBinaryModule) Init() {
	m.Name = "ReplaceOlaresdBinaryModule"
	m.Desc = "Replace olaresd"

	replace := &task.LocalTask{
		Name: "ReplaceOlaresdBinary",
		Desc: "Replace olaresd binary",
		Action: &InstallTerminusdBinary{
			ManifestAction: manifest.ManifestAction{
				BaseDir:  m.BaseDir,
				Manifest: m.Manifest,
			},
		},
		Retry: 3,
	}

	updateEnv := &task.LocalTask{
		Name:   "UpdateOlaresdEnv",
		Desc:   "Update olaresd env",
		Action: new(UpdateOlaresdServiceEnv),
	}

	restart := &task.LocalTask{
		Name: "RestartOlaresd",
		Desc: "Restart olaresd",
		Action: &terminus.SystemctlCommand{
			Command:   "restart",
			UnitNames: []string{"olaresd"},
		},
	}

	m.Tasks = []task.Interface{
		replace,
		updateEnv,
		restart,
	}

}

type InstallTerminusdBinaryModule struct {
	common.KubeModule
	manifest.ManifestModule
}

func (i *InstallTerminusdBinaryModule) Init() {
	i.Name = "InstallOlaresdBinaryModule"
	i.Desc = "Install olaresd"

	install := &task.RemoteTask{
		Name:  "InstallOlaresdBinary",
		Desc:  "Install olaresd using binary",
		Hosts: i.Runtime.GetHostsByRole(common.K8s),
		Action: &InstallTerminusdBinary{
			ManifestAction: manifest.ManifestAction{
				BaseDir:  i.BaseDir,
				Manifest: i.Manifest,
			},
		},
		Parallel: false,
		Retry:    1,
	}

	generateEnv := &task.RemoteTask{
		Name:     "GenerateOlaresdEnv",
		Desc:     "Generate olaresd service env",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Action:   new(GenerateTerminusdServiceEnv),
		Parallel: false,
		Retry:    1,
	}

	generateService := &task.RemoteTask{
		Name:     "GenerateOlaresdService",
		Desc:     "Generate olaresd service",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Action:   new(GenerateTerminusdService),
		Parallel: false,
		Retry:    1,
	}

	enableService := &task.RemoteTask{
		Name:     "EnableOlaresdService",
		Desc:     "enable olaresd service",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Action:   new(EnableTerminusdService),
		Parallel: false,
		Retry:    1,
	}

	i.Tasks = []task.Interface{
		install,
		generateEnv,
		generateService,
		enableService,
	}
}
