package upgrade

import (
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/action"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/etcd/templates"
	"github.com/beclab/Olares/cli/pkg/gpu"
	"github.com/beclab/Olares/cli/pkg/terminus"
	"github.com/beclab/Olares/cli/version"
)

var version_1_12_5 = semver.MustParse("1.12.5")

type upgrader_1_12_5 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_5) Version() *semver.Version {
	cliVersion, err := semver.NewVersion(version.VERSION)
	// tolerate local dev version
	if err != nil {
		return version_1_12_5
	}
	if samePatchLevelVersion(version_1_12_5, cliVersion) && getReleaseLineOfVersion(cliVersion) == mainLine {
		return cliVersion
	}
	return version_1_12_5
}

func (u upgrader_1_12_5) AddedBreakingChange() bool {
	if u.Version().Equal(version_1_12_5) {
		return true
	}
	return false
}

func (u upgrader_1_12_5) PrepareForUpgrade() []task.Interface {
	return append([]task.Interface{
		&task.LocalTask{
			Name: "GenerateETCDService",
			Desc: "Generate etcd service",
			Action: &action.Template{
				Name:     "GenerateETCDService",
				Template: templates.ETCDService,
				Dst:      "/etc/systemd/system/etcd.service",
			},
		},
		&task.LocalTask{
			Name: "ReloadSystemd",
			Desc: "Reload systemd",
			Action: &terminus.SystemctlCommand{
				DaemonReloadPreExec: true,
			},
		},
	}, u.upgraderBase.PrepareForUpgrade()...)
}

func (u upgrader_1_12_5) UpgradeSystemComponents() []task.Interface {
	pre := []task.Interface{
		&task.LocalTask{
			Name:   "UpgradeL4BFLProxy",
			Action: &upgradeL4BFLProxy{Tag: "v0.3.12-1125"},
			Retry:  3,
			Delay:  5 * time.Second,
		},
		&task.LocalTask{
			Name:   "UpdateNodeGPUInfo",
			Action: new(gpu.UpdateNodeGPUInfo),
			Retry:  3,
			Delay:  5 * time.Second,
		},
		&task.LocalTask{
			Name:   "AddEntrancePolicy",
			Action: new(addEntrancePolicy),
			Retry:  3,
			Delay:  5 * time.Second,
		},
	}
	tasks := append(pre, u.upgraderBase.UpgradeSystemComponents()...)
	tasks = append(tasks,
		&task.LocalTask{
			Name: "WaitForAppServiceReady",
			Action: &waitForStatefulSetReady{
				Namespace: "os-framework",
				Name:      "app-service",
				InitDelay: 5 * time.Second,
			},
			Retry: 30,
			Delay: 10 * time.Second,
		},
		&task.LocalTask{
			Name:   "BackfillAppGPUConfig",
			Action: new(backfillAppGPUConfig),
			Retry:  3,
			Delay:  5 * time.Second,
		},
	)
	return tasks
}

func init() {
	registerMainUpgrader(upgrader_1_12_5{})
}
