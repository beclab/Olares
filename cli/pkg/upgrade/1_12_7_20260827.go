package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/kubesphere/plugins"
)

type upgrader_1_12_7_20260827 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_7_20260827) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260827")
}

func (u upgrader_1_12_7_20260827) PrepareForUpgrade() []task.Interface {
	tasks := []task.Interface{&task.LocalTask{
		Name:   "CopyEmbeddedKSManifests",
		Action: new(plugins.CopyEmbedFiles),
	}}

	tasks = append(tasks, upgradePrometheusOperator()...)
	tasks = append(tasks, u.upgraderBase.PrepareForUpgrade()...)
	return tasks
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260827{})
}
