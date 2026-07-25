package system

import (
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/kubesphere"
	"github.com/beclab/Olares/cli/pkg/manifest"
	"github.com/beclab/Olares/cli/pkg/terminus"
)

var _ phaseBuilder = &macOsPhaseBuilder{}

type macOsPhaseBuilder struct {
	runtime     *common.KubeRuntime
	manifestMap manifest.InstallationManifest
}

func (m *macOsPhaseBuilder) build() []module.Module {
	// TODO: install minikube
	modules := []module.Module{
		&kubesphere.CreateMinikubeClusterModule{},
	}
	modules = append(modules, marketPreinstallModules(
		m.manifestMap,
		m.runtime.GetInstallerDir(),
		m.runtime.GetBaseDir(),
		productionPreinstallSelections(m.runtime),
	)...)
	return append(modules, &terminus.PreparedModule{})
}
