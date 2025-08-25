package amd

import (
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/prepare"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

// Module wires the actions to install the AMD GPU device plugin (ROCm).
type Module struct {
	common.KubeModule
}

func (m *Module) Init() {
	write := &task.KubeTask{
		Name:     "WriteAmdGpuManifest",
		Hosts:    m.Runtime.GetHostsByRole(common.Master),
		Prepare:  &prepare.PrepareCollection{new(common.OnlyFirstMaster)},
		Action:   new(WriteManifest),
		Parallel: false,
	}
	apply := &task.KubeTask{
		Name:     "ApplyAmdGpuManifest",
		Hosts:    m.Runtime.GetHostsByRole(common.Master),
		Prepare:  &prepare.PrepareCollection{new(common.OnlyFirstMaster)},
		Action:   new(ApplyManifest),
		Parallel: false,
	}
	wait := &task.KubeTask{
		Name:     "WaitAmdGpuReady",
		Hosts:    m.Runtime.GetHostsByRole(common.Master),
		Prepare:  &prepare.PrepareCollection{new(common.OnlyFirstMaster)},
		Action:   new(WaitReady),
		Parallel: false,
		Timeout:  5 * time.Minute,
	}

	m.Tasks = []task.Interface{write, apply, wait}
}
