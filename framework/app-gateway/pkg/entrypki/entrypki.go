// Package entrypki maintains the L4→app-gateway mTLS trust domain.
// It reuses Linkerd PKI rotation thresholds and ECDSA P-256 patterns but NEVER
// shares olares-linkerd-pki / mesh identity CA material.
package entrypki

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// SecretName holds Entry CA + server + client PEMs in os-gateway.
	SecretName = "olares-l4-agw-mtls"
	// DefaultNamespace hosts the L4→AGW channel Service.
	DefaultNamespace = "os-gateway"

	RotateThreshold = 180 * 24 * time.Hour
	CALifetimeDays   = 10950
	LeafLifetimeDays = 365

	keyCACrt           = "ca.crt"
	keyCAKey           = "ca.key"
	keyCACrtPrevious    = "ca.crt.previous"
	keyServerCrt       = "tls.crt"
	keyServerKey       = "tls.key"
	keyClientCrt       = "client.crt"
	keyClientKey       = "client.key"
	annotFingerprint   = "gateway.olares.io/entry-mtls-fingerprint"
)

// Material is Entry mTLS vault content.
type Material struct {
	CACrt, CAKey             []byte
	CACrtPrevious            []byte
	ServerCrt, ServerKey     []byte
	ClientCrt, ClientKey     []byte
}

// Enabled reports whether Entry PKI maintain should run.
func Enabled(raw string) bool {
	return raw == "1" || raw == "true" || raw == "TRUE" || raw == "yes"
}

// NeedsRotation reports whether leaf cert remaining life is below threshold.
func NeedsRotation(certPEM []byte, now time.Time) (bool, time.Duration, error) {
	notAfter, err := certNotAfter(certPEM)
	if err != nil {
		return false, 0, err
	}
	rem := notAfter.Sub(now)
	return rem < RotateThreshold, rem, nil
}

// MaintainEntryMTLS ensures the Entry Secret exists and rotates leaves near expiry.
// When enabled is false, this is a no-op (does not create or mutate Secrets).
func MaintainEntryMTLS(ctx context.Context, c client.Client, ns string, enabled bool, now time.Time) error {
	if !enabled {
		return nil
	}
	if ns == "" {
		ns = DefaultNamespace
	}
	var sec corev1.Secret
	err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: SecretName}, &sec)
	if apierrors.IsNotFound(err) {
		mat, err := bootstrapMaterial(now)
		if err != nil {
			return err
		}
		return writeSecret(ctx, c, ns, mat, nil)
	}
	if err != nil {
		return err
	}
	serverCrt := sec.Data[keyServerCrt]
	need, _, err := NeedsRotation(serverCrt, now)
	if err != nil {
		return err
	}
	if !need {
		return nil
	}
	mat, err := materialFromSecret(&sec)
	if err != nil {
		return err
	}
	rotated, err := rotateLeaves(mat, now)
	if err != nil {
		return err
	}
	return writeSecret(ctx, c, ns, rotated, &sec)
}

func bootstrapMaterial(now time.Time) (*Material, error) {
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	caCert, caPEM, err := createCA(caKey, now)
	if err != nil {
		return nil, err
	}
	caKeyPEM, err := marshalECPrivateKey(caKey)
	if err != nil {
		return nil, err
	}
	serverCrt, serverKey, err := issueLeaf(caCert, caKey, "app-gateway-l4-mtls.os-gateway.svc.cluster.local", now)
	if err != nil {
		return nil, err
	}
	clientCrt, clientKey, err := issueLeaf(caCert, caKey, "l4-bfl-proxy.entry.client", now)
	if err != nil {
		return nil, err
	}
	return &Material{
		CACrt: caPEM, CAKey: caKeyPEM,
		ServerCrt: serverCrt, ServerKey: serverKey,
		ClientCrt: clientCrt, ClientKey: clientKey,
	}, nil
}

