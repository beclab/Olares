package routecontrol

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestDesktopEntryProbeChecksNavigationInsteadOfAnyDenial(t *testing.T) {
	for _, tc := range []struct {
		name     string
		status   int
		location string
		path     string
		ok       bool
	}{
		{"redirect", 302, "https://auth.alice.example.org/?rd=desktop", "/", true},
		{"tenant-context-missing", 400, "", "/", false},
		{"stale-favicon-path", 400, "", "/favicon.ico", false},
		{"anonymous-bypass", 200, "", "/", false},
		{"browser-denied", 403, "", "/", false},
		{"desktop-loop", 302, "https://desktop.alice.example.org/api/verify", "/", false},
		{"plaintext", 302, "http://auth.alice.example.org/", "/", false},
		{"foreign-tenant", 302, "https://auth.bob.example.org/", "/", false},
		{"userinfo", 302, "https://bob@auth.alice.example.org/", "/", false},
		{"wrong-port", 302, "https://auth.alice.example.org:444/", "/", false},
		{"login-unavailable", 503, "", "login", false},
		{"auth-unavailable", 500, "", AutheliaVerifyPath, false},
		{"auth-denies-anonymous", 401, "", AutheliaVerifyPath, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("X-BFL-USER") != "" || r.Header.Get("Cookie") != "" || r.Header.Get("Authorization") != "" {
					t.Error("readiness supplied a tenant or session identity")
				}
				if r.Host == "auth.alice.example.org" {
					if tc.path == "login" {
						w.WriteHeader(tc.status)
					}
					return
				}
				original := r.URL.Path
				if original == AutheliaVerifyPath {
					original = "/"
				}
				if r.Header.Get("X-Original-URL") != "https://desktop.alice.example.org"+original || r.Header.Get("Accept") != "text/html" {
					t.Error("missing original browser context")
				}
				if r.URL.Path == tc.path {
					w.Header().Set("Location", tc.location)
					w.WriteHeader(tc.status)
					return
				}
				http.Redirect(w, r, "https://auth.alice.example.org/", 302)
			}))
			defer server.Close()
			c := server.Client()
			c.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
			if err := verifyDesktopEntryResponses(t.Context(), c, server.URL, server.URL, "desktop.alice.example.org", "auth.alice.example.org"); (err == nil) != tc.ok {
				t.Fatalf("unexpected result: %v", err)
			}
		})
	}
}

func entryProbeMaterial(t *testing.T) (tls.Certificate, *corev1.Secret) {
	t.Helper()
	pub, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{SerialNumber: big.NewInt(1), NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		DNSNames: []string{"app-gateway-l4-mtls.os-gateway.svc.cluster.local"}, IsCA: true, BasicConstraintsValid: true,
		KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}}
	der, err := x509.CreateCertificate(rand.Reader, template, template, pub, key)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: encoded})
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatal(err)
	}
	return cert, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: defaultGatewayNS, Name: "olares-l4-agw-client"}, Data: map[string][]byte{"client.crt": certPEM, "client.key": keyPEM, "ca.crt": certPEM}}
}

