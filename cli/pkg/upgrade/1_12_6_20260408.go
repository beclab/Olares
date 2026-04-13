package upgrade

import (
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

type upgrader_1_12_6_20260408 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_6_20260408) Version() *semver.Version {
	return semver.MustParse("1.12.6-20260408")
}

func (u upgrader_1_12_6_20260408) PrepareForUpgrade() []task.Interface {
	return u.upgraderBase.PrepareForUpgrade()
}

func (u upgrader_1_12_6_20260408) UpgradeSystemComponents() []task.Interface {
	tasks := append([]task.Interface{}, u.upgraderBase.UpgradeSystemComponents()...)
	tasks = append(tasks,
		&task.LocalTask{
			Name: "WaitForAppServiceReady",
			Action: &waitForStatefulSetReady{
				Namespace: "os-framework",
				Name:      "app-service",
				InitDelay: 5 * time.Second,
			},
			Retry: 30,
			Delay: 10 * time.Second,
		},
		&task.LocalTask{
			Name:   "BackfillAppGPUConfig",
			Action: new(backfillAppGPUConfig),
			Retry:  3,
			Delay:  5 * time.Second,
		},
	)
	return tasks
}

func init() {
	registerDailyUpgrader(upgrader_1_12_6_20260408{})
}
