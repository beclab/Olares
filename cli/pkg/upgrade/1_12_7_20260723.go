package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

// upgrader_1_12_7_20260723 migrates each node's containerd configuration from the
// legacy v2 layout to the v3 layout (config_path/certs.d + conf.d drop-ins). This
// is required because containerd 2.x ignores the old inline registry.mirrors, so
// Olares-managed docker.io mirrors only take effect via certs.d/hosts.toml. On
// GPU nodes the nvidia runtime is re-applied as a conf.d drop-in.
type upgrader_1_12_7_20260723 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_7_20260723) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260723")
}

// PreUpgradeNode rewrites the container runtime's configuration and restarts
// it. Every machine has its own /etc/containerd and its own GPU, so every
// machine migrates itself — the same reason the mainline 1.12.7 upgrader puts
// this work here rather than on the control node.
func (u upgrader_1_12_7_20260723) PreUpgradeNode() []task.Interface {
	return append(migrateContainerdConfigV3(), u.upgraderBase.PreUpgradeNode()...)
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260723{})
}
