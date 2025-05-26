package system

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/kubesphere"
	"bytetrade.io/web3os/installer/pkg/manifest"
	"bytetrade.io/web3os/installer/pkg/terminus"
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
