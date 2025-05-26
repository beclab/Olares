package upgrade

import (
	"time"

	"bytetrade.io/web3os/installer/pkg/bootstrap/precheck"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/terminus"
	"github.com/Masterminds/semver/v3"
)

type UpgradeModule struct {
	common.KubeModule
	CurrentVersion *semver.Version
	TargetVersion  *semver.Version
}

var (
	preTasks []*upgradeTask

	coreTasks = []*upgradeTask{
		{
			Task: &task.LocalTask{
				Name:   "PrepareUserInfoForUpgrade",
				Action: new(PrepareUserInfoForUpgrade),
				Retry:  5,
			},
			Current: atLeasVersion112,
			Target:  atLeasVersion112,
		},
		{
			Task: &task.LocalTask{
				Name:   "ClearAppChartValues",
				Action: new(terminus.ClearAppValues),
			},
			Current: atLeasVersion112,
			Target:  atLeasVersion112,
		},
		{
			Task: &task.LocalTask{
				Name:   "ClearBFLChartValues",
				Action: new(terminus.ClearBFLValues),
			},
			Current: atLeasVersion112,
			Target:  atLeasVersion112,
		},
		{
			Task: &task.LocalTask{
				Name:   "UpdateChartsInAppService",
				Action: new(terminus.CopyAppServiceHelmFiles),
				Retry:  5,
			},
			Current: atLeasVersion112,
			Target:  atLeasVersion112,
		},
		{
			Task: &task.LocalTask{
				Name:   "UpgradeUserComponents",
				Action: new(UpgradeUserComponents),
				Retry:  5,
				Delay:  15 * time.Second,
			},
			Current: atLeasVersion112,
			Target:  atLeasVersion112,
		},
		{
			Task: &task.LocalTask{
				Name:   "UpdateReleaseFile",
				Action: new(terminus.WriteReleaseFile),
			},
			Current: atLeasVersion112,
			Target:  atLeasVersion112,
		},
		// this task updates the version in the CR
		// so put this at last to make the whole pipeline
		// reentrant
		// maybe it should be put at the last of post tasks
		// when post tasks are actually needed
		{
			Task: &task.LocalTask{
				Name:   "UpgradeSystemComponents",
				Action: new(UpgradeSystemComponents),
				Retry:  10,
				Delay:  15 * time.Second,
			},
			Current: atLeasVersion112,
			Target:  atLeasVersion112,
		},
		{
			Task: &task.LocalTask{
				Name:   "EnsurePodsUpAndRunningAgain",
				Action: new(terminus.CheckKeyPodsRunning),
				Delay:  15 * time.Second,
				Retry:  60,
			},
			Current: atLeasVersion112,
			Target:  atLeasVersion112,
		},
	}

	postTasks []*upgradeTask
)

func (m *UpgradeModule) Init() {
	m.Name = "UpgradeOlares"

	// calculate tasks based on version difference
	tasks := m.calculateUpgradeTasks()

	m.Tasks = tasks
}

func (m *UpgradeModule) calculateUpgradeTasks() []task.Interface {
	var tasks []task.Interface

	// for now, tasks are grouped into pre-upgrade/core-upgrade/post-upgrade tasks
	// only for business logic compatibility
	// they are still a normal sequence of tasks to be executed
	// for the module layer
	tasks = append(tasks, m.calculatePreUpgradeTasks()...)
	tasks = append(tasks, m.calculateCoreUpgradeTasks()...)
	tasks = append(tasks, m.calculatePostUpgradeTasks()...)

	return tasks
}

func (m *UpgradeModule) getTasksToExecute(unfiltered []*upgradeTask) []task.Interface {
	var filtered []task.Interface
	for _, t := range unfiltered {
		if t.Match(m.CurrentVersion, m.TargetVersion) {
			filtered = append(filtered, t.Task)
		}
	}
	return filtered
}

func (m *UpgradeModule) calculatePreUpgradeTasks() []task.Interface {
	return m.getTasksToExecute(preTasks)
}

func (m *UpgradeModule) calculateCoreUpgradeTasks() []task.Interface {
	return m.getTasksToExecute(coreTasks)
}

func (m *UpgradeModule) calculatePostUpgradeTasks() []task.Interface {
	return m.getTasksToExecute(postTasks)
}

type PrecheckModule struct {
	common.KubeModule
}

func (m *PrecheckModule) Init() {
	m.Name = "UpgradePrecheck"

	checkers := []precheck.Checker{
		new(precheck.MasterNodeReadyCheck),
		new(precheck.RootPartitionAvailableSpaceCheck),
	}
	runPreChecks := &task.LocalTask{
		Name: "UpgradePrecheck",
		Action: &precheck.RunChecks{
			Checkers: checkers,
		},
	}

	m.Tasks = []task.Interface{
		runPreChecks,
	}
}
