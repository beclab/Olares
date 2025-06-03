package system

import (
	"context"

	"bytetrade.io/web3os/terminusd/internel/watcher"
	"bytetrade.io/web3os/terminusd/pkg/cluster/state"
	changeip "bytetrade.io/web3os/terminusd/pkg/commands/change_ip"
	"bytetrade.io/web3os/terminusd/pkg/commands/uninstall"
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
