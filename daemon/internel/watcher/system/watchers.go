package system

import (
	"context"

	"github.com/beclab/Olares/daemon/internel/watcher"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	changeip "github.com/beclab/Olares/daemon/pkg/commands/change_ip"
	"github.com/beclab/Olares/daemon/pkg/commands/uninstall"
	"github.com/beclab/Olares/daemon/pkg/utils"
	"k8s.io/klog/v2"
)

var _ watcher.Watcher = &systemWatcher{}
var _ watcher.Watcher = &autoRepair{}

type systemWatcher struct {
	watcher.Watcher
}

func NewSystemWatcher() *systemWatcher {
	w := &systemWatcher{}
	return w
}

func (w *systemWatcher) Watch(ctx context.Context) {
	switch state.CurrentState.TerminusState {
	case state.InvalidIpAddress, state.IPChangeFailed:
		// change ip automatically
		cmd := changeip.New()
		_, err := cmd.Execute(ctx, nil)
		if err != nil {
			klog.Error("change ip error, ", err)
		}
	}
}

type autoRepair struct {
	watcher.Watcher
}

func NewAutoRepair() *autoRepair {
	return &autoRepair{}
}

func (w *autoRepair) Watch(ctx context.Context) {
	switch state.CurrentState.TerminusState {
	case state.InstallFailed:
		klog.Info("previous olares installation failed, uninstall it to repair now")
		cmd := uninstall.New()
		_, err := cmd.Execute(ctx, nil)
		if err != nil {
			klog.Error("auto uninstall error, ", err)
		}
	}
}

type bridgeConnectionWatcher struct {
	watcher.Watcher
	ctx    context.Context
	cancel context.CancelFunc
}

func NewBridgeConnectionWatcher() *bridgeConnectionWatcher {
	return &bridgeConnectionWatcher{}
}

func (w *bridgeConnectionWatcher) Watch(ctx context.Context) {
	if c, err := utils.FindBridgeConnection(ctx); err != nil {
		klog.Error("find bridge connection error, ", err)
	} else if c == nil {
		// bridge connection is removed, stop watching
		if w.cancel != nil {
			w.cancel()
			w.cancel = nil
			w.ctx = nil
		}
	} else if w.ctx == nil {
		// bridge connection is back, start watching
		w.ctx, w.cancel = context.WithCancel(context.Background())
		klog.Info("start watching network carrier changes for bridge connection")
		go func() {
			err := utils.ListenNetworkCarrierChanges(w.ctx, func() {
				// disable the overlay gateway supported apps' option for all users
				apps, err := utils.GetOverlayGatewaySupportedApps(ctx, "")
				if err != nil {
					klog.Error("get overlay gateway supported apps error, ", err)
					return
				}

				// network carrier down, disable overlay gateway enabled apps
				for _, app := range apps {
					if app.Enabled {
						// set the app's option to disable overlay gateway
						err = utils.UpdateApplicationSettings(ctx, app.AppResourceName, "enableOverlayGateway", "false")
						if err != nil {
							klog.Error("disable overlay gateway supported app error, ", err)
							return
						}
					}
				}

				// restart the overlay gateway supported apps
				// call restarting from the frontend
				go func() {
					err = utils.RestartOverlayGatewaySupportedApps(ctx, apps)
					if err != nil {
						klog.Error("restart overlay gateway supported apps error, ", err)
					}
				}()
			})

			if err != nil {
				klog.Error("listen network carrier changes error, ", err)
			}
		}()
	}
}
