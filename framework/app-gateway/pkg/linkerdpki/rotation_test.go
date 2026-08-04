package linkerdpki

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"
)

// TC-PKI-G01: rotation threshold boundary (179d -> need; 181d -> !need).
func TestIssuerNeedsRotationBoundary(t *testing.T) {
	now := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)

	issuer179, err := pemCertWithNotAfter(now.Add(179 * 24 * time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	need, _, err := IssuerNeedsRotation(issuer179, now)
	if err != nil {
		t.Fatal(err)
	}
	if !need {
		t.Fatal("expected rotation when remaining < 180 days")
	}

	issuer181, err := pemCertWithNotAfter(now.Add(181 * 24 * time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	need, _, err = IssuerNeedsRotation(issuer181, now)
	if err != nil {
		t.Fatal(err)
	}
	if need {
		t.Fatal("expected no rotation when remaining >= 180 days")
	}
}

// TC-PKI-G02: an already-expired issuer needs rotation.
func TestIssuerNeedsRotationExpired(t *testing.T) {
	now := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)
	expired, err := pemCertWithNotAfter(now.Add(-24 * time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	need, remaining, err := IssuerNeedsRotation(expired, now)
	if err != nil {
		t.Fatal(err)
	}
	if !need {
		t.Fatal("expected rotation for expired issuer")
	}
	if remaining >= 0 {
		t.Fatalf("expected negative remaining, got %s", remaining)
	}
}

func TestCertificateNotAfter(t *testing.T) {
	end := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	pemBytes, err := pemCertWithNotAfter(end)
	if err != nil {
		t.Fatal(err)
	}
	got, err := certificateNotAfter(pemBytes)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(end) {
		t.Fatalf("notAfter: got %v want %v", got, end)
	}
}

// TC-PKI-G03: rotateIssuer produces a fresh ECDSA P-256 issuer valid for ~3 years.
func TestRotateIssuer(t *testing.T) {
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	caTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "root.linkerd.cluster.local"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour * 365 * 30),
		KeyUsage:     x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		IsCA:         true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTmpl, caTmpl, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	caCrtPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER})
	caKeyDER, err := x509.MarshalECPrivateKey(caKey)
	if err != nil {
		t.Fatal(err)
	}
	caKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: caKeyDER})

	mat, err := rotateIssuer(caCrtPEM, caKeyPEM)
	if err != nil {
		t.Fatal(err)
	}

	need, _, err := IssuerNeedsRotation(mat.IssuerCrt, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if need {
		t.Fatal("new issuer should be valid for >= 6 months")
	}

	block, _ := pem.Decode(mat.IssuerCrt)
	if block == nil {
		t.Fatal("issuer PEM did not decode")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := cert.PublicKey.(*ecdsa.PublicKey); !ok {
		t.Fatalf("issuer key is not ECDSA: %T", cert.PublicKey)
	}
	wantNotAfter := cert.NotBefore.Add(IssuerLifetimeDays * 24 * time.Hour)
	if !cert.NotAfter.Equal(wantNotAfter) {
		t.Fatalf("issuer notAfter: got %v want %v", cert.NotAfter, wantNotAfter)
	}
}

func pemCertWithNotAfter(notAfter time.Time) ([]byte, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test.linkerd.cluster.local"},
		NotBefore:    notAfter.Add(-24 * time.Hour),
		NotAfter:     notAfter,
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), nil
}
