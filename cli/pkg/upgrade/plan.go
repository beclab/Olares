package upgrade

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

// A Plan is one version's upgrade laid out over the stages of the flow.
//
// The stages are not derived from anything: they are UpgradeFlow, in order,
// every one of them, whether or not this version puts any tasks in them. What
// a version contributes is the task list inside each stage. So two releases
// produce plans with the same shape and different contents, and "which stages
// are there" has one answer for the whole system rather than one per release.
type Plan struct {
	// Version is the target this plan upgrades to. A node checks it before
	// running anything: a plan is only meaningful to the binary that made it.
	Version string `json:"version"`

	Stages []PlanStage `json:"stages"`
}

// PlanStage is one stage of the flow with the tasks this version put in it.
type PlanStage struct {
	Name           string    `json:"name"`
	Placement      Placement `json:"placement"`
	MaxParallel    int       `json:"maxParallel,omitempty"`
	TimeoutSeconds int       `json:"timeoutSeconds,omitempty"`
	AwaitRestart   bool      `json:"awaitRestart,omitempty"`
	Desc           string    `json:"desc,omitempty"`
	Tasks          []string  `json:"tasks"`

	// tasks is what actually runs. It is unexported: outside this process a
	// stage is its name, and the node resolves it against its own plan.
	tasks []task.Interface
}

// Empty reports a stage this version contributes nothing to. It stays in the
// plan — the flow is the flow — and the orchestrator skips it.
func (h PlanStage) Empty() bool { return len(h.tasks) == 0 }

// BuildPlan lays out an upgrade to target over the flow.
func BuildPlan(target *semver.Version) (Plan, error) {
	if target == nil {
		return Plan{}, fmt.Errorf("no target version to plan an upgrade for")
	}
	return planFor(target.Original(), getUpgraderByVersion(target)), nil
}

func planFor(version string, u upgrader) Plan {
	byStage := stageTasks(u)

	plan := Plan{Version: version, Stages: make([]PlanStage, 0, len(UpgradeFlow))}
	for _, stage := range UpgradeFlow {
		tasks := byStage[stage.Name]
		names := make([]string, 0, len(tasks))
		for _, t := range tasks {
			names = append(names, t.GetName())
		}
		plan.Stages = append(plan.Stages, PlanStage{
			Name:           stage.Name,
			Placement:      stage.Placement,
			MaxParallel:    stage.MaxParallel,
			TimeoutSeconds: stage.TimeoutSeconds,
			AwaitRestart:   stage.AwaitRestart,
			Desc:           stage.Desc,
			Tasks:          names,
			tasks:          tasks,
		})
	}
	return plan
}

// Stage finds one stage of the plan.
func (p Plan) Stage(name string) (PlanStage, bool) {
	for _, h := range p.Stages {
		if h.Name == name {
			return h, true
		}
	}
	return PlanStage{}, false
}

// StageNames lists the stages in execution order.
func (p Plan) StageNames() []string {
	names := make([]string, 0, len(p.Stages))
	for _, h := range p.Stages {
		names = append(names, h.Name)
	}
	return names
}

// TasksForStage resolves a stage name against the plan for target and returns
// the tasks in it.
//
// What makes the answer the same one the control node scheduled is that both
// derived it from the same olares-cli version, which the orchestrator checks
// before it dispatches anything; see the daemon's checkWorkersCanUpgrade.
func TasksForStage(target *semver.Version, stageName string) (Plan, []task.Interface, error) {
	if target == nil {
		return Plan{}, nil, fmt.Errorf("no target version to plan an upgrade for")
	}
	plan := planFor(target.Original(), getUpgraderByVersion(target))

	stage, ok := plan.Stage(stageName)
	if !ok {
		return plan, nil, fmt.Errorf("the upgrade flow has no stage %q; it has %s",
			stageName, strings.Join(plan.StageNames(), ", "))
	}
	return plan, stage.tasks, nil
}

// AllTasks is the whole upgrade as one sequence, in flow order. It is what a
// single-node cluster runs: one machine, every stage, nothing to schedule.
func AllTasks(u upgrader) []task.Interface {
	byStage := stageTasks(u)
	var all []task.Interface
	for _, stage := range UpgradeFlow {
		all = append(all, byStage[stage.Name]...)
	}
	return all
}