func rotateLeaves(mat *Material, now time.Time) (*Material, error) {
	caCert, err := parseCertificate(mat.CACrt)
	if err != nil {
		return nil, err
	}
	caKey, err := parseECPrivateKey(mat.CAKey)
	if err != nil {
		return nil, err
	}
	serverCrt, serverKey, err := issueLeaf(caCert, caKey, "app-gateway-l4-mtls.os-gateway.svc.cluster.local", now)
	if err != nil {
		return nil, err
	}
	clientCrt, clientKey, err := issueLeaf(caCert, caKey, "l4-bfl-proxy.entry.client", now)
	if err != nil {
		return nil, err
	}
	return &Material{
		CACrt: mat.CACrt, CAKey: mat.CAKey,
		CACrtPrevious: mat.CACrt, // dual-trust window
		ServerCrt: serverCrt, ServerKey: serverKey,
		ClientCrt: clientCrt, ClientKey: clientKey,
	}, nil
}

func materialFromSecret(sec *corev1.Secret) (*Material, error) {
	m := &Material{
		CACrt: sec.Data[keyCACrt], CAKey: sec.Data[keyCAKey],
		CACrtPrevious: sec.Data[keyCACrtPrevious],
		ServerCrt: sec.Data[keyServerCrt], ServerKey: sec.Data[keyServerKey],
		ClientCrt: sec.Data[keyClientCrt], ClientKey: sec.Data[keyClientKey],
	}
	if len(m.CACrt) == 0 || len(m.CAKey) == 0 {
		return nil, fmt.Errorf("entry pki secret missing ca material")
	}
	return m, nil
}

func writeSecret(ctx context.Context, c client.Client, ns string, mat *Material, existing *corev1.Secret) error {
	fp := fingerprint(mat.ServerCrt)
	data := map[string][]byte{
		keyCACrt: mat.CACrt, keyCAKey: mat.CAKey,
		keyServerCrt: mat.ServerCrt, keyServerKey: mat.ServerKey,
		keyClientCrt: mat.ClientCrt, keyClientKey: mat.ClientKey,
	}
	if len(mat.CACrtPrevious) > 0 {
		data[keyCACrtPrevious] = mat.CACrtPrevious
	}
	if existing == nil {
		sec := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      SecretName,
				Namespace: ns,
				Labels: map[string]string{
					"app.kubernetes.io/name":      "l4-agw-mtls",
					"app.kubernetes.io/component": "entry-pki",
				},
				Annotations: map[string]string{annotFingerprint: fp},
			},
			Type: corev1.SecretTypeOpaque,
			Data: data,
		}
		return c.Create(ctx, sec)
	}
	existing.Data = data
	if existing.Annotations == nil {
		existing.Annotations = map[string]string{}
	}
	existing.Annotations[annotFingerprint] = fp
	return c.Update(ctx, existing)
}

func createCA(key *ecdsa.PrivateKey, now time.Time) (*x509.Certificate, []byte, error) {
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}
	tpl := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: "olares-entry-mtls-ca"},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(time.Duration(CALifetimeDays) * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            0,
	}
	der, err := x509.CreateCertificate(rand.Reader, tpl, tpl, &key.PublicKey, key)
	if err != nil {
		return nil, nil, err
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, nil, err
	}
	return cert, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), nil
}

func issueLeaf(ca *x509.Certificate, caKey *ecdsa.PrivateKey, cn string, now time.Time) ([]byte, []byte, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}
	tpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    now.Add(-time.Hour),
		NotAfter:     now.Add(time.Duration(LeafLifetimeDays) * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		DNSNames:     []string{cn},
	}
	der, err := x509.CreateCertificate(rand.Reader, tpl, ca, &key.PublicKey, caKey)
	if err != nil {
		return nil, nil, err
	}
	keyPEM, err := marshalECPrivateKey(key)
	if err != nil {
		return nil, nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), keyPEM, nil
}

func certNotAfter(pemBytes []byte) (time.Time, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return time.Time{}, fmt.Errorf("invalid PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return time.Time{}, err
	}
	return cert.NotAfter, nil
}

func parseCertificate(pemBytes []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM")
	}
	return x509.ParseCertificate(block.Bytes)
}

func parseECPrivateKey(pemBytes []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("invalid key PEM")
	}
	if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	k, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	ec, ok := k.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not ECDSA key")
	}
	return ec, nil
}

func marshalECPrivateKey(key *ecdsa.PrivateKey) ([]byte, error) {
	b, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: b}), nil
}

func fingerprint(certPEM []byte) string {
	sum := sha256.Sum256(certPEM)
	return hex.EncodeToString(sum[:8])
}
