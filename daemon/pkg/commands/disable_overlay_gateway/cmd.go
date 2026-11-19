package disableoverlaygateway

import (
	"context"

	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/utils"
	"k8s.io/klog/v2"
)

type disableOverlayGateway struct {
	commands.Operation
}

var _ commands.Interface = &disableOverlayGateway{}

func New() commands.Interface {
	return &disableOverlayGateway{
		Operation: commands.Operation{
			Name: commands.DisableOverlayGateway,
		},
	}
}

// Execute withdraws the overlay gateway on this node: every application's
// switch is cleared and the desired state is removed. The alternative name and
// the CNI DHCP daemon stay in place; neither affects the host network, and the
// daemon keeps renewing leases for Pods until they are recreated.
func (d *disableOverlayGateway) Execute(ctx context.Context, p any) (res any, err error) {
	apps, err := utils.GetOverlayGatewaySupportedApps(ctx, "")
	if err != nil {
		klog.Errorf("overlay gateway disable: list supported apps failed: %v", err)
		return nil, err
	}
	for _, app := range apps {
		if !app.Enabled {
			continue
		}
		if err := utils.UpdateApplicationSettings(ctx, app.AppResourceName, "enableOverlayGateway", "false"); err != nil {
			klog.Errorf("overlay gateway disable: clear enableOverlayGateway for %s failed: %v", app.AppResourceName, err)
			return nil, err
		}
	}
	if err := utils.SetOverlayGatewayDesired(false); err != nil {
		klog.Errorf("overlay gateway disable: %v", err)
		return nil, err
	}
	klog.Info("overlay gateway disabled")
	return nil, nil
}
