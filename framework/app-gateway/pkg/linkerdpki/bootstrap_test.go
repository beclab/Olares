package linkerdpki

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// TC-EG6-001: greenfield bootstrap creates olares-linkerd-pki with real CA material.
func TestBootstrapIfMissing_CreatesVault(t *testing.T) {
	ctx := context.Background()
	ns := DefaultLinkerdNamespace
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	c := fake.NewClientBuilder().WithScheme(scheme).Build()

	created, err := BootstrapIfMissing(ctx, c, ns)
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	if !created {
		t.Fatal("expected created=true on greenfield")
	}
	var sec corev1.Secret
	if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: PKISecretName}, &sec); err != nil {
		t.Fatalf("get vault: %v", err)
	}
	subject, err := pemSubjectCN(sec.Data[pkiCACrtKey])
	if err != nil {
		t.Fatalf("parse ca.crt: %v", err)
	}
	if subject != "root.linkerd.cluster.local" {
		t.Fatalf("unexpected CA subject CN: %q", subject)
	}
}

// TC-EG6-002: existing vault is left unchanged (scheme A).
func TestBootstrapIfMissing_ExistingVaultNoOp(t *testing.T) {
	ctx := context.Background()
	ns := DefaultLinkerdNamespace
	mat, err := testCAAndIssuerMaterial(mustParseTime(t, "2030-01-01T00:00:00Z"))
	if err != nil {
		t.Fatal(err)
	}
	secret := testPKISecret(ns, mat)
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(secret).Build()
	beforeKey := append([]byte(nil), secret.Data[pkiCAKeyKey]...)

	created, err := BootstrapIfMissing(ctx, c, ns)
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	if created {
		t.Fatal("expected created=false when vault exists")
	}
	var got corev1.Secret
	if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: PKISecretName}, &got); err != nil {
		t.Fatalf("get vault: %v", err)
	}
	if string(got.Data[pkiCAKeyKey]) != string(beforeKey) {
		t.Fatal("ca.key must not change on existing vault")
	}
}

// TC-EG6-003: SyncIdentityToLinkerd is idempotent when mount points already match.
func TestSyncIdentityToLinkerd_Idempotent(t *testing.T) {
	ctx := context.Background()
	ns := DefaultLinkerdNamespace
	mat, err := testCAAndIssuerMaterial(mustParseTime(t, "2030-01-01T00:00:00Z"))
	if err != nil {
		t.Fatal(err)
	}
	secret := testPKISecret(ns, mat)
	issuer := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: identityIssuerSecret, Namespace: ns},
		Type:       corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			identityIssuerCrtKey: mat.IssuerCrt,
			identityIssuerKeyKey: mat.IssuerKey,
		},
	}
	trust := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: identityTrustRootsCM, Namespace: ns},
		Data:       map[string]string{identityTrustRootsKey: string(mat.CACrt)},
	}
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(secret, issuer, trust).Build()

	changed1, err := SyncIdentityToLinkerd(ctx, c, ns)
	if err != nil {
		t.Fatalf("first sync: %v", err)
	}
	if changed1 {
		t.Fatal("expected changed=false when already aligned")
	}
	changed2, err := SyncIdentityToLinkerd(ctx, c, ns)
	if err != nil {
		t.Fatalf("second sync: %v", err)
	}
	if changed2 {
		t.Fatal("expected changed=false on repeat sync")
	}
}

func pemSubjectCN(pemBytes []byte) (string, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return "", x509.ErrUnsupportedAlgorithm
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", err
	}
	return cert.Subject.CommonName, nil
}

func mustParseTime(t *testing.T, raw string) time.Time {
	t.Helper()
	ts, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		t.Fatal(err)
	}
	return ts
}
