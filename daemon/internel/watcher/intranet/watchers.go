package intranet

import (
	"context"

	"github.com/beclab/Olares/daemon/internel/intranet"
	"github.com/beclab/Olares/daemon/internel/watcher"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"k8s.io/klog/v2"
)

var _ watcher.Watcher = &applicationWatcher{}

type applicationWatcher struct {
	intranetServer *intranet.Server
}

func NewApplicationWatcher() *applicationWatcher {
	return &applicationWatcher{}
}

func (w *applicationWatcher) Watch(ctx context.Context) {
	switch state.CurrentState.TerminusState {
	case state.NotInstalled, state.Uninitialized, state.InitializeFailed:
		// Stop the intranet server if it's running
		if w.intranetServer != nil {
			w.intranetServer.Close()
			w.intranetServer = nil
			klog.Info("Intranet server stopped due to cluster state: ", state.CurrentState.TerminusState)
		}
	default:
		if w.intranetServer == nil {
			var err error
			w.intranetServer, err = intranet.NewServer()
			if err != nil {
				klog.Error("failed to create intranet server: ", err)
				return
			}

			w.intranetServer.Start()
		}
	}
}
