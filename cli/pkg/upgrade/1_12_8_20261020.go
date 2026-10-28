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

// Kubelet and systemd configuration is local to each node.
func (u upgrader_1_12_8_20261020) PostUpgradeNode() []task.Interface {
	return append(regenerateKubeFilesOnNode(), u.upgraderBase.PostUpgradeNode()...)
}

// Kubeadm control-plane files and cluster health are checked on the admin node.
func (u upgrader_1_12_8_20261020) PostUpgrade() []task.Interface {
	return append(regenerateKubeFilesOnControlNode(), u.upgraderBase.PostUpgrade()...)
}

func init() {
	registerDailyUpgrader(upgrader_1_12_8_20261020{})
}
