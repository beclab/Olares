package linkerdpki

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGenerateWebhookServingCert_SANandValidity(t *testing.T) {
	now := time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC)
	crtPEM, keyPEM, err := generateWebhookServingCert("linkerd-proxy-injector.os-mesh.svc", now)
	if err != nil {
		t.Fatal(err)
	}
	if len(keyPEM) == 0 {
		t.Fatal("empty key")
	}
	block, _ := pem.Decode(crtPEM)
	if block == nil {
		t.Fatal("bad cert pem")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if cert.Subject.CommonName != "linkerd-proxy-injector.os-mesh.svc" {
		t.Fatalf("CN=%q", cert.Subject.CommonName)
	}
	wantDNS := map[string]bool{
		"linkerd-proxy-injector.os-mesh.svc":               true,
		"linkerd-proxy-injector.os-mesh.svc.cluster.local": true,
	}
	for _, d := range cert.DNSNames {
		delete(wantDNS, d)
	}
	if len(wantDNS) != 0 {
		t.Fatalf("missing DNS SANs: %v", wantDNS)
	}
	days := cert.NotAfter.Sub(cert.NotBefore).Hours() / 24
	if days < float64(WebhookCertValidityDays-2) || days > float64(WebhookCertValidityDays+2) {
		t.Fatalf("validity days=%v, want ~%d", days, WebhookCertValidityDays)
	}
	if !webhookCertHasDNS(crtPEM, "linkerd-proxy-injector.os-mesh.svc") {
		t.Fatal("webhookCertHasDNS false")
	}
	if webhookCertHasDNS(crtPEM, "linkerd-proxy-injector.linkerd.svc") {
		t.Fatal("should not match linkerd.svc")
	}
}

func TestEnsureWebhookCerts_CreatesSecretAndCABundle(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = admissionregistrationv1.AddToScheme(scheme)

	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: DefaultLinkerdNamespace}}
	mwh := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: proxyInjectorWebhookName},
		Webhooks: []admissionregistrationv1.MutatingWebhook{{
			Name: "linkerd-proxy-injector.linkerd.io",
			ClientConfig: admissionregistrationv1.WebhookClientConfig{
				Service: &admissionregistrationv1.ServiceReference{
					Namespace: DefaultLinkerdNamespace,
					Name:      "linkerd-proxy-injector",
				},
			},
		}},
	}
	vwhSP := &admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: spValidatorWebhookName},
		Webhooks: []admissionregistrationv1.ValidatingWebhook{{
			Name: "linkerd-sp-validator.linkerd.io",
			ClientConfig: admissionregistrationv1.WebhookClientConfig{
				Service: &admissionregistrationv1.ServiceReference{
					Namespace: DefaultLinkerdNamespace,
					Name:      "linkerd-sp-validator",
				},
			},
		}},
	}
	vwhPol := &admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: policyValidatorWebhookName},
		Webhooks: []admissionregistrationv1.ValidatingWebhook{{
			Name: "linkerd-policy-validator.linkerd.io",
			ClientConfig: admissionregistrationv1.WebhookClientConfig{
				Service: &admissionregistrationv1.ServiceReference{
					Namespace: DefaultLinkerdNamespace,
					Name:      "linkerd-policy-validator",
				},
			},
		}},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ns, mwh, vwhSP, vwhPol).Build()
	ctx := context.Background()
	if err := EnsureWebhookCerts(ctx, c, DefaultLinkerdNamespace, false); err != nil {
		t.Fatalf("EnsureWebhookCerts: %v", err)
	}
	var sec corev1.Secret
	if err := c.Get(ctx, types.NamespacedName{Namespace: DefaultLinkerdNamespace, Name: proxyInjectorSecretName}, &sec); err != nil {
		t.Fatal(err)
	}
	if !webhookCertHasDNS(sec.Data[corev1.TLSCertKey], "linkerd-proxy-injector.os-mesh.svc") {
		t.Fatal("secret SAN mismatch")
	}
	var got admissionregistrationv1.MutatingWebhookConfiguration
	if err := c.Get(ctx, types.NamespacedName{Name: proxyInjectorWebhookName}, &got); err != nil {
		t.Fatal(err)
	}
	if string(got.Webhooks[0].ClientConfig.CABundle) != string(sec.Data[corev1.TLSCertKey]) {
		t.Fatal("caBundle must match secret tls.crt")
	}
	fp := string(sec.Data[corev1.TLSCertKey])
	if err := EnsureWebhookCerts(ctx, c, DefaultLinkerdNamespace, false); err != nil {
		t.Fatalf("idempotent: %v", err)
	}
	_ = c.Get(ctx, types.NamespacedName{Namespace: DefaultLinkerdNamespace, Name: proxyInjectorSecretName}, &sec)
	if string(sec.Data[corev1.TLSCertKey]) != fp {
		t.Fatal("idempotent run must not rotate cert without overwrite")
	}
}
