package linkerdpki

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"strings"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// TC-PKI-G04: when olares-linkerd-pki is absent, MaintainLinkerdPKI returns a
// transient error and never terminates the process.
func TestMaintainLinkerdPKISecretNotFound(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	c := fake.NewClientBuilder().WithScheme(scheme).Build()

	err := MaintainLinkerdPKI(context.Background(), c, DefaultLinkerdNamespace)
	if err == nil {
		t.Fatal("expected error when olares-linkerd-pki secret is missing")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMaintainLinkerdPKI_FreshIssuerNoRotation(t *testing.T) {
	ctx := context.Background()
	ns := DefaultLinkerdNamespace
	mat, err := testCAAndIssuerMaterial(time.Now().UTC().Add(200 * 24 * time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	secret := testPKISecret(ns, mat)
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(secret).Build()

	before := string(secret.Data[pkiIssuerCrtKey])
	if err := MaintainLinkerdPKI(ctx, c, ns); err != nil {
		t.Fatalf("maintain fresh issuer: %v", err)
	}
	var got corev1.Secret
	if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: PKISecretName}, &got); err != nil {
		t.Fatalf("get secret: %v", err)
	}
	if string(got.Data[pkiIssuerCrtKey]) != before {
		t.Fatal("expected fresh issuer to remain unchanged")
	}
}

func TestMaintainLinkerdPKI_NearExpiryRotatesIssuer(t *testing.T) {
	ctx := context.Background()
	ns := DefaultLinkerdNamespace
	mat, err := testCAAndIssuerMaterial(time.Now().UTC().Add(179 * 24 * time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	secret := testPKISecret(ns, mat)
	identityIssuer := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: identityIssuerSecret, Namespace: ns},
		Type:       corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			identityIssuerCrtKey: mat.IssuerCrt,
			identityIssuerKeyKey: mat.IssuerKey,
		},
	}
	identityDep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: identityDeployment, Namespace: ns},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{}},
			},
		},
	}
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := appsv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	c := fake.NewClientBuilder().WithScheme(scheme).
		WithObjects(secret, identityIssuer, identityDep).
		Build()

	before := string(secret.Data[pkiIssuerCrtKey])
	if err := MaintainLinkerdPKI(ctx, c, ns); err != nil {
		t.Fatalf("maintain near-expiry issuer: %v", err)
	}
	var gotSecret corev1.Secret
	if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: PKISecretName}, &gotSecret); err != nil {
		t.Fatalf("get pki secret: %v", err)
	}
	if string(gotSecret.Data[pkiIssuerCrtKey]) == before {
		t.Fatal("expected issuer certificate to rotate when near expiry")
	}
	var gotDep appsv1.Deployment
	if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: identityDeployment}, &gotDep); err != nil {
		t.Fatalf("get identity deployment: %v", err)
	}
	if gotDep.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] == "" {
		t.Fatal("expected linkerd-identity deployment restart annotation after rotation")
	}
}

func testPKISecret(ns string, mat *Material) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: PKISecretName, Namespace: ns},
		Type:       corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			pkiCACrtKey:     mat.CACrt,
			pkiCAKeyKey:     mat.CAKey,
			pkiIssuerCrtKey: mat.IssuerCrt,
			pkiIssuerKeyKey: mat.IssuerKey,
			pkiMetadataKey:  []byte(`{"version":1}`),
		},
	}
}

func testCAAndIssuerMaterial(issuerNotAfter time.Time) (*Material, error) {
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	caCrtPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER})
	caKeyDER, err := x509.MarshalECPrivateKey(caKey)
	if err != nil {
		return nil, err
	}
	caKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: caKeyDER})

	issuerKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	issuerTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "identity.linkerd.cluster.local"},
		NotBefore:    issuerNotAfter.Add(-24 * time.Hour),
		NotAfter:     issuerNotAfter,
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	issuerDER, err := x509.CreateCertificate(rand.Reader, issuerTmpl, caTmpl, &issuerKey.PublicKey, caKey)
	if err != nil {
		return nil, err
	}
	issuerCrtPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: issuerDER})
	issuerKeyDER, err := x509.MarshalECPrivateKey(issuerKey)
	if err != nil {
		return nil, err
	}
	issuerKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: issuerKeyDER})

	return &Material{
		CACrt:     caCrtPEM,
		CAKey:     caKeyPEM,
		IssuerCrt: issuerCrtPEM,
		IssuerKey: issuerKeyPEM,
	}, nil
}
