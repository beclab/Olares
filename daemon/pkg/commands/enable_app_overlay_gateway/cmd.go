package enableappoverlaygateway

import (
	"context"
	"errors"

	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/utils"
)

type enableAppOverlayGateway struct {
	commands.Operation
}

var _ commands.Interface = &enableAppOverlayGateway{}

func New() commands.Interface {
	return &enableAppOverlayGateway{
		Operation: commands.Operation{
			Name: commands.EnableAppOverlayGateway,
		},
	}
}

func (d *enableAppOverlayGateway) Execute(ctx context.Context, p any) (res any, err error) {
	param, ok := p.(*Param)
	if !ok {
		err = errors.New("invalid param")
		return
	}

	// enable the overlay gateway supported apps' option for the user
	apps, err := utils.GetOverlayGatewaySupportedApps(ctx, param.User)
	if err != nil {
		return nil, err
	}

	for _, app := range apps {
		if app.AppID != param.AppID {
			continue
		}

		if !app.SharedApp && app.Owner != param.User { // double check the app is owned by the user, the user parameter may be empty
			continue
		}

		if app.Enabled {
			return nil, errors.New("app is already enabled")
		}
		// set the app's option to disable overlay gateway
		err = utils.UpdateApplicationSettings(ctx, app.AppResourceName, "enableOverlayGateway", "true")
		if err != nil {
			return nil, err
		}

		// restart in a separate goroutine, cause the restarting process may take a while
		// call restarting from the frontend
		// go func() {
		// 	err = utils.RestartOverlayGatewaySupportedApps(ctx, []utils.OverlayGatewaySupportedApp{app})
		// 	if err != nil {
		// 		klog.Error("restart overlay gateway supported apps error, ", err)
		// 	}
		// }()

		return nil, nil
	}

	return nil, errors.New("app not found")
}
