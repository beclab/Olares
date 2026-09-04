package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

// upgrader_1_12_7_20260907 ships the Olares cni-plugins release with the
// stable DHCP client identifier and the ipam.sendRelease switch
// (beclab/plugins v1.6.2-olares2). The three steps run back to back inside
// UpgradeSystemComponents so the window in which running Pods have no daemon
// renewing their lease is as short as possible:
//  1. install the cni-plugins archive from the manifest and restart cni-dhcp
//     (the generic task set never swaps CNI binaries);
//  2. re-render the underlay-macvlan NAD so it carries ipam.sendRelease=false;
//  3. recreate Overlay Gateway Pods so their leases are owned by the new
//     daemon and created against the new NAD.
type upgrader_1_12_7_20260907 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_7_20260907) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260907")
}

func (u upgrader_1_12_7_20260907) UpgradeSystemComponents() []task.Interface {
	tasks := make([]task.Interface, 0)
	tasks = append(tasks, upgradeMultus()...)
	tasks = append(tasks, cniDhcpBinaryUpgradeTasks()...)
	tasks = append(tasks, overlayGatewayRecreateTasks()...)
	tasks = append(tasks, u.upgraderBase.UpgradeSystemComponents()...)
	return tasks
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260907{})
}
