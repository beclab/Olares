package intranet

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/beclab/Olares/daemon/internel/intranet"
	"github.com/beclab/Olares/daemon/internel/watcher"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/nets"
	"github.com/beclab/Olares/daemon/pkg/utils"
	"github.com/miekg/dns"
	"k8s.io/klog/v2"
)

var _ watcher.Watcher = &applicationWatcher{}

type applicationWatcher struct {
	intranetServer *intranet.Server
	// lastAppliedSig is the signature of the ServerOptions last successfully
	// applied to the intranet server. It lets us skip the per-tick Reload
	// (DNS SetHosts/StartAll + DSR reconfigure) when nothing changed.
	lastAppliedSig string
}

func NewApplicationWatcher() *applicationWatcher {
	return &applicationWatcher{}
}

// optionsSignature returns a stable hash of the ServerOptions so we can detect
// whether anything that affects the intranet server config has changed.
func optionsSignature(o *intranet.ServerOptions) string {
	if o == nil {
		return ""
	}
	domains := make([]string, 0, len(o.Hosts))
	for _, h := range o.Hosts {
		domains = append(domains, h.Domain)
	}
	sort.Strings(domains)

	h := sha256.New()
	for _, d := range domains {
		h.Write([]byte(d))
		h.Write([]byte{0})
	}
	fmt.Fprintf(h, "|%s|%s|%s|%s|%s", o.NodeIp, o.NodeIface, o.DnsPodIp, o.DnsPodMac, o.DnsPodCalicoIface)
	return hex.EncodeToString(h.Sum(nil))
}

func (w *applicationWatcher) Watch(ctx context.Context) {
	switch state.CurrentState.TerminusState {
	case state.NotInstalled, state.Uninitialized, state.InitializeFailed, state.IPChanging:
		// Stop the intranet server if it's running
		if w.intranetServer != nil {
			w.intranetServer.Close()
			w.intranetServer = nil
			w.lastAppliedSig = ""
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
			w.lastAppliedSig = ""
		}

		o, err := w.loadServerConfig(ctx, nodeIp)
		if err != nil {
			klog.Error("load intranet server config error, ", err)
			return
		}

		sig := optionsSignature(o)

		if w.intranetServer.IsStarted() {
			// Skip the reconfigure work when nothing relevant changed.
			if sig == w.lastAppliedSig {
				klog.V(8).Info("Intranet server config unchanged, skip reload")
				return
			}
			// Reload the intranet server config
			err = w.intranetServer.Reload(o)
			if err != nil {
				klog.Error("reload intranet server config error, ", err)
				return
			}
			w.lastAppliedSig = sig
			klog.V(8).Info("Intranet server config reloaded")
		} else {
			// Start the intranet server. Start only brings up the DNS/proxy/DSR
			// goroutines; the DSR backend, VIP and reconfigure are applied by
			// Reload. Leave lastAppliedSig empty so the next tick performs that
			// first Reload instead of treating the server as fully configured.
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

var adguardDnsPodIp string
var adguardHealth bool

func (w *applicationWatcher) loadDnsPodConfig(ctx context.Context, o *intranet.ServerOptions) error {
	// try to find adguard dns pod ip and mac
	dnsPods, err := utils.ListPods(ctx)
	if err != nil {
		klog.Error("list pods error, ", err)
		return err
	}

	var dnsPodIp, dnsPodMac, calicoRouteIface string
	const adguardDnsAppLabel = "applications.app.bytetrade.io/name"
	for _, pod := range dnsPods {
		switch {
		case pod.Labels[adguardDnsAppLabel] == "adguardhome":
			dnsPodIp = pod.Status.PodIP

			// try to connect adguard dns pod port 53 to verify it's running
			if adguardDnsPodIp != dnsPodIp || !adguardHealth {
				adguardDnsPodIp = dnsPodIp
				err := checkHealth(dnsPodIp)
				if err != nil {
					klog.Warning("dial adguard dns pod tcp 53 error, ", err)
					adguardHealth = false
				} else {
					adguardHealth = true
				}
			}

			if adguardHealth {
				dnsPodMac, calicoRouteIface, err = getPodNeighborInfo(dnsPodIp)
				if err != nil {
					klog.Error("get adguard dns pod mac by ip error, ", err)
					return err
				}

				// found adguard dns pod
				o.DnsPodIp = dnsPodIp
				o.DnsPodMac = dnsPodMac
				o.DnsPodCalicoIface = calicoRouteIface
				return nil
			}

		case pod.Labels["k8s-app"] == "kube-dns":
			dnsPodIp = pod.Status.PodIP
			dnsPodMac, calicoRouteIface, err = getPodNeighborInfo(dnsPodIp)
			if err != nil {
				klog.Error("get adguard dns pod mac by ip error, ", err)
				return err
			}
		}

	} // end for pods

	// not found adguard dns pod, but core dns pod exists
	if dnsPodIp != "" {
		o.DnsPodIp = dnsPodIp
		o.DnsPodMac = dnsPodMac
		o.DnsPodCalicoIface = calicoRouteIface
	}

	return nil
}

func checkHealth(server string) error {
	c := new(dns.Client)
	c.Timeout = time.Second

	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn("coredns.kube-system.svc.cluster.local."), dns.TypeA)
	msg.RecursionDesired = true

	_, _, err := c.Exchange(msg, server+":53")
	return err
}
