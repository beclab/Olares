package upgrade

import (
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/kubesphere/plugins"

	"github.com/Masterminds/semver/v3"
)

type upgrader_1_12_7_20260624 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_7_20260624) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260624")
}

func (u upgrader_1_12_7_20260624) UpgradeSystemComponents() []task.Interface {
	tasks := make([]task.Interface, 0)
	tasks = append(tasks, &task.LocalTask{
		Name:   "CopyEmbeddedKSManifests",
		Action: new(plugins.CopyEmbedFiles),
	})
	tasks = append(tasks, upgradePrometheusOperator()...)
	tasks = append(tasks, u.upgraderBase.UpgradeSystemComponents()...)

	return tasks
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260624{})
}
