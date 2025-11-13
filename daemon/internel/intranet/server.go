package intranet

import (
	"fmt"

	"k8s.io/klog/v2"
)

type Server struct {
	dnsServer   *mDNSServer
	proxyServer *proxyServer
	dsrProxy    *DSRProxy
	started     bool
}

type ServerOptions struct {
	Hosts             []DNSConfig
	NodeIp            string
	NodeIface         string
	DnsPodIp          string
	DnsPodMac         string
	DnsPodCalicoIface string
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

	if s.dsrProxy != nil {
		s.dsrProxy.Stop()
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
		dsrProxy:    NewDSRProxy(),
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

	if s.dsrProxy != nil {
		err := s.dsrProxy.Start()
		if err != nil {
			klog.Error("start intranet dsr proxy error, ", err)
			return err
		}
	}

	s.started = true
	klog.Info("Intranet server started")
	return nil
}

func (s *Server) Reload(o *ServerOptions) error {
	var errs []error
	if s.dnsServer != nil {
		s.dnsServer.SetHosts(o.Hosts, false)
		err := s.dnsServer.StartAll()
		if err != nil {
			klog.Error("reload intranet dns server error, ", err)
			errs = append(errs, err)
		}
	}

	if s.dsrProxy != nil {
		err := s.dsrProxy.WithBackend(o.DnsPodIp, o.DnsPodMac)
		if err != nil {
			klog.Error("reload dns dsr proxy error, ", err)
			errs = append(errs, err)
		}

		if err == nil {
			err = s.dsrProxy.WithCalicoInterface(o.DnsPodCalicoIface)
			if err != nil {
				klog.Error("reload dns dsr proxy backend interfaces error, ", err)
				errs = append(errs, err)
			}
		}

		if err == nil {
			err = s.dsrProxy.WithVIP(o.NodeIp, o.NodeIface)
			if err != nil {
				klog.Error("reload dns dsr proxy vip interface error, ", err)
				errs = append(errs, err)
			}
		}

		if err == nil {
			err = s.dsrProxy.regonfigure()
			if err != nil {
				klog.Error("reload dns dsr proxy regonfigure error, ", err)
				errs = append(errs, err)
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("reload intranet server with %d errors", len(errs))
	}

	klog.Info("Intranet server reloaded")
	return nil
}
