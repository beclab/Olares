package windows

import (
	"time"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/task"
)

type InstallWSLModule struct {
	common.KubeModule
}

func (u *InstallWSLModule) Init() {
	u.Name = "InstallWSL"

	downloadAppxPackage := &task.LocalTask{
		Name:   "DownloadAppxPackage",
		Action: &DownloadAppxPackage{},
	}

	installAppxPackage := &task.LocalTask{
		Name:   "InstallAppxPackage",
		Action: &InstallAppxPackage{},
		Retry:  2,
		Delay:  5 * time.Second,
	}

	downloadWslPackage := &task.LocalTask{
		Name:   "DownloadWslInstallPackage",
		Action: &DownloadWSLInstallPackage{},
	}

	updateWSL := &task.LocalTask{
		Name:   "UpdateWSL",
		Action: &UpdateWSL{},
	}

	u.Tasks = []task.Interface{
		downloadAppxPackage,
		installAppxPackage,
		downloadWslPackage,
		updateWSL,
	}
}

type InstallWSLUbuntuDistroModule struct {
	common.KubeModule
}

func (i *InstallWSLUbuntuDistroModule) Init() {
	i.Name = "InstallWSLUbuntuDistro"

	installWSLDistro := &task.LocalTask{
		Name:   "InstallWSLDistro",
		Action: &InstallWSLDistro{},
		Retry:  1,
	}

	i.Tasks = []task.Interface{
		installWSLDistro,
	}
}

// Move the distro to another drive to avoid excessive system disk space usage.
// If using the import method, you can directly specify the location, and it will later be switched to the import method
type MoveDistroModule struct {
	common.KubeModule
}

func (m *MoveDistroModule) Init() {
	m.Name = "MoveDistro"

	m.Tasks = []task.Interface{
		&task.LocalTask{
			Name:   "MoveDistro",
			Action: &MoveDistro{},
		},
	}
}

type ConfigWslModule struct {
	common.KubeModule
}

func (c *ConfigWslModule) Init() {
	c.Name = "ConfigWslConfig"

	configWslConf := &task.LocalTask{
		Name:   "ConfigWslConf",
		Action: &ConfigWslConf{},
	}

	configWSLForwardRules := &task.LocalTask{
		Name:   "ConfigWslConfig",
		Action: &ConfigWSLForwardRules{},
	}

	configWSLHostsAndDns := &task.LocalTask{
		Name:   "ConfigWslHostsAndDns",
		Action: &ConfigWSLHostsAndDns{},
	}

	configWindowsFirewallRule := &task.LocalTask{
		Name:   "ConfigFirewallRule",
		Action: &ConfigWindowsFirewallRule{},
	}

	c.Tasks = []task.Interface{
		configWslConf,
		configWSLForwardRules,
		configWSLHostsAndDns,
		configWindowsFirewallRule,
	}
}

type InstallTerminusModule struct {
	common.KubeModule
}

func (i *InstallTerminusModule) Init() {
	i.Name = "InstallOlares"
	i.Tasks = []task.Interface{
		&task.LocalTask{
			Name:   "InstallOlares",
			Action: &InstallTerminus{},
		},
	}
}

type UninstallOlaresModule struct {
	common.KubeModule
}

func (u *UninstallOlaresModule) Init() {
	u.Name = "UninstallOlares"
	u.Tasks = []task.Interface{
		&task.LocalTask{
			Name:   "UninstallOlares",
			Action: &UninstallOlares{},
		},
		&task.LocalTask{
			Name:   "RemoveFirewallRule",
			Action: &RemoveFirewallRule{},
		},
		&task.LocalTask{
			Name:   "RemovePortProxy",
			Action: &RemovePortProxy{},
		},
	}
}

type GetDiskPartitionModule struct {
	common.KubeModule
}

func (g *GetDiskPartitionModule) Init() {
	g.Name = "GetDiskPartition"

	g.Tasks = []task.Interface{
		&task.LocalTask{
			Name:   "GetDiskPartition",
			Action: &GetDiskPartition{},
		},
	}
}
