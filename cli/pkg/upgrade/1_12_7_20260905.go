package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

// upgrader_1_12_7_20260905 ships the Olares cni-plugins release with the
// stable DHCP client identifier and the ipam.sendRelease switch
// (beclab/plugins v1.6.2-olares2). Order matters:
//  1. PrepareForUpgrade replaces /opt/cni/bin from the manifest and restarts
//     cni-dhcp (the generic task set never swaps CNI binaries);
//  2. UpgradeSystemComponents re-renders the underlay-macvlan NAD so it
//     carries ipam.sendRelease=false;
//  3. PostUpgrade recreates Overlay Gateway pods so their leases are owned by
//     the new daemon and created against the new NAD.
type upgrader_1_12_7_20260905 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_7_20260905) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260905")
}

func (u upgrader_1_12_7_20260905) PrepareForUpgrade() []task.Interface {
	tasks := make([]task.Interface, 0)
	tasks = append(tasks, cniDhcpBinaryUpgradeTasks()...)
	tasks = append(tasks, u.upgraderBase.PrepareForUpgrade()...)
	return tasks
}

func (u upgrader_1_12_7_20260905) UpgradeSystemComponents() []task.Interface {
	tasks := make([]task.Interface, 0)
	tasks = append(tasks, upgradeMultus()...)
	tasks = append(tasks, u.upgraderBase.UpgradeSystemComponents()...)
	return tasks
}

func (u upgrader_1_12_7_20260905) PostUpgrade() []task.Interface {
	tasks := make([]task.Interface, 0)
	tasks = append(tasks, overlayGatewayRecreateTasks()...)
	tasks = append(tasks, u.upgraderBase.PostUpgrade()...)
	return tasks
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260905{})
}
