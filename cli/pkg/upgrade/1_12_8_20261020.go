package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

// upgrader_1_12_8_20261020 regenerates the kubelet configuration so existing
// nodes stop using the soft eviction settings removed from the templates.
// The version must match the daily release carrying this change because
// getUpgraderByVersion selects an exact target version.
type upgrader_1_12_8_20261020 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_8_20261020) Version() *semver.Version {
	return semver.MustParse("1.12.8-20261020")
}

// Restart K3s after component upgrades finish, then wait for the API server
// and system components to recover.
func (u upgrader_1_12_8_20261020) PostUpgrade() []task.Interface {
	return append(regenerateKubeFiles(), u.upgraderBase.PostUpgrade()...)
}

func init() {
	registerDailyUpgrader(upgrader_1_12_8_20261020{})
}
