package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

// upgrader_1_12_7_20260914 moves the overlay gateway parent interface from the
// br-olares bridge to the wired NIC, addressed through the alternative name
// olares-lan. Enabling or disabling the overlay gateway no longer switches the
// host's primary link. On nodes where the bridge is active the steps run in
// this order inside UpgradeSystemComponents:
//  1. tear the bridge down and give the host address back to the wired NIC;
//  2. add the alternative name to that NIC and persist it;
//  3. re-render the underlay NAD so its master is the alternative name;
//  4. record the overlay gateway as enabled and recreate its Pods, whose
//     macvlan interfaces vanished with the bridge.
type upgrader_1_12_7_20260914 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_7_20260914) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260914")
}

func (u upgrader_1_12_7_20260914) UpgradeSystemComponents() []task.Interface {
	tasks := make([]task.Interface, 0)
	tasks = append(tasks, overlayDirectPreTasks()...)
	tasks = append(tasks, upgradeMultus()...)
	tasks = append(tasks, overlayDirectPostTasks()...)
	tasks = append(tasks, u.upgraderBase.UpgradeSystemComponents()...)
	return tasks
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260914{})
}
