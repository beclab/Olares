package routecontrol

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/gateway"
	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *SharedRouteReconciler) servingReadiness(ctx context.Context, srr *srrv1alpha1.SharedRouteRegistry, route string) error {
	if err := servingReadiness(ctx, r.Client, GatewayRef{}, srr, route); err != nil {
		return err
	}
	if !isSystemDesktopRoute(srr) {
		return nil
	}
	reader := r.ReadinessReader
	if reader == nil {
		reader = r.Client
	}
	probe := r.desktopProbe
	if probe == nil {
		probe = probeDesktopEntry
	}
	return probe(ctx, reader, srr)
}

// Probe only the fixed Entry Service using verified Entry TLS. No user session
// or supplied tenant header is sent: the trusted gateway must establish the
// tenant context and reject anonymous Desktop navigation with a tenant-bound
// login redirect. Public portals require HTTPS; platform local aliases also
// retain their existing HTTP login contract.
func probeDesktopEntry(ctx context.Context, reader client.Reader, srr *srrv1alpha1.SharedRouteRegistry) error {
	return probeDesktopEntryWithDialer(ctx, reader, srr, (&net.Dialer{Timeout: 3 * time.Second}).DialContext)
}

func probeDesktopEntryWithDialer(ctx context.Context, reader client.Reader, srr *srrv1alpha1.SharedRouteRegistry, dial func(context.Context, string, string) (net.Conn, error)) error {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	if !isSystemDesktopRoute(srr) || srr.UID == "" {
		return fmt.Errorf("Desktop readiness requires a managed tenant-bound entrance")
	}
	var zone corev1.ConfigMap
	if err := reader.Get(ctx, client.ObjectKey{Namespace: srr.Namespace, Name: zoneSSLConfigMapName}, &zone); err != nil {
		return err
	}
	z := strings.ToLower(strings.TrimSpace(zone.Data["zone"]))
	desktop := "desktop." + z
	// Match the complete platform-generated domain set, including the local
	// alias, without depending on the API's list ordering. Checking only for
	// the primary domain would permit additional, unbound tenant authorities.
	expected := gateway.SystemPlaneHostPatterns(gateway.SystemPlaneService{Name: "desktop"}, z, srr.Annotations[gateway.AnnotationExtAuthUser], false)
	if z == "" || len(validation.IsDNS1123Subdomain(desktop)) != 0 || !sets.New(srr.Spec.HostPatterns...).Equal(sets.New(expected...)) {
		return fmt.Errorf("Desktop readiness requires its exact managed zone authorities")
	}
	var service corev1.Service
	if err := reader.Get(ctx, client.ObjectKey{Namespace: defaultGatewayNS, Name: entryMeshService.ServiceName}, &service); err != nil {
		return err
	}
	ip, err := netip.ParseAddr(service.Spec.ClusterIP)
	if err != nil || !ip.IsGlobalUnicast() || service.Spec.Type != corev1.ServiceTypeClusterIP || !service.DeletionTimestamp.IsZero() {
		return fmt.Errorf("Desktop readiness requires the fixed Entry ClusterIP Service")
	}
	var material corev1.Secret
	if err := reader.Get(ctx, client.ObjectKey{Namespace: defaultGatewayNS, Name: "olares-l4-agw-client"}, &material); err != nil {
		return err
	}
	cert, err := tls.X509KeyPair(material.Data["client.crt"], material.Data["client.key"])
	if err != nil {
		return fmt.Errorf("Desktop readiness has invalid Entry client material")
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(material.Data["ca.crt"]) {
		return fmt.Errorf("Desktop readiness has no valid Entry CA")
	}
	sni := entryMeshService.ServiceName + "." + defaultGatewayNS + ".svc.cluster.local"
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{RootCAs: roots, Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12},
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			_, port, err := net.SplitHostPort(address)
			if err != nil || (port != "8443" && port != "9444") {
				return nil, fmt.Errorf("invalid Entry probe destination")
			}
			return dial(ctx, network, net.JoinHostPort(ip.String(), port))
		},
	}
	defer transport.CloseIdleConnections()
	httpClient := &http.Client{Transport: transport, Timeout: 3 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	logins := gateway.SystemPlaneHostPatterns(gateway.SystemPlaneService{Name: "auth"}, z, srr.Annotations[gateway.AnnotationExtAuthUser], false)
	for i, host := range expected {
		if err := verifyDesktopEntryResponses(ctx, httpClient, "https://"+net.JoinHostPort(sni, strconv.Itoa(8443)), "https://"+net.JoinHostPort(sni, strconv.Itoa(9444)), host, logins[i]); err != nil {
			return err
		}
	}
	return nil
}

func verifyDesktopEntryResponses(ctx context.Context, c *http.Client, entry, auth, desktop, login string) error {
	for _, probe := range []struct {
		base, host, path string
		authorization    bool
	}{
		{entry, login, "/", false}, {entry, desktop, "/", false},
		{entry, desktop, "/favicon.ico", false}, {auth, desktop, AutheliaVerifyPath, true},
	} {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, probe.base+probe.path, nil)
		if err != nil {
			return err
		}
		req.Host = probe.host
		if probe.host == desktop {
			path := probe.path
			if probe.authorization {
				path = "/"
			}
			req.Header.Set("Accept", "text/html")
			req.Header.Set("X-Original-URL", "https://"+desktop+path)
			req.Header.Set("X-Original-Method", http.MethodGet)
			req.Header.Set("X-Forwarded-Host", desktop)
			req.Header.Set("X-Forwarded-Proto", "https")
		}
		resp, err := c.Do(req)
		if err != nil {
			return fmt.Errorf("Desktop Entry probe failed: %w", err)
		}
		_ = resp.Body.Close()
		if probe.host == login && resp.StatusCode == http.StatusOK {
			continue
		}
		if probe.host == desktop {
			if probe.authorization && (resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden) {
				continue
			}
			switch resp.StatusCode {
			case 302, 303, 307, 308:
				u, err := url.Parse(resp.Header.Get("Location"))
				if err == nil && u.User == nil && strings.EqualFold(u.Hostname(), login) &&
					((u.Scheme == "https" && (u.Port() == "" || u.Port() == "443")) ||
						(strings.HasSuffix(strings.ToLower(login), ".olares.local") && u.Scheme == "http" && (u.Port() == "" || u.Port() == "80"))) {
					continue
				}
			}
		}
		return fmt.Errorf("Desktop Entry anonymous probe for %s%s returned unexpected HTTP %d", probe.host, probe.path, resp.StatusCode)
	}
	return nil
}
