package preinstall

import (
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/action"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/storage"
)

type PublishDeclarationModule struct {
	common.KubeModule
	InstallerDir      string
	RootDir           string
	OSVersion         string
	ProfileSelections ProfileSelections
}

func (m *PublishDeclarationModule) Init() {
	m.Name = "PublishMarketPreinstallDeclaration"
	m.Tasks = []task.Interface{
		&task.LocalTask{
			Name: "PublishMarketPreinstallDeclaration",
			Action: &PublishDeclarationAction{
				InstallerDir:      m.InstallerDir,
				RootDir:           m.RootDir,
				OSVersion:         m.OSVersion,
				ProfileSelections: m.ProfileSelections,
			},
		},
	}
}

// PublishDeclarationAction declares what this version expects a device to have.
//
// InstallerDir is the one field whose emptiness means something: no medium was
// brought, which is the ordinary shape of an upgrade, and the catalog apps of the
// release being installed are declared all the same. The other two are required
// -- Publish refuses an empty version, and an empty root would write the
// declaration where nothing mounts it -- so every caller states them rather than
// falling through to a default here, where the caller's reason for the value is
// no longer visible.
type PublishDeclarationAction struct {
	action.BaseAction
	InstallerDir      string
	RootDir           string
	OSVersion         string
	ProfileSelections ProfileSelections
}

func (a *PublishDeclarationAction) Execute(_ connector.Runtime) error {
	return Publish(a.InstallerDir, a.RootDir, a.OSVersion, a.ProfileSelections)
}

type HFCacheMaterializeModule struct {
	common.KubeModule
	Skip bool
}

func (m *HFCacheMaterializeModule) IsSkip() bool {
	return m.Skip
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