func TestDesktopEntryProbeUsesFreshMaterialVerifiedTLSAndFixedDestination(t *testing.T) {
	cert, secret := entryProbeMaterial(t)
	_, foreignCA := entryProbeMaterial(t)
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(secret.Data["ca.crt"])
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.TLS.PeerCertificates) != 1 {
			t.Error("probe did not authenticate its Entry client")
		}
		if r.Host != "auth.alice.example.org" {
			http.Redirect(w, r, "https://auth.alice.example.org/", 302)
		}
	}))
	server.TLS = &tls.Config{Certificates: []tls.Certificate{cert}, ClientAuth: tls.RequireAndVerifyClientCert, ClientCAs: pool, MinVersion: tls.VersionTLS12}
	server.StartTLS()
	defer server.Close()
	srr := desktopEntrance()
	cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Namespace: srr.Namespace, Name: zoneSSLConfigMapName}, Data: map[string]string{"zone": "alice.example.org"}}
	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: defaultGatewayNS, Name: entryMeshService.ServiceName}, Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeClusterIP, ClusterIP: "10.96.0.10"}}
	c := fake.NewClientBuilder().WithScheme(testScheme(t)).WithObjects(cm, svc, secret).Build()
	requests := 0
	dial := func(ctx context.Context, network, address string) (net.Conn, error) {
		if address != "10.96.0.10:8443" && address != "10.96.0.10:9444" {
			t.Errorf("probe escaped fixed Entry: %s", address)
		}
		requests++
		return (&net.Dialer{}).DialContext(ctx, network, server.Listener.Addr().String())
	}
	if err := probeDesktopEntryWithDialer(t.Context(), c, srr, dial); err != nil || requests == 0 {
		t.Fatal("verified Entry probe failed", err)
	}

	// Exercise the platform-generated pair, including the local alias, without
	// relying on list order. No alias is removed to satisfy the readiness check.
	t.Run("generated-domain-pair", func(t *testing.T) {
		if len(srr.Spec.HostPatterns) != 2 || srr.Spec.HostPatterns[0] != "desktop.alice.example.org" || srr.Spec.HostPatterns[1] != "desktop.alice.olares.local" {
			t.Fatal("fixture must use the platform's primary and local domains")
		}
		candidate := srr.DeepCopy()
		candidate.Spec.HostPatterns[0], candidate.Spec.HostPatterns[1] = candidate.Spec.HostPatterns[1], candidate.Spec.HostPatterns[0]
		before := requests
		if err := probeDesktopEntryWithDialer(t.Context(), c, candidate, dial); err != nil || requests == before {
			t.Fatal("reordered platform domains did not reach the authenticated Entry probe", err)
		}
	})
	for _, tc := range []struct {
		name  string
		hosts []string
	}{
		{"foreign-primary", []string{"desktop.bob.example.org", "desktop.alice.olares.local"}},
		{"foreign-local-alias", []string{"desktop.alice.example.org", "desktop.bob.olares.local"}},
		{"foreign-pair", []string{"desktop.bob.example.org", "desktop.bob.olares.local"}},
		{"same-user-different-zone", []string{"desktop.alice.other.org", "desktop.alice.olares.local"}},
		{"extra-foreign-domain", []string{"desktop.alice.example.org", "desktop.alice.olares.local", "desktop.bob.example.org"}},
		{"wildcard", []string{"desktop.alice.example.org", "*.alice.olares.local"}},
		{"missing-local-alias", []string{"desktop.alice.example.org"}},
		{"missing-primary", []string{"desktop.alice.olares.local"}},
		{"empty-domains", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			candidate := srr.DeepCopy()
			candidate.Spec.HostPatterns = tc.hosts
			before := requests
			err := probeDesktopEntryWithDialer(t.Context(), c, candidate, dial)
			if err == nil || !strings.Contains(err.Error(), "managed zone authorities") {
				t.Fatal("unbound domain set was not rejected", err)
			}
			if requests != before {
				t.Fatal("unbound domains reached the network")
			}
		})
	}
	// Invalid route snapshots must not poison a later valid probe.
	if err := probeDesktopEntryWithDialer(t.Context(), c, srr, dial); err != nil {
		t.Fatal("valid generated domains did not recover", err)
	}
	for _, mode := range []string{"wrong-host", "bad-key", "bad-ca", "untrusted-server", "missing-secret", "wrong-service", "wrong-zone", "cancelled"} {
		t.Run(mode, func(t *testing.T) {
			candidate, sec, service, zone := srr.DeepCopy(), secret.DeepCopy(), svc.DeepCopy(), cm.DeepCopy()
			ctx := t.Context()
			switch mode {
			case "wrong-host":
				candidate.Spec.HostPatterns = []string{"desktop.bob.example.org"}
			case "bad-key":
				sec.Data["client.key"] = nil
			case "bad-ca":
				sec.Data["ca.crt"] = nil
			case "untrusted-server":
				sec.Data["ca.crt"] = foreignCA.Data["ca.crt"]
			case "wrong-service":
				service.Spec.ClusterIP = "127.0.0.1"
			case "wrong-zone":
				zone.Data["zone"] = "bad zone"
			case "cancelled":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			objects := []client.Object{service, zone}
			if mode != "missing-secret" {
				objects = append(objects, sec)
			}
			reader := fake.NewClientBuilder().WithScheme(testScheme(t)).WithObjects(objects...).Build()
			if err := probeDesktopEntryWithDialer(ctx, reader, candidate, dial); err == nil {
				t.Fatal("invalid or cancelled Entry probe passed")
			}
		})
	}
	// Re-reading the Secret must immediately reject revoked client material.
	secret.Data["client.key"] = []byte("revoked")
	if err := c.Update(t.Context(), secret); err != nil {
		t.Fatal(err)
	}
	if err := probeDesktopEntryWithDialer(t.Context(), c, srr, dial); err == nil {
		t.Fatal("stale client material was reused")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := verifyDesktopEntryResponses(ctx, server.Client(), server.URL, server.URL, "desktop.alice.example.org", "auth.alice.example.org"); err == nil || !strings.Contains(err.Error(), "probe failed") {
		t.Fatal(fmt.Sprint(err))
	}
}
