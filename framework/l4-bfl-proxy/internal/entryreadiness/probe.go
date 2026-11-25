package entryreadiness

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/beclab/l4-bfl-proxy/internal/message"
)

type Gateway struct {
	Host      string
	EntryPort uint32
	AuthPort  uint32
}

type verifyEntry func(ctx context.Context, zone string) error
type entryClient func(*message.EntryMTLSMaterial) (verifyEntry, func(), error)

func (g Gateway) client(material *message.EntryMTLSMaterial) (verifyEntry, func(), error) {
	cert, err := tls.X509KeyPair([]byte(material.ClientCert), []byte(material.ClientKey))
	if err != nil {
		return nil, nil, fmt.Errorf("load Entry client key pair: %w", err)
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM([]byte(material.CA)) {
		return nil, nil, fmt.Errorf("load Entry CA")
	}
	transport := &http.Transport{TLSClientConfig: &tls.Config{
		RootCAs: roots, Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12,
	}}
	client := &http.Client{Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	return func(ctx context.Context, zone string) error {
		return g.verify(ctx, client, zone)
	}, transport.CloseIdleConnections, nil
}

func (g Gateway) verify(ctx context.Context, client *http.Client, zone string) error {
	desktop := "desktop." + zone
	login := "auth." + zone
	for _, probe := range []struct {
		port uint32
		host string
		path string
	}{
		{g.EntryPort, login, "/"},
		{g.EntryPort, desktop, "/"},
		{g.EntryPort, desktop, "/favicon.ico"},
		{g.AuthPort, desktop, "/api/verify"},
	} {
		address := "https://" + net.JoinHostPort(g.Host, strconv.FormatUint(uint64(probe.port), 10)) + probe.path
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
		if err != nil {
			return err
		}
		req.Host = probe.host
		if probe.host == desktop {
			originalPath := probe.path
			if probe.port == g.AuthPort {
				originalPath = "/"
			}
			req.Header.Set("Accept", "text/html")
			req.Header.Set("X-Original-URL", "https://"+desktop+originalPath)
			req.Header.Set("X-Original-Method", http.MethodGet)
			req.Header.Set("X-Forwarded-Host", desktop)
			req.Header.Set("X-Forwarded-Proto", "https")
		}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("probe Entry mTLS on port %d for %s%s: %w", probe.port, probe.host, probe.path, err)
		}
		_ = resp.Body.Close()
		if probe.host == login && resp.StatusCode == http.StatusOK {
			continue
		}
		if probe.host == desktop {
			switch resp.StatusCode {
			case http.StatusUnauthorized, http.StatusForbidden:
				// A browser navigation must receive a login redirect. The
				// separate authorization listener can also deny anonymously.
				if probe.port == g.AuthPort {
					continue
				}
			case http.StatusFound, http.StatusSeeOther, http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
				target, err := url.Parse(resp.Header.Get("Location"))
				if err == nil && target.Scheme == "https" && target.User == nil &&
					strings.EqualFold(target.Hostname(), login) && (target.Port() == "" || target.Port() == "443") {
					continue
				}
				return fmt.Errorf("Entry authorization on port %d for %s%s did not redirect to its HTTPS login host", probe.port, probe.host, probe.path)
			}
		}
		return fmt.Errorf("Entry readiness probe on port %d for %s%s returned HTTP %d", probe.port, probe.host, probe.path, resp.StatusCode)
	}
	return nil
}
