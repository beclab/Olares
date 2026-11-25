package entryreadiness

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"math/big"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/beclab/l4-bfl-proxy/internal/message"
)

type probeRoundTrip func(*http.Request) (*http.Response, error)

func (f probeRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestGatewayProbeUsesTLSAndCorrectAuthority(t *testing.T) {
	g := Gateway{Host: "gateway.example.org", EntryPort: 8443, AuthPort: 9444}
	const loginLocation = "https://auth.alice.example.org/?rd=desktop"
	for _, tc := range []struct {
		name                        string
		entry, desktop, asset, auth int
		location                    string
		netFailure, wantPass        bool
	}{
		{"ready", 200, 302, 302, 401, "", false, true},
		{"browser login redirect", 200, 302, 302, 302, loginLocation, false, true},
		{"login unavailable", 503, 302, 302, 401, "", false, false},
		{"wrong login endpoint", 404, 302, 302, 401, "", false, false},
		{"desktop path normalization redirect", 200, 301, 302, 302, loginLocation, false, false},
		{"desktop asset appended to verify path", 200, 302, 400, 302, loginLocation, false, false},
		{"desktop accepted anonymously", 200, 200, 302, 302, loginLocation, false, false},
		{"desktop denial without browser redirect", 200, 403, 302, 302, loginLocation, false, false},
		{"auth route missing", 200, 302, 302, 404, "", false, false},
		{"auth path normalization redirect", 200, 302, 302, 301, loginLocation, false, false},
		{"redirect without location", 200, 302, 302, 302, "", false, false},
		{"plaintext redirect", 200, 302, 302, 302, "http://auth.alice.example.org/", false, false},
		{"desktop redirect loop", 200, 302, 302, 302, "https://desktop.alice.example.org/api/verify", false, false},
		{"redirect to another tenant", 200, 302, 302, 302, "https://auth.bob.example.org/", false, false},
		{"redirect with authority credentials", 200, 302, 302, 302, "https://user@auth.alice.example.org/", false, false},
		{"redirect to another port", 200, 302, 302, 302, "https://auth.alice.example.org:9444/", false, false},
		{"anonymous auth accepted", 200, 302, 302, 200, "", false, false},
		{"TLS failure", 200, 302, 302, 401, "", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var seen []string
			client := &http.Client{Transport: probeRoundTrip(func(r *http.Request) (*http.Response, error) {
				if r.URL.Scheme != "https" || r.URL.Hostname() != g.Host || r.Method != http.MethodGet {
					t.Fatal("probe bypassed the secure gateway")
				}
				if tc.netFailure {
					return nil, errors.New("remote TLS alert")
				}
				seen = append(seen, r.URL.Port()+" "+r.Host+r.URL.Path)
				var status int
				location := loginLocation
				originalPath := r.URL.Path
				switch r.URL.Port() {
				case "8443":
					switch {
					case r.Host == "auth.alice.example.org" && r.URL.Path == "/":
						status = tc.entry
					case r.Host == "desktop.alice.example.org" && r.URL.Path == "/":
						status = tc.desktop
					case r.Host == "desktop.alice.example.org" && r.URL.Path == "/favicon.ico":
						status = tc.asset
					default:
						t.Fatalf("unexpected Entry route: %s %s", r.Host, r.URL.Path)
					}
				case "9444":
					if r.Host != "desktop.alice.example.org" || r.URL.Path != "/api/verify" {
						t.Fatalf("unexpected authorization route: %s %s", r.Host, r.URL.Path)
					}
					status = tc.auth
					location = tc.location
					originalPath = "/"
				default:
					t.Fatalf("unexpected port %s", r.URL.Port())
				}
				if r.Host == "desktop.alice.example.org" &&
					(r.Header.Get("Accept") != "text/html" ||
						r.Header.Get("X-Original-URL") != "https://"+r.Host+originalPath ||
						r.Header.Get("X-Original-Method") != "GET" ||
						r.Header.Get("X-Forwarded-Host") != r.Host || r.Header.Get("X-Forwarded-Proto") != "https") {
					t.Fatal("authorization probe did not preserve the Desktop request context")
				}
				return &http.Response{
					StatusCode: status,
					Header:     http.Header{"Location": []string{location}},
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil
			}), CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			}}
			if err := g.verify(t.Context(), client, "alice.example.org"); (err == nil) != tc.wantPass {
				t.Fatalf("probe = %v, want ready=%v", err, tc.wantPass)
			}
			want := []string{"8443 auth.alice.example.org/", "8443 desktop.alice.example.org/",
				"8443 desktop.alice.example.org/favicon.ico", "9444 desktop.alice.example.org/api/verify"}
			if tc.wantPass && !slices.Equal(seen, want) {
				t.Fatalf("checked %v, want %v", seen, want)
			}
		})
	}
}

func TestEntryProbeRequiresValidCredentialAndTrust(t *testing.T) {
	g := Gateway{Host: "gateway.example.org", EntryPort: 8443, AuthPort: 9444}
	if _, _, err := g.client(&message.EntryMTLSMaterial{}); err == nil {
		t.Fatal("missing credentials admitted")
	}
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	spec := &x509.Certificate{SerialNumber: big.NewInt(1), NotBefore: now.Add(-time.Minute), NotAfter: now.Add(time.Minute)}
	cert, err := x509.CreateCertificate(rand.Reader, spec, spec, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	material := &message.EntryMTLSMaterial{
		ClientCert: string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert})),
		ClientKey:  string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})),
	}
	if _, _, err := g.client(material); err == nil {
		t.Fatal("missing CA admitted")
	}
	material.CA = material.ClientCert
	probe, cleanup, err := g.client(material)
	if err != nil || probe == nil || cleanup == nil {
		t.Fatalf("valid client configuration rejected: %v", err)
	}
	cleanup()
}
