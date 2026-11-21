package upgrade

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

// A downstream distribution can register one extension per exact target from
// its own Go file in this package. The upstream upgrader still supplies all
// its tasks, including when a later sync adds an upgrader for the same target.
type upgradeExtension struct {
	Target *semver.Version
	Stages map[string]stageExtension

	// Breaking makes this target a required stop for subsequent upgrades on
	// its release line. NeedRestart is reported in the target's upgrade spec.
	Breaking    bool
	NeedRestart bool
}

// Factories must return fresh tasks for each plan. Task.Init attaches state
// to the task and its action when a node executes the plan.
type stageExtension struct {
	Before func() []task.Interface
	After  func() []task.Interface
}

var upgradeExtensions []upgradeExtension

// registerUpgradeExtension is called from init. Keeping all contributions
// for one target together makes their order explicit in the downstream file.
func registerUpgradeExtension(ext upgradeExtension) {
	if ext.Target == nil {
		panic("upgrade extension needs a target version")
	}
	if extensionFor(ext.Target) != nil {
		panic(fmt.Sprintf("duplicate upgrade extension for %s", ext.Target))
	}
	for stage, tasks := range ext.Stages {
		definition, ok := StageByName(stage)
		if !ok {
			panic(fmt.Sprintf("upgrade extension for %s: unknown stage %q", ext.Target, stage))
		}
		if definition.Runner != RunnerOlaresCLI {
			panic(fmt.Sprintf("upgrade extension for %s: stage %q runs in olaresd", ext.Target, stage))
		}
		if tasks.Before == nil && tasks.After == nil {
			panic(fmt.Sprintf("upgrade extension for %s: no task factory for stage %q", ext.Target, stage))
		}
	}
	upgradeExtensions = append(upgradeExtensions, ext)
}

func extensionFor(target *semver.Version) *upgradeExtension {
	for i := range upgradeExtensions {
		if upgradeExtensions[i].Target.Equal(target) {
			return &upgradeExtensions[i]
		}
	}
	return nil
}

type extendedUpgrader struct {
	upgrader
	ext *upgradeExtension
}

func (u extendedUpgrader) AddedBreakingChange() bool {
	return u.ext.Breaking || u.upgrader.AddedBreakingChange()
}

func (u extendedUpgrader) NeedRestart() bool {
	return u.ext.NeedRestart || u.upgrader.NeedRestart()
}

func (e stageExtension) withUpstream(upstream []task.Interface) []task.Interface {
	var tasks []task.Interface
	if e.Before != nil {
		tasks = e.Before()
	}
	tasks = append(tasks, upstream...)
	if e.After != nil {
		tasks = append(tasks, e.After()...)
	}
	return tasks
}
