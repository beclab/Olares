package linkerdpki

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"strings"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

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
	fp := issuerFingerprint(mat.IssuerCrt)
	secret := testPKISecret(ns, mat)
	identityIssuer := testIdentityIssuerSecret(ns, mat, true)
	identityDep := testIdentityDeployment(ns, fp, "already-rolled")
	scheme := testScheme(t)
	c := fake.NewClientBuilder().WithScheme(scheme).
		WithObjects(secret, identityIssuer, identityDep).
		Build()

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
	var gotDep appsv1.Deployment
	if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: identityDeployment}, &gotDep); err != nil {
		t.Fatalf("get identity deployment: %v", err)
	}
	if gotDep.Spec.Template.Annotations[restartedAtAnnotation] != "already-rolled" {
		t.Fatal("expected no identity rollout when vault and fingerprint already match")
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
	identityIssuer := testIdentityIssuerSecret(ns, mat, false)
	identityDep := testIdentityDeployment(ns, "", "")
	scheme := testScheme(t)
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
	var gotIdentity corev1.Secret
	if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: identityIssuerSecret}, &gotIdentity); err != nil {
		t.Fatalf("get identity issuer: %v", err)
	}
	if string(gotIdentity.Data[identityIssuerCrtKey]) != string(gotSecret.Data[pkiIssuerCrtKey]) {
		t.Fatal("expected identity issuer to match rotated vault")
	}
	var gotDep appsv1.Deployment
	if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: identityDeployment}, &gotDep); err != nil {
		t.Fatalf("get identity deployment: %v", err)
	}
	if gotDep.Spec.Template.Annotations[restartedAtAnnotation] == "" {
		t.Fatal("expected linkerd-identity deployment restart annotation after rotation")
	}
	wantFP := issuerFingerprint(gotSecret.Data[pkiIssuerCrtKey])
	if gotDep.Annotations[identityIssuerFingerprintAnnotation] != wantFP {
		t.Fatalf("fingerprint = %q, want %q", gotDep.Annotations[identityIssuerFingerprintAnnotation], wantFP)
	}
}

func TestMaintainLinkerdPKI_HealsStaleIdentityIssuer(t *testing.T) {
	ctx := context.Background()
	ns := DefaultLinkerdNamespace
	vault, err := testCAAndIssuerMaterial(time.Now().UTC().Add(200 * 24 * time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	stale, err := testCAAndIssuerMaterial(time.Now().UTC().Add(190 * 24 * time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	secret := testPKISecret(ns, vault)
	identityIssuer := testIdentityIssuerSecret(ns, stale, false)
	identityDep := testIdentityDeployment(ns, issuerFingerprint(stale.IssuerCrt), "old")
	scheme := testScheme(t)
	c := fake.NewClientBuilder().WithScheme(scheme).
		WithObjects(secret, identityIssuer, identityDep).
		Build()

	beforeVault := string(secret.Data[pkiIssuerCrtKey])
	if err := MaintainLinkerdPKI(ctx, c, ns); err != nil {
		t.Fatalf("heal stale identity: %v", err)
	}
	var gotVault corev1.Secret
	if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: PKISecretName}, &gotVault); err != nil {
		t.Fatalf("get vault: %v", err)
	}
	if string(gotVault.Data[pkiIssuerCrtKey]) != beforeVault {
		t.Fatal("expected fresh vault issuer to remain unchanged")
	}
	var gotIdentity corev1.Secret
	if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: identityIssuerSecret}, &gotIdentity); err != nil {
		t.Fatalf("get identity issuer: %v", err)
	}
	if string(gotIdentity.Data[identityIssuerCrtKey]) != beforeVault {
		t.Fatal("expected identity issuer to be rewritten from vault")
	}
	var gotDep appsv1.Deployment
	if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: identityDeployment}, &gotDep); err != nil {
		t.Fatalf("get identity deployment: %v", err)
	}
	if gotDep.Annotations[identityIssuerFingerprintAnnotation] != issuerFingerprint(vault.IssuerCrt) {
		t.Fatal("expected fingerprint updated to vault issuer")
	}
	if gotDep.Spec.Template.Annotations[restartedAtAnnotation] == "old" {
		t.Fatal("expected identity rollout when fingerprint lagged vault")
	}
}

func TestMaintainLinkerdPKI_IdempotentWhenConverged(t *testing.T) {
	ctx := context.Background()
	ns := DefaultLinkerdNamespace
	mat, err := testCAAndIssuerMaterial(time.Now().UTC().Add(200 * 24 * time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	fp := issuerFingerprint(mat.IssuerCrt)
	scheme := testScheme(t)
	c := fake.NewClientBuilder().WithScheme(scheme).
		WithObjects(
			testPKISecret(ns, mat),
			testIdentityIssuerSecret(ns, mat, true),
			testIdentityDeployment(ns, fp, "stable"),
		).
		Build()

	if err := MaintainLinkerdPKI(ctx, c, ns); err != nil {
		t.Fatalf("maintain converged state: %v", err)
	}
	var gotDep appsv1.Deployment
	if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: identityDeployment}, &gotDep); err != nil {
		t.Fatalf("get identity deployment: %v", err)
	}
	if gotDep.Spec.Template.Annotations[restartedAtAnnotation] != "stable" {
		t.Fatal("expected no rollout when already converged")
	}
}

