package intranet

import (
	"errors"
	"net"
	"os"

	"github.com/beclab/Olares/daemon/pkg/nets"
	"github.com/eball/zeroconf"
	"k8s.io/klog/v2"
)

type DNSConfig struct {
	Domain string
}

type mDNSServer struct {
	server *zeroconf.Server
	hosts  []*DNSConfig
}

func NewMDNSServer() (*mDNSServer, error) {
	s := &mDNSServer{}
	return s, nil
}

func (s *mDNSServer) Close() {
	if s.server != nil {
		s.server.Shutdown()
		klog.Info("Intranet mDNS server closed")
		s.server = nil
	}
}

func (s *mDNSServer) Restart() error {
	if s.server != nil {
		klog.Info("Intranet mDNS server restarting")
		s.Close()
	}

	iface, err := s.findIntranetInterface()
	if err != nil {
		klog.Error("find intranet interface error, ", err)
		return err
	}

	hostname, err := os.Hostname()
	if err != nil {
		klog.Error("get hostname error, ", err)
		return err
	}

	for _, host := range s.hosts {
		klog.Infof("Registering mDNS service for domain: %s", host.Domain)
		// Register the mDNS service
		var err error
		s.server, err = zeroconf.Register(host.Domain, "_http._tcp", "local.", hostname, 80, []string{"txtv=0", "lo=1", "la=0", "path=/"}, []net.Interface{*iface})
		if err != nil {
			klog.Errorf("Failed to register mDNS service for domain %s: %v", host.Domain, err)
			return err
		}
	}
	klog.Info("Intranet mDNS server started")
	return nil
}

func (s *mDNSServer) SetHosts(hosts []*DNSConfig) {
	s.hosts = hosts
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

	// host ip in priority, next is the ethernet ip
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
