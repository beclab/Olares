package terminus

import (
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

// InstallAppGatewaySystemModule installs the app-gateway-system control plane before os-framework
// so Gateway API CRDs and the ingress control plane are ready for Shared app routing.
type InstallAppGatewaySystemModule struct {
	common.KubeModule
}

func (m *InstallAppGatewaySystemModule) Init() {
	logger.InfoInstallationProgress("Installing app-gateway (Envoy Gateway control plane) ...")
	m.Name = "InstallAppGatewaySystemModule"

	checkInstaller := &task.LocalTask{
		Name:   "ValidateAppGatewaySystemInstaller",
		Action: &ValidateAppGatewaySystemInstaller{},
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
				"app-gateway": {"app.kubernetes.io/name=envoy-gateway"},
			},
		},
		Retry: 20,
		Delay: 10 * time.Second,
	}

	m.Tasks = []task.Interface{
		checkInstaller,
		applyNetworkCRDs,
		installSystem,
		checkControlPlane,
	}
}
