package upgrade

import (
	"github.com/beclab/Olares/cli/pkg/bootstrap/precheck"
	"github.com/beclab/Olares/cli/pkg/manifest"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

type Module struct {
	common.KubeModule
	manifest.ManifestModule
	TargetVersion *semver.Version

	// Stage restricts this run to one stage of the flow. Empty runs every stage
	// in order on this machine, which is what a single-node cluster does and
	// what this module did before an upgrade could be scheduled across nodes.
	Stage string
}

func (m *Module) Init() {
	m.Name = "UpgradeOlares"
	if m.Stage != "" {
		m.Name = "UpgradeOlares/" + m.Stage
	}

	tasks, err := m.stageTasks()
	if err != nil {
		// Init cannot fail, and refusing to run is the only correct answer
		// here, so the refusal becomes the module's single task.
		m.Tasks = []task.Interface{&task.LocalTask{
			Name:   "ResolveUpgradeStage",
			Action: &refuseUpgrade{err: err},
		}}
		return
	}
	m.Tasks = tasks
}

func (m *Module) stageTasks() ([]task.Interface, error) {
	if m.Stage == "" {
		return AllTasks(getUpgraderByVersion(m.TargetVersion)), nil
	}
	_, tasks, err := TasksForStage(m.TargetVersion, m.Stage)
	return tasks, err
}

// refuseUpgrade turns a planning failure into a failed task, so it is reported
// through the same path as everything else the pipeline can fail at.
type refuseUpgrade struct {
	common.KubeAction
	err error
}

func (a *refuseUpgrade) Execute(connector.Runtime) error { return a.err }

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
