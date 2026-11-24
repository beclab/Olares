// Package entrypki maintains the L4→app-gateway mTLS trust domain.
// It reuses Linkerd PKI rotation thresholds and ECDSA P-256 patterns but NEVER
// shares olares-linkerd-pki / mesh identity CA material.
package entrypki

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"maps"
	"math/big"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// SecretName is the server credential used by the Gateway listener.
	SecretName       = "olares-l4-agw-mtls"
	ClientSecretName = "olares-l4-agw-client"
	signerSecretName = "olares-l4-agw-signer"
	// DefaultNamespace hosts the L4→AGW channel Service.
	DefaultNamespace = "os-gateway"

	RotateThreshold  = 10 * 24 * time.Hour
	CALifetimeDays   = 10950
	LeafLifetimeDays = 30

	keyCACrt     = "ca.crt"
	keyCAKey     = "ca.key"
	keyServerCrt = "tls.crt"
	keyServerKey = "tls.key"
	keyClientCrt = "client.crt"
	keyClientKey = "client.key"
)

// Material is Entry mTLS vault content.
type Material struct {
	CACrt, CAKey         []byte
	ServerCrt, ServerKey []byte
	ClientCrt, ClientKey []byte
}

// Enabled reports whether Entry PKI maintain should run.
func Enabled(raw string) bool {
	return raw == "1" || raw == "true" || raw == "TRUE" || raw == "yes"
}

// NeedsRotation reports whether a leaf is due for renewal or exceeds the current lifetime.
func NeedsRotation(certPEM []byte, now time.Time) (bool, time.Duration, error) {
	cert, err := parseCertificate(certPEM)
	if err != nil {
		return false, 0, err
	}
	rem := cert.NotAfter.Sub(now)
	return rem <= RotateThreshold || cert.NotAfter.Sub(cert.NotBefore) > LeafLifetimeDays*24*time.Hour, rem, nil
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
	var signer corev1.Secret
	err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: signerSecretName}, &signer)
	if apierrors.IsNotFound(err) {
		// Persist the signer before redacting the legacy combined Secret.
		// Retain its trust anchor during migration. Publish the isolated
		// client material before redacting the legacy Secret.
		var legacy corev1.Secret
		getErr := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: SecretName}, &legacy)
		var mat *Material
		switch {
		case getErr == nil:
			mat, err = materialFromSecret(&legacy)
		case apierrors.IsNotFound(getErr):
			// A surviving client credential means this is recovery, not bootstrap.
			// Generating a new CA here would silently replace the trust anchor.
			getErr := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: ClientSecretName}, &legacy)
			if getErr == nil {
				return fmt.Errorf("entry signer missing while client credential exists")
			}
			if !apierrors.IsNotFound(getErr) {
				return getErr
			}
			mat, err = bootstrapMaterial(now)
		default:
			return getErr
		}
		if err != nil {
			return err
		}
		signer = corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: signerSecretName},
			Data: map[string][]byte{keyCACrt: mat.CACrt, keyCAKey: mat.CAKey}}
		if err := c.Create(ctx, &signer); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	mat, err := materialFromSecret(&signer)
	if err != nil {
		return err
	}
	if err := reconcileLeaf(ctx, c, ns, ClientSecretName, corev1.SecretTypeOpaque, mat,
		keyClientCrt, keyClientKey, "l4-bfl-proxy.entry.client", x509.ExtKeyUsageClientAuth, now); err != nil {
		return err
	}
	return reconcileLeaf(ctx, c, ns, SecretName, corev1.SecretTypeTLS, mat,
		keyServerCrt, keyServerKey, "app-gateway-l4-mtls.os-gateway.svc.cluster.local", x509.ExtKeyUsageServerAuth, now)
}

// reconcileLeaf repairs incomplete credentials and only publishes material
// intended for the leaf consumer. The signing key never leaves the signer.
func reconcileLeaf(ctx context.Context, c client.Client, ns, name string, secretType corev1.SecretType,
	mat *Material, certKey, privateKey, dnsName string, usage x509.ExtKeyUsage, now time.Time) error {
	var sec corev1.Secret
	err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: name}, &sec)
	isNew := apierrors.IsNotFound(err)
	if err != nil && !isNew {
		return err
	}
	cert, key := sec.Data[certKey], sec.Data[privateKey]
	_, pairErr := tls.X509KeyPair(cert, key)
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(mat.CACrt) {
		return fmt.Errorf("invalid Entry CA certificate")
	}
	rotate, _, expiryErr := NeedsRotation(cert, now)
	if pairErr == nil && expiryErr == nil && !rotate {
		var leaf *x509.Certificate
		leaf, err = parseCertificate(cert)
		if err == nil {
			_, err = leaf.Verify(x509.VerifyOptions{Roots: roots, DNSName: dnsName,
				KeyUsages: []x509.ExtKeyUsage{usage}, CurrentTime: now})
		}
	}
	if pairErr != nil || expiryErr != nil || rotate || err != nil {
		ca, err := parseCertificate(mat.CACrt)
		if err != nil {
			return err
		}
		caKey, err := parseECPrivateKey(mat.CAKey)
		if err != nil {
			return err
		}
		cert, key, err = issueLeaf(ca, caKey, dnsName, usage, now)
		if err != nil {
			return err
		}
	}
	data := map[string][]byte{keyCACrt: mat.CACrt, certKey: cert, privateKey: key}
	if isNew {
		sec = corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name}, Type: secretType, Data: data}
		return c.Create(ctx, &sec)
	}
	if maps.EqualFunc(sec.Data, data, bytes.Equal) {
		return nil
	}
	// Secret type is immutable. Existing Opaque server Secrets remain Opaque.
	sec.Data = data
	return c.Update(ctx, &sec)
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
	serverCrt, serverKey, err := issueLeaf(caCert, caKey, "app-gateway-l4-mtls.os-gateway.svc.cluster.local", x509.ExtKeyUsageServerAuth, now)
	if err != nil {
		return nil, err
	}
	clientCrt, clientKey, err := issueLeaf(caCert, caKey, "l4-bfl-proxy.entry.client", x509.ExtKeyUsageClientAuth, now)
	if err != nil {
		return nil, err
	}
	return &Material{
		CACrt: caPEM, CAKey: caKeyPEM,
		ServerCrt: serverCrt, ServerKey: serverKey,
		ClientCrt: clientCrt, ClientKey: clientKey,
	}, nil
}

func materialFromSecret(sec *corev1.Secret) (*Material, error) {
	m := &Material{
		CACrt: sec.Data[keyCACrt], CAKey: sec.Data[keyCAKey],
		ServerCrt: sec.Data[keyServerCrt], ServerKey: sec.Data[keyServerKey],
		ClientCrt: sec.Data[keyClientCrt], ClientKey: sec.Data[keyClientKey],
	}
	if len(m.CACrt) == 0 || len(m.CAKey) == 0 {
		return nil, fmt.Errorf("entry pki secret missing ca material")
	}
	return m, nil
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

func issueLeaf(ca *x509.Certificate, caKey *ecdsa.PrivateKey, cn string, usage x509.ExtKeyUsage, now time.Time) ([]byte, []byte, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}
	notBefore := now.Add(-time.Hour)
	tpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    notBefore,
		NotAfter:     notBefore.Add(LeafLifetimeDays * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{usage},
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
