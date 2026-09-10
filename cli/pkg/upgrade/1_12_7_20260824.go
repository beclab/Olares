package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

// upgrader_1_12_7_20260824 regenerates the kubelet configuration so an existing
// node picks up the larger system-reserved memory. The reserve is computed from
// the host's RAM and swap at generation time, so regenerating the files is all
// it takes.
//
// getUpgraderByVersion matches the target version exactly, so this only runs for
// a release cut as 1.12.7-20260824. If the release carrying this change ends up
// stamped with another date, rename this file and the version below to match, or
// existing clusters silently keep the old reserve.
type upgrader_1_12_7_20260824 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_7_20260824) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260824")
}

// PostUpgrade, rather than PrepareForUpgrade, is where the regeneration goes:
// it restarts k3s, and every phase before this one runs tasks that exec into
// pods. Doing it up front made those tasks race the kubelet's re-sync and fail
// with "pod does not exist". Here the only thing that follows is the base's
// wait for the system components to come back, which is exactly the guard a
// restart wants.
func (u upgrader_1_12_7_20260824) PostUpgrade() []task.Interface {
	return append(regenerateKubeFiles(), u.upgraderBase.PostUpgrade()...)
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260824{})
}
