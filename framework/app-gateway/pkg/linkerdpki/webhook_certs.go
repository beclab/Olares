package linkerdpki

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"time"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// WebhookCertValidityDays is 100 years, aligned with EG control-plane certgen.
	WebhookCertValidityDays = 36500

	proxyInjectorSecretName   = "linkerd-proxy-injector-k8s-tls"
	spValidatorSecretName     = "linkerd-sp-validator-k8s-tls"
	policyValidatorSecretName = "linkerd-policy-validator-k8s-tls"

	proxyInjectorWebhookName   = "linkerd-proxy-injector-webhook-config"
	spValidatorWebhookName     = "linkerd-sp-validator-webhook-config"
	policyValidatorWebhookName = "linkerd-policy-validator-webhook-config"
)

// webhookTLSTarget pairs a kubernetes.io/tls Secret with the admission webhook
// configuration that must embed the same serving cert as caBundle.
type webhookTLSTarget struct {
	SecretName     string
	ServiceDNS     string // linkerd-proxy-injector.<ns>.svc
	MutatingName   string // empty if validating
	ValidatingName string
}

func webhookTargets(ns string) []webhookTLSTarget {
	return []webhookTLSTarget{
		{
			SecretName:   proxyInjectorSecretName,
			ServiceDNS:   fmt.Sprintf("linkerd-proxy-injector.%s.svc", ns),
			MutatingName: proxyInjectorWebhookName,
		},
		{
			SecretName:     spValidatorSecretName,
			ServiceDNS:     fmt.Sprintf("linkerd-sp-validator.%s.svc", ns),
			ValidatingName: spValidatorWebhookName,
		},
		{
			SecretName:     policyValidatorSecretName,
			ServiceDNS:     fmt.Sprintf("linkerd-policy-validator.%s.svc", ns),
			ValidatingName: policyValidatorWebhookName,
		},
	}
}

// EnsureWebhookCerts creates per-cluster Linkerd admission webhook TLS Secrets
// (100y, SAN=*.<ns>.svc) and syncs caBundle on the matching webhook configs.
// When overwrite is false and a Secret already has a matching SAN, the Secret is
// left unchanged (search3 lookup semantics) but caBundle is still reconciled.
func EnsureWebhookCerts(ctx context.Context, c client.Client, ns string, overwrite bool) error {
	if c == nil || ns == "" {
		return fmt.Errorf("client and namespace are required")
	}
	for _, t := range webhookTargets(ns) {
		crtPEM, _, err := ensureWebhookSecret(ctx, c, ns, t, overwrite)
		if err != nil {
			return err
		}
		if err := patchWebhookCABundle(ctx, c, t, crtPEM); err != nil {
			return err
		}
		slog.Info("linkerd webhook tls ready",
			"secret", t.SecretName, "dns", t.ServiceDNS, "overwrite", overwrite)
	}
	return nil
}

func ensureWebhookSecret(ctx context.Context, c client.Client, ns string, t webhookTLSTarget, overwrite bool) (crtPEM, keyPEM []byte, err error) {
	var sec corev1.Secret
	err = c.Get(ctx, types.NamespacedName{Namespace: ns, Name: t.SecretName}, &sec)
	switch {
	case apierrors.IsNotFound(err):
		crtPEM, keyPEM, err = generateWebhookServingCert(t.ServiceDNS, time.Now().UTC())
		if err != nil {
			return nil, nil, err
		}
		sec = corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      t.SecretName,
				Namespace: ns,
				Labels: map[string]string{
					"app.kubernetes.io/name":             "app-gateway",
					"app.kubernetes.io/component":        "linkerd-webhook-certgen",
					"linkerd.io/control-plane-component": componentLabel(t),
					"linkerd.io/control-plane-ns":        ns,
				},
			},
			Type: corev1.SecretTypeTLS,
			Data: map[string][]byte{
				corev1.TLSCertKey:       crtPEM,
				corev1.TLSPrivateKeyKey: keyPEM,
			},
		}
		if err := c.Create(ctx, &sec); err != nil {
			return nil, nil, fmt.Errorf("create secret %s/%s: %w", ns, t.SecretName, err)
		}
		return crtPEM, keyPEM, nil
	case err != nil:
		return nil, nil, fmt.Errorf("get secret %s/%s: %w", ns, t.SecretName, err)
	}

	crtPEM = sec.Data[corev1.TLSCertKey]
	keyPEM = sec.Data[corev1.TLSPrivateKeyKey]
	if !overwrite && webhookCertHasDNS(crtPEM, t.ServiceDNS) {
		return crtPEM, keyPEM, nil
	}
	crtPEM, keyPEM, err = generateWebhookServingCert(t.ServiceDNS, time.Now().UTC())
	if err != nil {
		return nil, nil, err
	}
	if sec.Data == nil {
		sec.Data = map[string][]byte{}
	}
	sec.Type = corev1.SecretTypeTLS
	sec.Data[corev1.TLSCertKey] = crtPEM
	sec.Data[corev1.TLSPrivateKeyKey] = keyPEM
	if err := c.Update(ctx, &sec); err != nil {
		return nil, nil, fmt.Errorf("update secret %s/%s: %w", ns, t.SecretName, err)
	}
	return crtPEM, keyPEM, nil
}

