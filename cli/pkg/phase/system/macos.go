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
	// Market preinstall only supports Linux; see linux.go. minikube also loads
	// no images locally, which is why wsl.go keeps the preload this omits.
	return []module.Module{
		&kubesphere.CreateMinikubeClusterModule{},
		&terminus.PreparedModule{},
	}
}