func TestMaintainLinkerdPKI_PartialSyncHealsOnRetry(t *testing.T) {
	ctx := context.Background()
	ns := DefaultLinkerdNamespace
	vault, err := testCAAndIssuerMaterial(time.Now().UTC().Add(200 * 24 * time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	stale, err := testCAAndIssuerMaterial(time.Now().UTC().Add(190 * 24 * time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	scheme := testScheme(t)
	objs := []client.Object{
		testPKISecret(ns, vault),
		testIdentityIssuerSecret(ns, stale, false),
		testIdentityDeployment(ns, issuerFingerprint(stale.IssuerCrt), "old"),
	}
	failing := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).
		WithInterceptorFuncs(interceptor.Funcs{
			Update: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
				if sec, ok := obj.(*corev1.Secret); ok && sec.Name == identityIssuerSecret {
					return errors.New("simulated identity issuer update failure")
				}
				return c.Update(ctx, obj, opts...)
			},
		}).Build()

	err = MaintainLinkerdPKI(ctx, failing, ns)
	if err == nil || !strings.Contains(err.Error(), "sync identity issuer") {
		t.Fatalf("expected sync identity issuer error, got %v", err)
	}

	healing := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
	if err := MaintainLinkerdPKI(ctx, healing, ns); err != nil {
		t.Fatalf("heal after failed sync: %v", err)
	}
	var gotIdentity corev1.Secret
	if err := healing.Get(ctx, types.NamespacedName{Namespace: ns, Name: identityIssuerSecret}, &gotIdentity); err != nil {
		t.Fatalf("get identity issuer: %v", err)
	}
	if string(gotIdentity.Data[identityIssuerCrtKey]) != string(vault.IssuerCrt) {
		t.Fatal("expected identity issuer to converge to vault on retry")
	}
}

func TestMaintainLinkerdPKI_RolloutFailureRetried(t *testing.T) {
	ctx := context.Background()
	ns := DefaultLinkerdNamespace
	vault, err := testCAAndIssuerMaterial(time.Now().UTC().Add(200 * 24 * time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	stale, err := testCAAndIssuerMaterial(time.Now().UTC().Add(190 * 24 * time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	scheme := testScheme(t)
	objs := []client.Object{
		testPKISecret(ns, vault),
		testIdentityIssuerSecret(ns, stale, false),
		testIdentityDeployment(ns, issuerFingerprint(stale.IssuerCrt), "old"),
	}
	failing := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).
		WithInterceptorFuncs(interceptor.Funcs{
			Update: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
				if dep, ok := obj.(*appsv1.Deployment); ok && dep.Name == identityDeployment {
					return errors.New("simulated identity deployment update failure")
				}
				return c.Update(ctx, obj, opts...)
			},
		}).Build()

	err = MaintainLinkerdPKI(ctx, failing, ns)
	if err == nil || !strings.Contains(err.Error(), "rollout linkerd-identity") {
		t.Fatalf("expected rollout error, got %v", err)
	}

	// Identity sync succeeded on the failing client; fingerprint did not.
	var syncedIdentity corev1.Secret
	if err := failing.Get(ctx, types.NamespacedName{Namespace: ns, Name: identityIssuerSecret}, &syncedIdentity); err != nil {
		t.Fatalf("get identity after partial success: %v", err)
	}
	if string(syncedIdentity.Data[identityIssuerCrtKey]) != string(vault.IssuerCrt) {
		t.Fatal("expected identity issuer synced before rollout failure")
	}
	var stuckDep appsv1.Deployment
	if err := failing.Get(ctx, types.NamespacedName{Namespace: ns, Name: identityDeployment}, &stuckDep); err != nil {
		t.Fatalf("get deployment after partial success: %v", err)
	}
	if stuckDep.Annotations[identityIssuerFingerprintAnnotation] == issuerFingerprint(vault.IssuerCrt) {
		t.Fatal("fingerprint must not persist when deployment update failed")
	}

	// Rebuild from post-sync state and confirm rollout retries.
	retryObjs := []client.Object{
		testPKISecret(ns, vault),
		&syncedIdentity,
		&stuckDep,
	}
	healing := fake.NewClientBuilder().WithScheme(scheme).WithObjects(retryObjs...).Build()
	if err := MaintainLinkerdPKI(ctx, healing, ns); err != nil {
		t.Fatalf("retry rollout: %v", err)
	}
	var gotDep appsv1.Deployment
	if err := healing.Get(ctx, types.NamespacedName{Namespace: ns, Name: identityDeployment}, &gotDep); err != nil {
		t.Fatalf("get deployment after heal: %v", err)
	}
	if gotDep.Annotations[identityIssuerFingerprintAnnotation] != issuerFingerprint(vault.IssuerCrt) {
		t.Fatal("expected fingerprint set on retry")
	}
	if gotDep.Spec.Template.Annotations[restartedAtAnnotation] == "old" ||
		gotDep.Spec.Template.Annotations[restartedAtAnnotation] == "" {
		t.Fatal("expected restartedAt updated on rollout retry")
	}
}

func testScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := appsv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	return scheme
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

func testIdentityIssuerSecret(ns string, mat *Material, withHelmKeep bool) *corev1.Secret {
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: identityIssuerSecret, Namespace: ns},
		Type:       corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			identityIssuerCrtKey: mat.IssuerCrt,
			identityIssuerKeyKey: mat.IssuerKey,
		},
	}
	if withHelmKeep {
		sec.Annotations = map[string]string{
			helmResourcePolicyKeep: helmResourcePolicyKeepValue,
		}
	}
	return sec
}

func testIdentityDeployment(ns, fingerprint, restartedAt string) *appsv1.Deployment {
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      identityDeployment,
			Namespace: ns,
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{}},
			},
		},
	}
	if fingerprint != "" {
		dep.Annotations = map[string]string{
			identityIssuerFingerprintAnnotation: fingerprint,
		}
	}
	if restartedAt != "" {
		dep.Spec.Template.Annotations[restartedAtAnnotation] = restartedAt
	}
	return dep
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
