package gateway

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
)

func mustTestCertPEM(t *testing.T, dnsNames ...string) (certPEM, keyPEM string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: dnsNames[0]},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		DNSNames:     dnsNames,
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	certPEM = string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	keyPEM = string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}))
	return certPEM, keyPEM
}

func eligibleCustomDomainBlob(t *testing.T, entrance, fqdn, cert, key string) string {
	t.Helper()
	entry := map[string]interface{}{
		"third_party_domain":  fqdn,
		"cert":                cert,
		"key":                 key,
		"cname_target_status": "set",
		"cname_status":        "active",
	}
	b, err := json.Marshal(map[string]interface{}{entrance: entry})
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestEligibleCustomHost(t *testing.T) {
	cert, key := mustTestCertPEM(t, "chat.example.com")
	base := customDomainEntranceFields{
		ThirdPartyDomain:  "chat.example.com",
		Cert:              cert,
		Key:               key,
		CnameTargetStatus: "set",
		CnameStatus:       "active",
	}
	if ok, reason := EligibleCustomHost(base, "olares.com", nil); !ok || reason != "" {
		t.Fatalf("want pass, got ok=%v reason=%q", ok, reason)
	}

	cases := []struct {
		name   string
		mutate func(*customDomainEntranceFields)
		reason string
	}{
		{"platform suffix", func(c *customDomainEntranceFields) { c.ThirdPartyDomain = "x.olares.com" }, DenyReservedSuffix},
		{"no cert", func(c *customDomainEntranceFields) { c.Cert = "" }, DenyNoCert},
		{"cname", func(c *customDomainEntranceFields) { c.CnameStatus = "pending" }, DenyCNAMENotActive},
		{"bad dns", func(c *customDomainEntranceFields) { c.ThirdPartyDomain = "http://x.com" }, DenyInvalidDNS},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := base
			tc.mutate(&cfg)
			ok, reason := EligibleCustomHost(cfg, "olares.com", nil)
			if ok || reason != tc.reason {
				t.Fatalf("ok=%v reason=%q want %q", ok, reason, tc.reason)
			}
		})
	}

	wrongCert, wrongKey := mustTestCertPEM(t, "other.example.com")
	badSAN := base
	badSAN.Cert, badSAN.Key = wrongCert, wrongKey
	if ok, reason := EligibleCustomHost(badSAN, "olares.com", nil); ok || reason != DenyCertNameMismatch {
		t.Fatalf("SAN mismatch: ok=%v reason=%q", ok, reason)
	}
}

func TestCollectEligibleExactHostsSharedOverlay(t *testing.T) {
	cert, key := mustTestCertPEM(t, "chat.example.com")
	blob := eligibleCustomDomainBlob(t, "web", "chat.example.com", cert, key)
	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "demo"},
		Spec: appv1alpha1.ApplicationSpec{
			Name:  "demo",
			Owner: "installer",
			UserSettings: map[string]map[string]string{
				"alice": {"customDomain": blob},
			},
		},
	}
	got := CollectEligibleExactHosts(app, "web", "olares.com", nil)
	if len(got) != 1 || got[0].Host != "chat.example.com" || got[0].Owner != "alice" {
		t.Fatalf("got %#v", got)
	}
}

func TestBuildSpecForEntranceAppendsEligibleExactHost(t *testing.T) {
	cert, key := mustTestCertPEM(t, "chat.example.com")
	blob := eligibleCustomDomainBlob(t, "web", "chat.example.com", cert, key)
	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo",
			Labels: map[string]string{
				constants.AppApiVersionLabel: "v3",
				constants.AppSharedLabel:     constants.AppSharedTrue,
			},
		},
		Spec: appv1alpha1.ApplicationSpec{
			Appid:     "demo1234",
			Name:      "demo",
			Namespace: "demo-shared",
			Owner:     "alice",
			UserSettings: map[string]map[string]string{
				"alice": {"customDomain": blob},
			},
			SharedEntrances: []appv1alpha1.Entrance{
				{Name: "web", Host: "demo-svc", Port: 8080},
			},
		},
	}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-svc", Namespace: "demo-shared"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 8080, Protocol: corev1.ProtocolTCP}}},
	}
	spec, err := BuildSpecForEntrance(app, app.Spec.SharedEntrances[0], 0, 1, svc, "olares.com",
		srrv1alpha1.EntranceClassShared, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(spec.HostPatterns) < 2 {
		t.Fatalf("want dual-host, got %v", spec.HostPatterns)
	}
	foundExact := false
	for _, h := range spec.HostPatterns {
		if h == "chat.example.com" {
			foundExact = true
		}
	}
	if !foundExact {
		t.Fatalf("exact host missing: %v", spec.HostPatterns)
	}
}

