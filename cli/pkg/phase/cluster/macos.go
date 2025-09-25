package cluster

import (
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/kubesphere/plugins"
	"github.com/beclab/Olares/cli/pkg/manifest"
	"github.com/beclab/Olares/cli/pkg/terminus"
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
		&terminus.InstallKubeblocksModule{},
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
