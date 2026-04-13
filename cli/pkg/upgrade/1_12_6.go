package upgrade

import (
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/version"
)

var version_1_12_6 = semver.MustParse("1.12.6")

type upgrader_1_12_6 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_6) Version() *semver.Version {
	cliVersion, err := semver.NewVersion(version.VERSION)
	// tolerate local dev version
	if err != nil {
		return version_1_12_6
	}
	if samePatchLevelVersion(version_1_12_6, cliVersion) && getReleaseLineOfVersion(cliVersion) == mainLine {
		return cliVersion
	}
	return version_1_12_6
}

func (u upgrader_1_12_6) AddedBreakingChange() bool {
	if u.Version().Equal(version_1_12_6) {
		return true
	}
	return false
}

func (u upgrader_1_12_6) PrepareForUpgrade() []task.Interface {
	tasks := make([]task.Interface, 0)
	tasks = append(tasks, upgradeKsConfig()...)
	tasks = append(tasks, upgradeKubernetesPrometheusRule()...)
	tasks = append(tasks, upgradeNodeExporterServiceMonitor()...)
	tasks = append(tasks, upgradeNodeExporter()...)
	tasks = append(tasks, upgradeKSCore()...)

	tasks = append(tasks, u.upgraderBase.PrepareForUpgrade()...)
	return tasks
}

func (u upgrader_1_12_6) UpgradeSystemComponents() []task.Interface {
	pre := []task.Interface{
		&task.LocalTask{
			Name:   "UpdateL4DeploymentSpec",
			Action: new(updateL4DeploymentSpec),
			Retry:  3,
			Delay:  5 * time.Second,
		},
		&task.LocalTask{
			Name:   "PatchNodeAffinityToRequired",
			Action: new(patchNodeAffinityToRequired),
			Retry:  3,
			Delay:  5 * time.Second,
		},
	}
	pre = append(pre, u.upgraderBase.UpgradeSystemComponents()...)
	return append(pre,
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
}

func init() {
	registerMainUpgrader(upgrader_1_12_6{})
}
