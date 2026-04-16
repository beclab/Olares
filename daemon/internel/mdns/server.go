package mdns

import (
	"context"
	"errors"
	"net"
	"os"
	"strings"

	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/nets"
	"github.com/beclab/Olares/daemon/pkg/tools"
	"github.com/beclab/Olares/daemon/pkg/utils"
	"github.com/eball/zeroconf"
	"k8s.io/klog/v2"
)

const (
	SERVICE_NAME  = "_terminus._tcp"
	INSTANCE_NAME = "olaresd"
)

type serverInf interface {
	Restart() error
	Close()
}

type server struct {
	server       *zeroconf.Server
	port         int
	name         string
	registeredIP string
	serviceName  string
}

type sunshineServer struct {
	server
	ctx context.Context
}

func NewServer(apiPort int) (serverInf, error) {
	s := &server{
		port:        apiPort,
		serviceName: SERVICE_NAME,
		name:        INSTANCE_NAME + "-" + tools.RandomString(6),
	}
	return s, s.Restart()
}

func NewSunShineProxyWithoutStart(ctx context.Context) serverInf {
	s := &sunshineServer{server: server{port: 47989, name: "", serviceName: "_nvstream._tcp"}, ctx: ctx}
	return s
}

func (s *server) Close() {
	if s.server != nil {
		klog.Info("mDNS server shutdown ")
		s.server.Shutdown()
		s.registeredIP = "" // clear the registered IP
		s.server = nil
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

func (s *sunshineServer) Restart() error {
	switch state.CurrentState.TerminusState {
	case state.NotInstalled, state.Uninitialized, state.InitializeFailed, state.IPChanging:
		// Stop the sunshine mdns if it's running
		s.server.Close()
		return nil
	default:
		client, err := utils.GetKubeClient()
		if err != nil {
			klog.Error("failed to get kube client: ", err)
			return nil
		}

		_, _, role, err := utils.GetThisNodeName(s.ctx, client)
		if err != nil {
			klog.Error("failed to get this node role: ", err)
			return nil
		}

		if role != "master" {
			// Only master nodes run the sunshine mdns proxy
			s.server.Close()
			return nil
		}

		return s.server.Restart()
	}
}
