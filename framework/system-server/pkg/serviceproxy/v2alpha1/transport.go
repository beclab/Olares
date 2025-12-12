package v2alpha1

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"time"
)

func initTransport(upstreamCAPool *x509.CertPool, upstreamClientCertPath, upstreamClientKeyPath string) (http.RoundTripper, error) {
	// http.Transport sourced from go 1.10.7
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 1800 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	}

	if upstreamCAPool == nil {
		return transport, nil
	}

	var certKeyPair tls.Certificate
	if len(upstreamClientCertPath) > 0 {
		var err error
		certKeyPair, err = tls.LoadX509KeyPair(upstreamClientCertPath, upstreamClientKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read upstream client cert/key: %w", err)
		}
	}

	// add ssl options
	transport.TLSHandshakeTimeout = 10 * time.Second
	transport.TLSClientConfig = &tls.Config{
		RootCAs: upstreamCAPool,
	}

	if certKeyPair.Certificate != nil {
		transport.TLSClientConfig.Certificates = []tls.Certificate{certKeyPair}
	}

	return transport, nil
}
