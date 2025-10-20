package intranet

import (
	"context"
	"strings"

	"github.com/beclab/Olares/daemon/internel/intranet"
	"github.com/beclab/Olares/daemon/internel/watcher"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/utils"
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

		}

		err := w.loadServerConfig(ctx)
		if err != nil {
			klog.Error("load intranet server config error, ", err)
			return
		}
	}
}

func (w *applicationWatcher) loadServerConfig(ctx context.Context) error {
	if w.intranetServer == nil {
		klog.Warning("intranet server is nil")
		return nil
	}

	urls, err := utils.GetApplicationUrlAll(ctx)
	if err != nil {
		klog.Error("get application urls error, ", err)
		return err
	}

	var hosts []intranet.DNSConfig
	for _, url := range urls {
		urlToken := strings.Split(url, ".")
		if len(urlToken) > 2 {
			domain := strings.Join([]string{urlToken[0], urlToken[1], "olares", "local"}, ".")

			hosts = append(hosts, intranet.DNSConfig{
				Domain: domain,
			})
		}
	}

	options := &intranet.ServerOptions{
		Hosts: hosts,
	}

	// reload intranet server config
	return w.intranetServer.Reload(options)
}
