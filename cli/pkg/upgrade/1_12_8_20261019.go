package upgrade

import (
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/kubesphere/plugins"
)

type upgrader_1_12_8_20261019 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_8_20261019) Version() *semver.Version {
	return semver.MustParse("1.12.8-20261019")
}

func (u upgrader_1_12_8_20261019) UpgradeSystemComponents() []task.Interface {
	tasks := []task.Interface{
		&task.LocalTask{
			Name:   "CopyEmbeddedKSManifests",
			Action: new(plugins.CopyEmbedFiles),
		},
		&task.LocalTask{
			Name:   "ApplyPrometheus",
			Action: new(applyPrometheusAction),
			Retry:  5,
			Delay:  5 * time.Second,
		},
	}
	tasks = append(tasks, u.upgraderBase.UpgradeSystemComponents()...)
	return tasks
}

func init() {
	registerDailyUpgrader(upgrader_1_12_8_20261019{})
}
