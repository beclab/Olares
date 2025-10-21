package intranet

import (
	"context"
	"fmt"
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
		client, err := utils.GetKubeClient()
		if err != nil {
			klog.Error("failed to get kube client: ", err)
			return
		}

		_, _, role, err := utils.GetThisNodeName(ctx, client)
		if err != nil {
			klog.Error("failed to get this node role: ", err)
			return
		}

		if role != "master" {
			// Only master nodes run the intranet server
			return
		}

		if w.intranetServer == nil {
			var err error
			w.intranetServer, err = intranet.NewServer()
			if err != nil {
				klog.Error("failed to create intranet server: ", err)
				return
			}

		}

		o, err := w.loadServerConfig(ctx)
		if err != nil {
			klog.Error("load intranet server config error, ", err)
			return
		}

		if w.intranetServer.IsStarted() {
			// Reload the intranet server config
			err = w.intranetServer.Reload(o)
			if err != nil {
				klog.Error("reload intranet server config error, ", err)
				return
			}
			klog.Info("Intranet server config reloaded")
		} else {
			// Start the intranet server
			err = w.intranetServer.Start(o)
			if err != nil {
				klog.Error("start intranet server error, ", err)
				return
			}
			klog.Info("Intranet server started")
		}
	}
}

func (w *applicationWatcher) loadServerConfig(ctx context.Context) (*intranet.ServerOptions, error) {
	if w.intranetServer == nil {
		klog.Warning("intranet server is nil")
		return nil, nil
	}

	urls, err := utils.GetApplicationUrlAll(ctx)
	if err != nil {
		klog.Error("get application urls error, ", err)
		return nil, err
	}

	var hosts []intranet.DNSConfig
	for _, url := range urls {
		urlToken := strings.Split(url, ".")
		if len(urlToken) > 2 {
			domain := strings.Join([]string{urlToken[0], urlToken[1], "olares"}, ".")

			hosts = append(hosts, intranet.DNSConfig{
				Domain: domain,
			})
		}
	}

	dynamicClient, err := utils.GetDynamicClient()
	if err != nil {
		err = fmt.Errorf("failed to get dynamic client: %v", err)
		klog.Error(err.Error())
		return nil, err
	}

	users, err := utils.ListUsers(ctx, dynamicClient)
	if err != nil {
		err = fmt.Errorf("failed to list users: %v", err)
		klog.Error(err.Error())
		return nil, err
	}

	adminUser, err := utils.GetAdminUser(ctx, dynamicClient)
	if err != nil {
		err = fmt.Errorf("failed to get admin user: %v", err)
		klog.Error(err.Error())
		return nil, err
	}

	for _, user := range users {
		domain := fmt.Sprintf("%s.olares", user.GetName())
		hosts = append(hosts, intranet.DNSConfig{
			Domain: domain,
		})

		domain = fmt.Sprintf("desktop.%s.olares", user.GetName())
		hosts = append(hosts, intranet.DNSConfig{
			Domain: domain,
		})

		domain = fmt.Sprintf("auth.%s.olares", user.GetName())
		hosts = append(hosts, intranet.DNSConfig{
			Domain: domain,
		})

		if user.GetAnnotations()["bytetrade.io/is-ephemeral"] == "true" {
			domain = fmt.Sprintf("wizard-%s.%s.olares", user.GetName(), adminUser.GetName())
			hosts = append(hosts, intranet.DNSConfig{
				Domain: domain,
			})
		}
	}

	options := &intranet.ServerOptions{
		Hosts: hosts,
	}

	// reload intranet server config
	return options, nil
}
