package intranet

import "k8s.io/klog/v2"

type Server struct {
	dnsServer   *mDNSServer
	proxyServer *proxyServer
	started     bool
}

type ServerOptions struct {
	Hosts []DNSConfig
}

func (s *Server) Close() {
	if !s.started {
		return
	}

	if s.dnsServer != nil {
		s.dnsServer.Close()
	}

	if s.proxyServer != nil {
		s.proxyServer.Close()
	}

	s.started = false
	klog.Info("Intranet server closed")
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

func (s *Server) IsStarted() bool {
	return s.started
}

func (s *Server) Start(o *ServerOptions) error {
	if s.started {
		return nil
	}

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

	s.started = true
	klog.Info("Intranet server started")
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
