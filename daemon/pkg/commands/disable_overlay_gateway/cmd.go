package disableoverlaygateway

import (
	"context"
	"os"
	"os/exec"

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

func (d *disableOverlayGateway) Execute(ctx context.Context, p any) (res any, err error) {
	// disable the bridge connection
	err = utils.ResetBridgeConnection(ctx)
	if err != nil {
		return nil, err
	}

	// turn off the CNI-DHCP service
	cmd := exec.CommandContext(ctx, "systemctl", "disable", "--now", "cni-dhcp.service")
	cmd.Env = os.Environ()
	_, err = cmd.Output()
	if err != nil {
		return nil, err
	}

	// disable the overlay gateway supported apps' option for all users
	apps, err := utils.GetOverlayGatewaySupportedApps(ctx, "")
	if err != nil {
		return nil, err
	}

	for _, app := range apps {
		if app.Enabled {
			// set the app's option to disable overlay gateway
			err = utils.UpdateApplicationSettings(ctx, app.AppResourceName, "enableOverlayGateway", "false")
			if err != nil {
				return nil, err
			}
		}
	}

	// restart the overlay gateway supported apps
	// restart in a separate goroutine, cause the restarting process may take a while
	go func() {
		err = utils.RestartOverlayGatewaySupportedApps(ctx, apps)
		if err != nil {
			klog.Error("restart overlay gateway supported apps error, ", err)
		}
	}()

	return nil, nil
}
