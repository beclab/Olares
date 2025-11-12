package intranet

import (
	"context"
	"fmt"
	"strings"

	"github.com/beclab/Olares/daemon/internel/intranet"
	"github.com/beclab/Olares/daemon/internel/watcher"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/nets"
	"github.com/beclab/Olares/daemon/pkg/utils"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

		_, nodeIp, role, err := utils.GetThisNodeName(ctx, client)
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

		o, err := w.loadServerConfig(ctx, nodeIp)
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

func (w *applicationWatcher) loadServerConfig(ctx context.Context, nodeIp string) (*intranet.ServerOptions, error) {
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

	nodeIface, err := nets.GetInterfaceByIp(nodeIp)
	if err != nil {
		klog.Error("get node interface by ip error, ", err)
		return nil, err
	}

	options := &intranet.ServerOptions{
		Hosts:     hosts,
		NodeIp:    nodeIp,
		NodeIface: nodeIface.Name,
	}

	err = w.loadDnsPodConfig(ctx, options)
	if err != nil {
		klog.Error("load dns pod config error, ", err)
		return nil, err
	}

	// reload intranet server config
	return options, nil
}

func (w *applicationWatcher) loadDnsPodConfig(ctx context.Context, o *intranet.ServerOptions) error {
	// try to find adguard dns pod ip and mac
	k8sClient, err := utils.GetKubeClient()
	if err != nil {
		klog.Error("get kube client error, ", err)
		return err
	}

	dnsPods, err := k8sClient.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error("list pods error, ", err)
		return err
	}

	var dnsPodIp, dnsPodMac, calicoRouteIface string
	const adguardDnsAppLabel = "applications.app.bytetrade.io/name"
	for _, pod := range dnsPods.Items {
		switch {
		case pod.Labels[adguardDnsAppLabel] == "adguardhome", pod.Labels["k8s-app"] == "kube-dns":
			dnsPodIp = pod.Status.PodIP
			dnsPodMac, calicoRouteIface, err = getPodNeighborInfo(dnsPodIp)
			if err != nil {
				klog.Error("get adguard dns pod mac by ip error, ", err)
				return err
			}
		}

		if pod.Labels[adguardDnsAppLabel] == "adguardhome" {
			o.DnsPodIp = dnsPodIp
			o.DnsPodMac = dnsPodMac
			o.DnsPodCalicoIface = calicoRouteIface
			return nil
		}
	}

	// not found adguard dns pod, but core dns pod exists
	if dnsPodIp != "" {
		o.DnsPodIp = dnsPodIp
		o.DnsPodMac = dnsPodMac
		o.DnsPodCalicoIface = calicoRouteIface
	}

	return nil
}

func getPodNeighborInfo(podIp string) (mac, iface string, err error) {
	// family: unix.AF_INET for IPv4, unix.AF_INET6 for IPv6
	neighs, err := netlink.NeighList(0, unix.AF_INET) // 0 => all links
	if err != nil {
		klog.Error("list neighbor error, ", err)
		return
	}

	for _, n := range neighs {
		if n.IP.String() == podIp {
			mac = n.HardwareAddr.String()
			if mac == "<nil>" {
				mac = ""
			}

			if link, err := netlink.LinkByIndex(n.LinkIndex); err == nil {
				iface = link.Attrs().Name
			}

			return
		}
	}

	return "", "", fmt.Errorf("not found pod neighbor info for ip %s", podIp)
}
