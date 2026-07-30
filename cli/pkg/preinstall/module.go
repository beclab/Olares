package preinstall

import (
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/action"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/storage"
)

type MaterializeModule struct {
	common.KubeModule
	InstallerDir      string
	RootDir           string
	ProfileSelections ProfileSelections
}

func (m *MaterializeModule) Init() {
	m.Name = "MaterializeMarketPreinstall"
	m.Tasks = []task.Interface{
		&task.LocalTask{
			Name: "MaterializeMarketPreinstall",
			Action: &MaterializeAction{
				InstallerDir:      m.InstallerDir,
				RootDir:           m.RootDir,
				ProfileSelections: m.ProfileSelections,
			},
		},
	}
}

type MaterializeAction struct {
	action.BaseAction
	InstallerDir      string
	RootDir           string
	ProfileSelections ProfileSelections
}

func (a *MaterializeAction) Execute(_ connector.Runtime) error {
	rootDir := a.RootDir
	if rootDir == "" {
		rootDir = storage.OlaresRootDir
	}
	return Materialize(a.InstallerDir, rootDir, a.ProfileSelections)
}

type HFCacheMaterializeModule struct {
	common.KubeModule
}

func (m *HFCacheMaterializeModule) Init() {
	m.Name = "MaterializeHuggingFaceCache"
	m.Tasks = []task.Interface{
		&task.LocalTask{
			Name:   "MaterializeHuggingFaceCache",
			Action: &HFCacheMaterializeAction{},
		},
	}
}

type hfMaterializeFunc func(string, string, *hfOwnership) error

type HFCacheMaterializeAction struct {
	action.BaseAction
	materialize hfMaterializeFunc
}

func (a *HFCacheMaterializeAction) Execute(runtime connector.Runtime) error {
	materialize := a.materialize
	if materialize == nil {
		materialize = materializeHFArtifacts
	}
	return materialize(runtime.GetInstallerDir(), storage.HuggingFaceCacheDir, productionHFOwnership())
}
