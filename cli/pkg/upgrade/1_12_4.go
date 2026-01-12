package upgrade

import (
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/version"
)

var version_1_12_4 = semver.MustParse("1.12.4")

type upgrader_1_12_4 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_4) Version() *semver.Version {
	cliVersion, err := semver.NewVersion(version.VERSION)
	// tolerate local dev version
	if err != nil {
		return version_1_12_4
	}
	if samePatchLevelVersion(version_1_12_4, cliVersion) && getReleaseLineOfVersion(cliVersion) == mainLine {
		return cliVersion
	}
	return version_1_12_4
}

func (u upgrader_1_12_4) AddedBreakingChange() bool {
	if u.Version().Equal(version_1_12_4) {
		return true
	}
	return false
}

func (u upgrader_1_12_4) NeedRestart() bool {
	return true
}

func (u upgrader_1_12_4) PrepareForUpgrade() []task.Interface {
	tasks := make([]task.Interface, 0)

	tasks = append(tasks, upgradeKsConfig()...)
	tasks = append(tasks, upgradePrometheusServiceMonitorKubelet()...)
	tasks = append(tasks, upgradeKSCore()...)
	tasks = append(tasks, upgradeNodeExporter()...)
	tasks = append(tasks,
		&task.LocalTask{
			Name:   "DeleteArgoProjV1alpha1CRDs",
			Action: new(deleteArgoProjV1alpha1CRDs),
			Retry:  3,
			Delay:  5 * time.Second,
		},
	)
	tasks = append(tasks, regenerateKubeFiles()...)
	tasks = append(tasks, u.upgraderBase.PrepareForUpgrade()...)
	return tasks
}

func (u upgrader_1_12_4) UpgradeSystemComponents() []task.Interface {
	pre := []task.Interface{
		&task.LocalTask{
			Name:   "UpgradeL4BFLProxy",
			Action: &upgradeL4BFLProxy{Tag: "v0.3.9"},
			Retry:  3,
			Delay:  5 * time.Second,
		},
	}
	return append(pre, u.upgraderBase.UpgradeSystemComponents()...)
}

func (u upgrader_1_12_4) UpdateOlaresVersion() []task.Interface {
	var tasks []task.Interface
	tasks = append(tasks,
		&task.LocalTask{
			Name:   "UpgradeGPUDriver",
			Action: new(upgradeGPUDriverIfNeeded),
		},
	)
	tasks = append(tasks, u.upgraderBase.UpdateOlaresVersion()...)
	tasks = append(tasks,
		&task.LocalTask{
			Name:   "RebootIfNeeded",
			Action: new(rebootIfNeeded),
		},
	)
	return tasks
}

func init() {
	registerMainUpgrader(upgrader_1_12_4{})
}
