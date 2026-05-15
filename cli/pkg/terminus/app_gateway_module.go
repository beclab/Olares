package terminus

import (
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

// InstallAppGatewayVendorModule installs Linkerd + Envoy Gateway before os-framework chart
// so Gateway API CRDs exist when app-gateway deploy templates are applied.
type InstallAppGatewayVendorModule struct {
	common.KubeModule
}

func (m *InstallAppGatewayVendorModule) Init() {
	logger.InfoInstallationProgress("Installing app-gateway platform (Linkerd + Envoy Gateway) ...")
	m.Name = "InstallAppGatewayVendorModule"

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
		installVendor,
		waitEG,
		installChart,
	}
}
