package intranet

import "k8s.io/klog/v2"

type Server struct {
	dnsServer   *mDNSServer
	proxyServer *proxyServer
}

type ServerOptions struct {
	Hosts []DNSConfig
}

func (s *Server) Close() {
	if s.dnsServer != nil {
		s.dnsServer.Close()
	}

	if s.proxyServer != nil {
		s.proxyServer.Close()
	}
}

func NewServer() (*Server, error) {
	dnsServer, err := NewMDNSServer()
	if err != nil {
		return nil, err
	}

	proxyServer, err := NewProxyServer()
	if err != nil {
		return nil, err
	}

	return &Server{
		dnsServer:   dnsServer,
		proxyServer: proxyServer,
	}, nil
}

func (s *Server) Start(o *ServerOptions) error {
	if s.dnsServer != nil {
		s.dnsServer.SetHosts(o.Hosts, true)
		err := s.dnsServer.StartAll()
		if err != nil {
			klog.Error("start intranet dns server error, ", err)
			return err
		}
	}

	if s.proxyServer != nil {
		err := s.proxyServer.Start()
		if err != nil {
			klog.Error("start intranet proxy server error, ", err)
			return err
		}
	}

	return nil
}

func (s *Server) Reload(o *ServerOptions) error {
	if s.dnsServer != nil {
		s.dnsServer.SetHosts(o.Hosts, false)
		err := s.dnsServer.StartAll()
		if err != nil {
			klog.Error("reload intranet dns server error, ", err)
			return err
		}
	}

	klog.Info("Intranet server reloaded")
	return nil
}
