package terminus

import (
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

// InstallAppGatewaySystemModule installs app-gateway-system dependencies and control plane
// before os-framework so Gateway API CRDs and ingress control plane are ready.
type InstallAppGatewaySystemModule struct {
	common.KubeModule
}

func (m *InstallAppGatewaySystemModule) Init() {
	logger.InfoInstallationProgress("Installing unified ingress (Linkerd + Envoy Gateway + app-gateway-system) ...")
	m.Name = "InstallAppGatewaySystemModule"

	checkInstaller := &task.LocalTask{
		Name:   "ValidateAppGatewaySystemInstaller",
		Action: &ValidateAppGatewaySystemInstaller{},
	}

	preparePKI := &task.LocalTask{
		Name:   "PrepareLinkerdPKI",
		Action: &PrepareLinkerdPKI{},
		Retry:  2,
		Delay:  30 * time.Second,
	}

	applyNetworkCRDs := &task.LocalTask{
		Name:   "ApplyNetworkCRDs",
		Action: &ApplyNetworkCRDs{},
		Retry:  2,
		Delay:  30 * time.Second,
	}

	installSystem := &task.LocalTask{
		Name:   "InstallAppGatewaySystem",
		Action: &InstallAppGatewaySystem{},
		Retry:  2,
		Delay:  30 * time.Second,
	}

	checkControlPlane := &task.LocalTask{
		Name: "CheckAppGatewayControlPlaneReady",
		Action: &CheckPodsRunning{
			labels: map[string][]string{
				"linkerd":     {"linkerd.io/control-plane-component"},
				"app-gateway": {"app.kubernetes.io/name=envoy-gateway"},
			},
		},
		Retry: 20,
		Delay: 10 * time.Second,
	}

	m.Tasks = []task.Interface{
		checkInstaller,
		preparePKI,
		applyNetworkCRDs,
		installSystem,
		checkControlPlane,
	}
}

// InstallAppGatewayVendorModule is kept as a compatibility wrapper for existing call sites.
// New module wiring should use InstallAppGatewaySystemModule directly.
type InstallAppGatewayVendorModule struct {
	common.KubeModule
}

func (m *InstallAppGatewayVendorModule) Init() {
	systemModule := &InstallAppGatewaySystemModule{}
	systemModule.KubeModule = m.KubeModule
	systemModule.Init()
	m.Name = systemModule.Name
	m.Tasks = systemModule.Tasks
}

// ValidateAppGatewayInstaller checks release bundle before cluster install (legacy path).
type ValidateAppGatewayInstaller struct {
	common.KubeAction
}

func (t *ValidateAppGatewayInstaller) Execute(runtime connector.Runtime) error {
	if !appGatewayStackEnabled() {
		return nil
	}
	return ValidateAppGatewayInstallerArtifacts(resolveInstallerDir(runtime))
}
