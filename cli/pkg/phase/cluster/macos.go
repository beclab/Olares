package cluster

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/kubesphere/plugins"
	"bytetrade.io/web3os/installer/pkg/manifest"
	"bytetrade.io/web3os/installer/pkg/terminus"
)

type macosInstallPhaseBuilder struct {
	runtime     *common.KubeRuntime
	manifestMap manifest.InstallationManifest
}

func (m *macosInstallPhaseBuilder) base() phase {
	mo := []module.Module{
		&plugins.CopyEmbed{},
		&terminus.CheckPreparedModule{Force: true},
	}

	return mo
}

func (m *macosInstallPhaseBuilder) installCluster() phase {
	return NewDarwinClusterPhase(m.runtime, m.manifestMap)
}

func (m *macosInstallPhaseBuilder) installTerminus() phase {
	return []module.Module{
		&terminus.GetNATGatewayIPModule{},
		&terminus.InstallAccountModule{},
		&terminus.InstallSettingsModule{},
		&terminus.InstallOsSystemModule{},
		&terminus.InstallLauncherModule{},
		&terminus.InstallAppsModule{},
	}
}

func (m *macosInstallPhaseBuilder) build() []module.Module {
	return m.base().
		addModule(m.installCluster()...).
		addModule(m.installTerminus()...).
		addModule(&terminus.InstalledModule{}).
		addModule(&terminus.WelcomeModule{})
}
