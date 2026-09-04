package upgrade

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

type upgrader_1_12_7_20260604 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_7_20260604) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260604")
}

func (u upgrader_1_12_7_20260604) PrepareForUpgrade() []task.Interface {
	tasks := make([]task.Interface, 0)
	tasks = append(tasks, &task.LocalTask{
		Name:   "UpgradeCniPluginsBinary",
		Action: new(upgradeCniPluginsBinary),
	},
	)
	tasks = append(tasks, upgradeNetworkManagerConfig()...)

	tasks = append(tasks, u.upgraderBase.PrepareForUpgrade()...)
	return tasks

}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260604{})
}

// only for daily build to fix cni-plugins binary issue, will be removed in future
type upgradeCniPluginsBinary struct {
	common.KubeAction
}

func (u *upgradeCniPluginsBinary) Execute(runtime connector.Runtime) error {
	dst, err := syncCniPluginsArchive(runtime, u.KubeConf.Arg.Manifest)
	if err != nil {
		return err
	}
	// this upgrader deliberately replaces every plugin in the archive: it
	// shipped as a fix for broken cni-plugins binaries on the node
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("tar -zxf %s -C %s", dst, cniBinDir), false, false); err != nil {
		return err
	}
	if _, err := runtime.GetRunner().SudoCmd("systemctl restart cni-dhcp", false, false); err != nil {
		return err
	}
	return nil
}
