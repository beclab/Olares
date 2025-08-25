package nfd

import (
	"path"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/prepare"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

// Module wires the actions to write and apply the NFD manifest.
type Module struct {
	common.KubeModule
}

// Init defines the task graph for installing Node Feature Discovery.
func (m *Module) Init() {
	// Write nfd.yaml to addons dir on first control-plane node,
	// then kubectl apply it (cluster-scoped resources), then wait for readiness.
	write := &task.KubeTask{
		Name:  "WriteNfdManifest",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		// only run once on the first master
		Prepare: &prepare.PrepareCollection{new(common.OnlyFirstMaster)},
		Action:  new(WriteNfdFile),
	}
	apply := &task.KubeTask{
		Name:     "ApplyNfdManifest",
		Hosts:    m.Runtime.GetHostsByRole(common.Master),
		Prepare:  &prepare.PrepareCollection{new(common.OnlyFirstMaster)},
		Action:   new(ApplyNfd),
		Parallel: false,
		Retry:    1,
		Delay:    2 * time.Second,
		Timeout:  2 * time.Minute,
	}

	wait := &task.KubeTask{
		Name:     "WaitNfdReady",
		Hosts:    m.Runtime.GetHostsByRole(common.Master),
		Prepare:  &prepare.PrepareCollection{new(common.OnlyFirstMaster)},
		Action:   new(WaitNfdReady),
		Parallel: false,
		Timeout:  5 * time.Minute,
	}

	m.Tasks = []task.Interface{write, apply, wait}
}

// ManifestPath returns where we will place the manifest on the host.
func ManifestPath() string {
	return path.Join(common.KubeAddonsDir, "nfd.yaml")
}
