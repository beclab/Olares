package intranet

import (
	"errors"
	"net"
	"slices"
	"strings"

	"github.com/beclab/Olares/daemon/pkg/nets"
	"github.com/eball/zeroconf"
	"k8s.io/klog/v2"
)

type DNSConfig struct {
	Domain string
}

type instanceServer struct {
	queryServer *zeroconf.Server
	host        *DNSConfig
	aliases     []string
}

type mDNSServer struct {
	servers map[string]*instanceServer
}

func NewMDNSServer() (*mDNSServer, error) {
	s := &mDNSServer{
		servers: make(map[string]*instanceServer),
	}
	return s, nil
}

func (s *mDNSServer) Close() {
	if s.servers != nil {
		for host, server := range s.servers {
			if server == nil {
				continue
			}

			// Shutdown the mDNS server
			server.queryServer.Shutdown()
			s.servers[host] = nil
			klog.Info("Intranet mDNS server closed, ", host)
		}
	}
}

func (s *mDNSServer) StartAll() error {
	iface, err := s.findIntranetInterface()
	if err != nil {
		klog.Error("find intranet interface error, ", err)
		return err
	}

	for domain := range s.servers {
		if s.servers[domain] != nil {
			continue
		}

		klog.Infof("Registering mDNS service for domain: %s", domain)
		// Register the mDNS service
		var err error
		server, err := zeroconf.Register("olares", "_http._tcp", "local.", domain, 80, []string{"txtv=0", "lo=1", "la=0", "path=/"}, []net.Interface{*iface})
		if err != nil {
			klog.Errorf("Failed to register mDNS service for domain %s: %v", domain, err)
			return err
		}

		// add host alias
		domainTokens := strings.Split(domain, ".")
		alias := []string{strings.Join(domainTokens, "-") + ".local."}

		// TODO: add more alias if needed
		klog.Info("add host alias, ", alias[0])
		server.AddHostAlias(alias[0])

		s.servers[domain] = &instanceServer{
			queryServer: server,
			host:        &DNSConfig{Domain: domain},
		}
	}
	klog.V(8).Info("Intranet mDNS server started")
	return nil
}

// SetHosts sets the hosts for the mDNS server
// if reset is true, it will remove all existing hosts before adding new ones
func (s *mDNSServer) SetHosts(hosts []DNSConfig, reset bool) {
	for _, host := range hosts {
		if host.Domain == "" {
			continue
		}

		if server, exists := s.servers[host.Domain]; !exists {
			s.servers[host.Domain] = nil
		} else {

			if reset {
				server.queryServer.Shutdown()
				s.servers[host.Domain] = nil
			}
		}
	}

	// remove not exist hosts
	for domain := range s.servers {
		if slices.ContainsFunc(hosts, func(a DNSConfig) bool {
			return a.Domain == domain
		}) {
			continue
		}

		klog.Info("removing domain ", domain)
		s.servers[domain].queryServer.Shutdown()
		delete(s.servers, domain)
	}
}

func (s *mDNSServer) findIntranetInterface() (*net.Interface, error) {
	ips, err := nets.GetInternalIpv4Addr()
	if err != nil {
		return nil, err
	}

	if len(ips) == 0 {
		return nil, errors.New("cannot get any ip on server")
	}

	hostIp, err := nets.GetHostIp()
	if err != nil {
		klog.Error("get host ip error, ", err)
	}

	// host ip in priority, next is the ethernet ip-
	var (
		iface *net.Interface
	)

	for _, i := range ips {
		if i.IP == hostIp {
			iface = i.Iface
			break
		}
	}

	if iface == nil {
		iface = ips[0].Iface
	}

	return iface, nil
}
