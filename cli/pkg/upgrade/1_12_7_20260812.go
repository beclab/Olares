package upgrade

import (
	"time"

	"github.com/beclab/Olares/cli/pkg/core/task"

	"github.com/Masterminds/semver/v3"
)

type upgrader_1_12_7_20260812 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_7_20260812) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260812")
}

func (u upgrader_1_12_7_20260812) PrepareForUpgrade() []task.Interface {
	pre := []task.Interface{
		&task.LocalTask{
			Name:   "UpgradeUserReverseProxyAgent",
			Action: new(upgradeUserReverseProxyAgent),
			Retry:  5,
			Delay:  10 * time.Second,
		},
	}
	return append(pre, u.upgraderBase.PrepareForUpgrade()...)
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260812{})
}
