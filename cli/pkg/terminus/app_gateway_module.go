package terminus

import (
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

// InstallAppGatewayVendorModule installs Linkerd (CRDs + control plane) + Envoy Gateway before os-framework chart
// so Gateway API CRDs exist when app-gateway deploy templates are applied.
type InstallAppGatewayVendorModule struct {
	common.KubeModule
}

func (m *InstallAppGatewayVendorModule) Init() {
	logger.InfoInstallationProgress("Installing unified ingress (Linkerd + Envoy Gateway + app-gateway) ...")
	m.Name = "InstallAppGatewayVendorModule"

	checkInstaller := &task.LocalTask{
		Name:   "ValidateAppGatewayInstaller",
		Action: &ValidateAppGatewayInstaller{},
	}

	installVendor := &task.LocalTask{
		Name:   "InstallAppGatewayVendor",
		Action: &InstallAppGatewayVendor{},
		Retry:  2,
		Delay:  30 * time.Second,
	}

	waitEG := &task.LocalTask{
		Name:   "WaitAppGatewayReady",
		Action: &WaitAppGatewayReady{},
		Retry:  30,
		Delay:  10 * time.Second,
	}

	installChart := &task.LocalTask{
		Name:   "InstallAppGatewayChart",
		Action: &InstallAppGatewayChart{},
		Retry:  2,
		Delay:  20 * time.Second,
	}

	m.Tasks = []task.Interface{
		checkInstaller,
		installVendor,
		waitEG,
		installChart,
	}
}

// ValidateAppGatewayInstaller checks release bundle before cluster install (standard Olares install path).
type ValidateAppGatewayInstaller struct {
	common.KubeAction
}

func (t *ValidateAppGatewayInstaller) Execute(runtime connector.Runtime) error {
	if !appGatewayStackEnabled() {
		return nil
	}
	return ValidateAppGatewayInstallerArtifacts(resolveInstallerDir(runtime))
}
