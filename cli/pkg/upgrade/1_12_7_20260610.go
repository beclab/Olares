package upgrade

import (
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

const reverseProxyAgentImage = "beclab/reverse-proxy:v0.1.11"

type upgrader_1_12_7_20260610 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_7_20260610) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260610")
}

func (u upgrader_1_12_7_20260610) PrepareForUpgrade() []task.Interface {
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
	registerDailyUpgrader(upgrader_1_12_7_20260610{})
}
