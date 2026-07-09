package linkerdpki

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// BootstrapIfMissing ensures olares-linkerd-pki exists with valid CA+issuer material.
// If the Secret already exists with required keys, returns (false, nil) without mutation.
func BootstrapIfMissing(ctx context.Context, c client.Client, linkerdNS string) (created bool, err error) {
	mat, ok, err := loadPKISecret(ctx, c, linkerdNS)
	if err != nil {
		slog.Error("linkerd pki bootstrap load failed", "op", "bootstrap", "namespace", linkerdNS, "error", err)
		return false, fmt.Errorf("linkerd_pki_bootstrap_corrupt_secret: %w", err)
	}
	if ok {
		return false, nil
	}
	mat, err = generateInitialMaterial()
	if err != nil {
		slog.Error("linkerd pki bootstrap generate failed", "op", "bootstrap", "namespace", linkerdNS, "error", err)
		return false, fmt.Errorf("linkerd_pki_bootstrap_generate_failed: %w", err)
	}
	if err := writePKISecret(ctx, c, linkerdNS, mat, 1); err != nil {
		slog.Error("linkerd pki bootstrap write secret failed", "op", "bootstrap", "namespace", linkerdNS, "error", err)
		return false, fmt.Errorf("linkerd_pki_bootstrap_write_secret_failed: %w", err)
	}
	slog.Info("linkerd pki bootstrap created vault", "op", "bootstrap", "namespace", linkerdNS, "created", true)
	return true, nil
}

// SyncIdentityToLinkerd copies issuer/trust from olares-linkerd-pki into Linkerd identity
// mount points. Idempotent: no-op when remote PEM already matches material.
func SyncIdentityToLinkerd(ctx context.Context, c client.Client, linkerdNS string) (changed bool, err error) {
	mat, ok, err := loadPKISecret(ctx, c, linkerdNS)
	if err != nil {
		slog.Error("linkerd pki sync load vault failed", "op", "bootstrap", "namespace", linkerdNS, "error", err)
		return false, fmt.Errorf("linkerd_pki_bootstrap_corrupt_secret: %w", err)
	}
	if !ok {
		slog.Error("linkerd pki sync vault missing", "op", "bootstrap", "namespace", linkerdNS)
		return false, fmt.Errorf("linkerd_pki_sync_issuer_missing: vault %s not found", PKISecretName)
	}
	issuerChanged, err := syncIdentityIssuerSecret(ctx, c, linkerdNS, mat)
	if err != nil {
		slog.Error("linkerd pki sync issuer failed", "op", "bootstrap", "namespace", linkerdNS, "error", err)
		return false, fmt.Errorf("linkerd_pki_sync_patch_failed: %w", err)
	}
	trustChanged, err := syncIdentityTrustRoots(ctx, c, linkerdNS, mat.CACrt)
	if err != nil {
		slog.Error("linkerd pki sync trust roots failed", "op", "bootstrap", "namespace", linkerdNS, "error", err)
		return false, fmt.Errorf("linkerd_pki_sync_patch_failed: %w", err)
	}
	changed = issuerChanged || trustChanged
	slog.Info("linkerd pki sync identity complete", "op", "bootstrap", "namespace", linkerdNS, "sync_changed", changed)
	return changed, nil
}

func generateInitialMaterial() (*Material, error) {
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	notBefore := time.Now().UTC().Add(-time.Hour)
	notAfter := notBefore.Add(CALifetimeDays * 24 * time.Hour)
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, err
	}
	caTemplate := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: "root.linkerd.cluster.local"},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, err
	}
	caCrtPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER})
	caKeyPEM, err := marshalECPrivateKey(caKey)
	if err != nil {
		return nil, err
	}
	return rotateIssuer(caCrtPEM, caKeyPEM)
}

func syncIdentityIssuerSecret(ctx context.Context, c client.Client, ns string, mat *Material) (bool, error) {
	var sec corev1.Secret
	err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: identityIssuerSecret}, &sec)
	if apierrors.IsNotFound(err) {
		desired := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      identityIssuerSecret,
				Namespace: ns,
				Labels: map[string]string{
					"linkerd.io/control-plane-component": "identity",
					"linkerd.io/control-plane-ns":        ns,
				},
			},
			Type: corev1.SecretTypeOpaque,
			Data: map[string][]byte{
				identityIssuerCrtKey: mat.IssuerCrt,
				identityIssuerKeyKey: mat.IssuerKey,
			},
		}
		if err := c.Create(ctx, desired); err != nil {
			return false, err
		}
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("get linkerd-identity-issuer: %w", err)
	}
	return patchIdentityIssuerIfChanged(ctx, c, &sec, mat)
}

func patchIdentityIssuerIfChanged(ctx context.Context, c client.Client, sec *corev1.Secret, mat *Material) (bool, error) {
	if sec.Data == nil {
		sec.Data = map[string][]byte{}
	}
	if bytesEqual(sec.Data[identityIssuerCrtKey], mat.IssuerCrt) &&
		bytesEqual(sec.Data[identityIssuerKeyKey], mat.IssuerKey) {
		return false, nil
	}
	sec.Data[identityIssuerCrtKey] = mat.IssuerCrt
	sec.Data[identityIssuerKeyKey] = mat.IssuerKey
	return true, c.Update(ctx, sec)
}

func syncIdentityTrustRoots(ctx context.Context, c client.Client, ns string, caCrt []byte) (bool, error) {
	var cm corev1.ConfigMap
	err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: identityTrustRootsCM}, &cm)
	if apierrors.IsNotFound(err) {
		desired := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      identityTrustRootsCM,
				Namespace: ns,
				Labels: map[string]string{
					"linkerd.io/control-plane-component": "controller",
					"linkerd.io/control-plane-ns":        ns,
				},
			},
			Data: map[string]string{
				identityTrustRootsKey: string(caCrt),
			},
		}
		if err := c.Create(ctx, desired); err != nil {
			return false, err
		}
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("get linkerd-identity-trust-roots: %w", err)
	}
	if cm.Data == nil {
		cm.Data = map[string]string{}
	}
	desired := string(caCrt)
	if cm.Data[identityTrustRootsKey] == desired {
		return false, nil
	}
	cm.Data[identityTrustRootsKey] = desired
	return true, c.Update(ctx, &cm)
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
