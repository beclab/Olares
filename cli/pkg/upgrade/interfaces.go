package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

type upgrader interface {
	PrepareForUpgrade() []task.Interface
	ClearAppChartValues() []task.Interface
	ClearBFLChartValues() []task.Interface
	UpdateChartsInAppService() []task.Interface
	UpgradeUserComponents() []task.Interface
	UpdateReleaseFile() []task.Interface
	UpgradeSystemComponents() []task.Interface
	UpdateOlaresVersion() []task.Interface
	PostUpgrade() []task.Interface
	AddedBreakingChange() bool
	NeedRestart() bool
}

type breakingUpgrader interface {
	upgrader
	Version() *semver.Version
}
