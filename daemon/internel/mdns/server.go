package mdns

import (
	"errors"
	"net"
	"os"
	"strings"

	"github.com/beclab/Olares/daemon/pkg/nets"
	"github.com/beclab/Olares/daemon/pkg/tools"
	"github.com/eball/zeroconf"
	"k8s.io/klog/v2"
)

const (
	SERVICE_NAME  = "_terminus._tcp"
	INSTANCE_NAME = "olaresd"
)

type server struct {
	server       *zeroconf.Server
	port         int
	name         string
	registeredIP string
	serviceName  string
}

func NewServer(apiPort int) (*server, error) {
	s := &server{
		port:        apiPort,
		serviceName: SERVICE_NAME,
		name:        INSTANCE_NAME + "-" + tools.RandomString(6),
	}
	return s, s.Restart()
}

func NewSunShineProxyWithoutStart() *server {
	s := &server{port: 47989, name: "", serviceName: "_nvstream._tcp"}
	return s
}

func (s *server) Close() {
	if s.server != nil {
		klog.Info("mDNS server shutdown ")
		s.server.Shutdown()
		s.registeredIP = "" // clear the registered IP
	}
}

func (s *server) Restart() error {
	ips, err := nets.GetInternalIpv4Addr()
	if err != nil {
		return err
	}

	if len(ips) == 0 {
		return errors.New("cannot get any ip on server")
	}

	hostIp, err := nets.GetHostIp()
	if err != nil {
		klog.Error("get host ip error, ", err)
	}

	// host ip in priority, next is the ethernet ip
	var (
		iface *net.Interface
		ip    string
	)

	for _, i := range ips {
		if i.IP == hostIp {
			iface = i.Iface
			ip = i.IP
			break
		}
	}
	if iface == nil {
		iface = ips[0].Iface
		ip = ips[0].IP
	}

	hostname, err := os.Hostname()
	if err != nil {
		klog.Error("cannot get hostname, ", err)
	} else {
		iptoken := strings.Split(ip, ".")
		hostname = strings.Join([]string{hostname, iptoken[len(iptoken)-1]}, "-")
	}

	if s.registeredIP != ip {
		if s.server != nil {
			s.Close()
		}

		s.registeredIP = ip
		instanceName := s.name
		if instanceName == "" {
			instanceName = hostname
		}

		s.server, err = zeroconf.RegisterAll(instanceName, s.serviceName, "local.", hostname, s.port, []string{""}, []net.Interface{*iface}, false, false, false)
		if err != nil {
			klog.Error("create mdns server error, ", err)
			return err
		}

		klog.Info("mDNS server started, ", s.serviceName)
	}

	return nil
}
