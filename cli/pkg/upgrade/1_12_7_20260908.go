package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

// upgrader_1_12_7_20260908 regenerates the kubelet configuration and the systemd
// memory protection drop-ins, so that an existing node picks up the reserve and
// the memory.min values from pkg/utils/kubelet_reserved.go. Both are derived from
// the host at generation time, so regenerating the files is all it takes.
//
// The two have to be refreshed together: system.slice is asked to protect
// exactly what the reserve holds back, and a node carrying one without the other
// would either protect memory already promised to pods or leave the control
// plane reclaimable. GenerateK3sService writes both, which is why it is the task
// that does this.
//
// getUpgraderByVersion matches the target version exactly, so this only runs for
// a release cut as 1.12.7-20260908. If the release carrying this change ends up
// stamped with another date, rename this file and the version below to match, or
// existing clusters silently keep the old reserve.
type upgrader_1_12_7_20260908 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_7_20260908) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260908")
}

// PostUpgrade, rather than PrepareForUpgrade, is where the regeneration goes:
// it restarts k3s, and the phases before this one run tasks that exec into pods,
// which then race the kubelet's re-sync and fail with "pod does not exist". Here
// the only thing that follows is the base's wait for the system components to
// come back, which is exactly the guard a restart wants.
func (u upgrader_1_12_7_20260908) PostUpgrade() []task.Interface {
	return append(regenerateKubeFiles(), u.upgraderBase.PostUpgrade()...)
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260908{})
}
