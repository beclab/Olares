package preinstall

import (
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/action"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/storage"
	"github.com/beclab/Olares/cli/version"
)

type MaterializeModule struct {
	common.KubeModule
	InstallerDir      string
	RootDir           string
	OSVersion         string
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
				OSVersion:         m.OSVersion,
				ProfileSelections: m.ProfileSelections,
			},
		},
	}
}

type MaterializeAction struct {
	action.BaseAction
	InstallerDir      string
	RootDir           string
	OSVersion         string
	ProfileSelections ProfileSelections
}

func (a *MaterializeAction) Execute(_ connector.Runtime) error {
	return Materialize(a.InstallerDir, publishRootDir(a.RootDir), publishVersion(a.OSVersion), a.ProfileSelections)
}

// PublishCatalogDeclarationAction declares what an upgrade brings: the apps this
// version expects, taken from the catalog rather than from any medium.
type PublishCatalogDeclarationAction struct {
	action.BaseAction
	RootDir   string
	OSVersion string
}

func (a *PublishCatalogDeclarationAction) Execute(_ connector.Runtime) error {
	return PublishCatalogDeclaration(publishRootDir(a.RootDir), publishVersion(a.OSVersion))
}

func publishRootDir(rootDir string) string {
	if rootDir == "" {
		return storage.OlaresRootDir
	}
	return rootDir
}

// publishVersion falls back to the version of this binary, which is the version
// it is installing or upgrading to.
func publishVersion(osVersion string) string {
	if osVersion == "" {
		return version.VERSION
	}
	return osVersion
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
