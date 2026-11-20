// Package lifecycle lets distributions register tasks at specific points
// in the prepare, install and uninstall pipelines.
package lifecycle

import (
	"fmt"

	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

type Point string

const (
	BeforePrepared       Point = "BeforePrepared"
	BeforeInstalled      Point = "BeforeInstalled"
	BeforeClusterRemoval Point = "BeforeClusterRemoval"
	BeforePrepareCleanup Point = "BeforePrepareCleanup"
)

type TaskFactory func() []task.Interface

var factories = map[Point][]TaskFactory{
	BeforePrepared:       nil,
	BeforeInstalled:      nil,
	BeforeClusterRemoval: nil,
	BeforePrepareCleanup: nil,
}

// Register adds a factory during package initialization. Factories run in
// registration order and must return fresh tasks for each pipeline run.
func Register(point Point, factory TaskFactory) {
	existing, ok := factories[point]
	if !ok || factory == nil {
		panic(fmt.Sprintf("invalid lifecycle task registration for %q", point))
	}
	factories[point] = append(existing, factory)
}

// Modules builds an ordinary task module only when the distribution supplies
// work for this point.
func (p Point) Modules() []module.Module {
	var tasks []task.Interface
	for _, factory := range factories[p] {
		tasks = append(tasks, factory()...)
	}
	if len(tasks) == 0 {
		return nil
	}
	return []module.Module{&module.BaseTaskModule{
		BaseModule: module.BaseModule{Name: string(p)},
		Tasks:      tasks,
	}}
}
