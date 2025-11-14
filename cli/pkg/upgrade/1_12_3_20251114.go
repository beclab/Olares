package upgrade

import (
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

type upgrader_1_12_3_20251114 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_3_20251114) Version() *semver.Version {
	return semver.MustParse("1.12.3-20251114")
}

func (u upgrader_1_12_3_20251114) UpgradeSystemComponents() []task.Interface {
	pre := []task.Interface{
		&task.LocalTask{
			Name:   "UpgradeL4BFLProxy",
			Action: &upgradeL4BFLProxy{Tag: "v0.3.7"},
			Retry:  3,
			Delay:  5 * time.Second,
		},
	}
	return append(pre, u.upgraderBase.UpgradeSystemComponents()...)
}

func init() {
	registerDailyUpgrader(upgrader_1_12_3_20251114{})
}