func TestCollectUserThirdLevelExactHosts(t *testing.T) {
	blobAlice := `{"web":{"third_level_domain":"api"}}`
	blobBob := `{"web":{"third_level_domain":"chat"}}`
	app := &appv1alpha1.Application{
		Spec: appv1alpha1.ApplicationSpec{
			Owner: "installer",
			UserSettings: map[string]map[string]string{
				"alice": {"customDomain": blobAlice},
				"bob":   {"customDomain": blobBob},
			},
		},
	}
	got := CollectUserThirdLevelExactHosts(app, "web", "olares.com")
	want := map[string]string{
		"api.alice.olares.com":  "alice",
		"chat.bob.olares.com":   "bob",
	}
	if len(got) != 2 {
		t.Fatalf("got %#v", got)
	}
	for _, o := range got {
		if want[o.Host] != o.Owner {
			t.Fatalf("unexpected %#v want %v", got, want)
		}
	}
}

func TestBuildSpecForEntranceAppendsUserThirdLevelExact(t *testing.T) {
	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo",
			Labels: map[string]string{
				constants.AppApiVersionLabel: "v3",
				constants.AppSharedLabel:     constants.AppSharedTrue,
			},
		},
		Spec: appv1alpha1.ApplicationSpec{
			Appid:     "demo1234",
			Name:      "demo",
			Namespace: "demo-shared",
			Owner:     "alice",
			UserSettings: map[string]map[string]string{
				"alice": {"customDomain": `{"web":{"third_level_domain":"api"}}`},
				"bob":   {"customDomain": `{"web":{"third_level_domain":"chat"}}`},
			},
			SharedEntrances: []appv1alpha1.Entrance{
				{Name: "web", Host: "demo-svc", Port: 8080},
			},
		},
	}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-svc", Namespace: "demo-shared"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 8080, Protocol: corev1.ProtocolTCP}}},
	}
	spec, err := BuildSpecForEntrance(app, app.Spec.SharedEntrances[0], 0, 1, svc, "olares.com",
		srrv1alpha1.EntranceClassShared, nil)
	if err != nil {
		t.Fatal(err)
	}
	foundAlice, foundBob, foundLogical := false, false, false
	for _, h := range spec.HostPatterns {
		switch h {
		case "api.alice.olares.com":
			foundAlice = true
		case "chat.bob.olares.com":
			foundBob = true
		case "api.*.olares.com", "chat.*.olares.com":
			foundLogical = true
		}
	}
	if !foundAlice || !foundBob {
		t.Fatalf("per-owner exact third_level missing: %v", spec.HostPatterns)
	}
	if foundLogical {
		t.Fatalf("must not broadcast logical third_level prefixes: %v", spec.HostPatterns)
	}
}

func TestEligibleCustomHostUsesMaterializer(t *testing.T) {
	cert, key := mustTestCertPEM(t, "chat.example.com")
	cfg := customDomainEntranceFields{
		ThirdPartyDomain:  "chat.example.com",
		CnameTargetStatus: "set",
		CnameStatus:       "active",
	}
	if ok, reason := EligibleCustomHost(cfg, "olares.com", nil); ok || reason != DenyNoCert {
		t.Fatalf("without materializer: ok=%v reason=%q", ok, reason)
	}
	mat := CertMaterializer(func(host string) (string, string, bool) {
		if host == "chat.example.com" {
			return cert, key, true
		}
		return "", "", false
	})
	if ok, reason := EligibleCustomHost(cfg, "olares.com", mat); !ok || reason != "" {
		t.Fatalf("with materializer: ok=%v reason=%q", ok, reason)
	}
}

func TestConfigMapCertMaterializerRetriesAfterListFailure(t *testing.T) {
	cert, key := mustTestCertPEM(t, "shop.example.com")
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "shop-cert",
			Namespace: "user-space-alice",
			Labels:    map[string]string{customDomainCertLabel: customDomainCertLabelValue},
		},
		Data: map[string]string{"zone": "shop.example.com", "cert": cert, "key": key},
	}
	okClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cm).Build()
	fc := &flakyListClient{Client: okClient, failLeft: 1}
	mat := NewConfigMapCertMaterializer(context.Background(), fc)
	if _, _, ok := mat("shop.example.com"); ok {
		t.Fatal("first call should miss while list fails")
	}
	gotCert, gotKey, ok := mat("shop.example.com")
	if !ok || gotCert == "" || gotKey == "" {
		t.Fatalf("second call should load after list succeeds: ok=%v", ok)
	}
}

type flakyListClient struct {
	client.Client
	failLeft int
}

func (c *flakyListClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	if c.failLeft > 0 {
		c.failLeft--
		return fmt.Errorf("simulated list failure")
	}
	return c.Client.List(ctx, list, opts...)
}

func TestReservedExactHostAllowsAuthOnPublicDomain(t *testing.T) {
	if reason := reservedExactHostReason("auth.myshop.com", "olares.com"); reason != "" {
		t.Fatalf("public auth.myshop.com should not be reserved, got %q", reason)
	}
	if reason := reservedExactHostReason("auth.olares.com", "olares.com"); reason != DenyReservedSuffix {
		t.Fatalf("auth under platform must be reserved, got %q", reason)
	}
}
