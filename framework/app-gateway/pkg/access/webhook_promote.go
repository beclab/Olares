package access

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	AdmissionWebhookConfigName = "access-olares-validate"
	AdmissionTLSSecretName     = "access-admission-tls"
	AdmissionTLSNamespace      = "os-app-access"
	AdmissionServiceName       = "access-admission"
)

// PrepareAdmissionTLS bootstraps the serving credential without changing
// the API server's webhook failure policy.
func PrepareAdmissionTLS(ctx context.Context, c client.Client) ([]byte, error) {
	return ensureAdmissionTLS(ctx, c, AdmissionTLSNamespace, AdmissionTLSSecretName)
}

// PromoteAdmissionFailurePolicy pins the verified CA and enforces Fail after
// the caller has checked the serving endpoint.
func PromoteAdmissionFailurePolicy(ctx context.Context, c client.Client, caBundle []byte) error {
	if len(caBundle) == 0 {
		return fmt.Errorf("webhook CA bundle is empty")
	}
	var wh admissionregistrationv1.ValidatingWebhookConfiguration
	if err := c.Get(ctx, types.NamespacedName{Name: AdmissionWebhookConfigName}, &wh); err != nil {
		return err
	}
	if len(wh.Webhooks) == 0 {
		return fmt.Errorf("webhook config %s has no webhooks", AdmissionWebhookConfigName)
	}
	fail := admissionregistrationv1.Fail
	changed := false
	for i := range wh.Webhooks {
		w := &wh.Webhooks[i]
		if w.FailurePolicy == nil || *w.FailurePolicy != fail || !bytes.Equal(w.ClientConfig.CABundle, caBundle) {
			w.FailurePolicy = &fail
			w.ClientConfig.CABundle = caBundle
			changed = true
		}
	}
	if !changed {
		return nil
	}
	return c.Update(ctx, &wh)
}

func ensureAdmissionTLS(ctx context.Context, c client.Client, ns, name string) ([]byte, error) {
	var sec corev1.Secret
	getErr := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: name}, &sec)
	if getErr == nil {
		if ca := sec.Data["ca.crt"]; len(ca) > 0 {
			if len(sec.Data["tls.crt"]) > 0 && len(sec.Data["tls.key"]) > 0 {
				return ca, nil
			}
		}
	}
	if getErr != nil && !apierrors.IsNotFound(getErr) {
		return nil, getErr
	}
	caPEM, certPEM, keyPEM, err := selfSignedWebhookCert(AdmissionServiceName + "." + ns + ".svc")
	if err != nil {
		return nil, err
	}
	out := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels:    map[string]string{"app.kubernetes.io/name": "access-admission"},
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			"ca.crt":  caPEM,
			"tls.crt": certPEM,
			"tls.key": keyPEM,
		},
	}
	if apierrors.IsNotFound(getErr) {
		if err := c.Create(ctx, out); err != nil {
			return nil, err
		}
		return caPEM, nil
	}
	sec.Data = out.Data
	sec.Type = out.Type
	if err := c.Update(ctx, &sec); err != nil {
		return nil, err
	}
	return caPEM, nil
}

func selfSignedWebhookCert(dnsName string) (caPEM, certPEM, keyPEM []byte, err error) {
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, nil, err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, nil, err
	}
	now := time.Now().UTC()
	caTpl := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: "access-admission-ca"},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(10 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTpl, caTpl, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, nil, err
	}
	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		return nil, nil, nil, err
	}
	leafKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, nil, err
	}
	leafSerial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, nil, err
	}
	leafTpl := &x509.Certificate{
		SerialNumber: leafSerial,
		Subject:      pkix.Name{CommonName: dnsName},
		NotBefore:    now.Add(-time.Hour),
		NotAfter:     now.Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{dnsName, AdmissionServiceName, AdmissionServiceName + "." + AdmissionTLSNamespace, AdmissionServiceName + "." + AdmissionTLSNamespace + ".svc"},
	}
	leafDER, err := x509.CreateCertificate(rand.Reader, leafTpl, caCert, &leafKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, nil, err
	}
	keyDER, err := x509.MarshalECPrivateKey(leafKey)
	if err != nil {
		return nil, nil, nil, err
	}
	caPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER})
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: leafDER})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	return caPEM, certPEM, keyPEM, nil
}
