package intranet

type server struct {
	dnsServer   *mDNSServer
	proxyServer *proxyServer
}

func (s *server) Close() {
	if s.dnsServer != nil {
		s.dnsServer.Close()
		s.dnsServer = nil
	}

	if s.proxyServer != nil {
	}
}

func NewServer() (*server, error) {
	dnsServer, err := NewMDNSServer()
	if err != nil {
		return nil, err
	}

	return &server{
		dnsServer: dnsServer,
	}, nil
}

func (s *server) Start() error {
	if s.dnsServer != nil {
		// test
		s.dnsServer.SetHosts([]*DNSConfig{
			{Domain: "myapp.liuyu.olares"},
			{Domain: "desktop.liuyu.olares"},
		})
		// end test
		return s.dnsServer.Restart()
	}
	return nil
}
