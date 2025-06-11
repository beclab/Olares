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
	return []module.Module{
		&kubesphere.CreateMinikubeClusterModule{},
		&terminus.PreparedModule{},
	}
}