func componentLabel(t webhookTLSTarget) string {
	switch {
	case strings.Contains(t.SecretName, "proxy-injector"):
		return "proxy-injector"
	default:
		return "destination"
	}
}

func patchWebhookCABundle(ctx context.Context, c client.Client, t webhookTLSTarget, crtPEM []byte) error {
	deadline := time.Now().Add(2 * time.Minute)
	var last error
	for time.Now().Before(deadline) {
		if t.MutatingName != "" {
			last = patchMutatingCABundle(ctx, c, t.MutatingName, crtPEM)
		} else {
			last = patchValidatingCABundle(ctx, c, t.ValidatingName, crtPEM)
		}
		if last == nil {
			return nil
		}
		if !apierrors.IsNotFound(last) {
			return last
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("webhook config not ready: %w", last)
}

func patchMutatingCABundle(ctx context.Context, c client.Client, name string, crtPEM []byte) error {
	var wh admissionregistrationv1.MutatingWebhookConfiguration
	if err := c.Get(ctx, types.NamespacedName{Name: name}, &wh); err != nil {
		return err
	}
	changed := false
	for i := range wh.Webhooks {
		if string(wh.Webhooks[i].ClientConfig.CABundle) == string(crtPEM) {
			continue
		}
		wh.Webhooks[i].ClientConfig.CABundle = crtPEM
		changed = true
	}
	if !changed {
		return nil
	}
	return c.Update(ctx, &wh)
}

func patchValidatingCABundle(ctx context.Context, c client.Client, name string, crtPEM []byte) error {
	var wh admissionregistrationv1.ValidatingWebhookConfiguration
	if err := c.Get(ctx, types.NamespacedName{Name: name}, &wh); err != nil {
		return err
	}
	changed := false
	for i := range wh.Webhooks {
		if string(wh.Webhooks[i].ClientConfig.CABundle) == string(crtPEM) {
			continue
		}
		wh.Webhooks[i].ClientConfig.CABundle = crtPEM
		changed = true
	}
	if !changed {
		return nil
	}
	return c.Update(ctx, &wh)
}

// generateWebhookServingCert returns a self-signed RSA serving cert (PEM) whose
// CN/SAN include shortSvcDNS and shortSvcDNS+".cluster.local".
func generateWebhookServingCert(shortSvcDNS string, now time.Time) (crtPEM, keyPEM []byte, err error) {
	if shortSvcDNS == "" {
		return nil, nil, fmt.Errorf("service DNS is required")
	}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}
	notBefore := now.Add(-time.Hour)
	notAfter := notBefore.Add(WebhookCertValidityDays * 24 * time.Hour)
	dns := []string{shortSvcDNS, shortSvcDNS + ".cluster.local"}
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: shortSvcDNS},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     dns,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return nil, nil, err
	}
	crtPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	return crtPEM, keyPEM, nil
}

func webhookCertHasDNS(crtPEM []byte, wantDNS string) bool {
	cert, err := parseCertificate(crtPEM)
	if err != nil {
		return false
	}
	if cert.Subject.CommonName == wantDNS {
		return true
	}
	for _, d := range cert.DNSNames {
		if d == wantDNS {
			return true
		}
	}
	return false
}
