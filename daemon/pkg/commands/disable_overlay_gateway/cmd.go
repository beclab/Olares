package disableoverlaygateway

import (
	"context"
	"errors"
	"os"
	"os/exec"

	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/utils"
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
	param, ok := p.(*Param)
	if !ok {
		err = errors.New("invalid param")
		return
	}

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

	// disable the overlay gateway supported apps' option
	apps, err := utils.GetOverlayGatewaySupportedApps(ctx, param.User)
	if err != nil {
		return nil, err
	}

	for _, app := range apps {
		if app.Enabled {
			// set the app's option to disable overlay gateway
		}
	}

	return nil, nil
}
