package terminus

import (
	"context"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestWaitSyncLinkerdIdentitySecrets_issuerDelayedThenPresent(t *testing.T) {
	mat, err := generateInitialLinkerdPKIMaterial()
	if err != nil {
		t.Fatalf("generate pki material: %v", err)
	}

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	c := fake.NewClientBuilder().WithScheme(scheme).Build()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const ns = "linkerd"
	done := make(chan error, 1)
	go func() {
		done <- waitSyncLinkerdIdentitySecrets(ctx, c, ns, mat, 2*time.Second, 20*time.Millisecond)
	}()

	time.Sleep(30 * time.Millisecond)
	if err := c.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: linkerdIdentityIssuerSecret, Namespace: ns},
	}); err != nil {
		t.Fatalf("create issuer secret: %v", err)
	}
	if err := c.Create(ctx, &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: linkerdIdentityTrustRootsCM, Namespace: ns},
	}); err != nil {
		t.Fatalf("create trust roots configmap: %v", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected success after identity resources created, got %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for waitSyncLinkerdIdentitySecrets")
	}

	var issuer corev1.Secret
	if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: linkerdIdentityIssuerSecret}, &issuer); err != nil {
		t.Fatalf("get issuer secret: %v", err)
	}
	if string(issuer.Data[linkerdIdentityIssuerCrtKey]) != string(mat.IssuerCrt) {
		t.Fatal("issuer secret was not patched with olares pki material")
	}

	var trust corev1.ConfigMap
	if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: linkerdIdentityTrustRootsCM}, &trust); err != nil {
		t.Fatalf("get trust roots configmap: %v", err)
	}
	if trust.Data[linkerdIdentityTrustRootsKey] != string(mat.CACrt) {
		t.Fatal("trust roots configmap was not patched with olares ca")
	}
}

func TestWaitSyncLinkerdIdentitySecrets_timeoutWhenIssuerMissing(t *testing.T) {
	mat, err := generateInitialLinkerdPKIMaterial()
	if err != nil {
		t.Fatalf("generate pki material: %v", err)
	}

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	c := fake.NewClientBuilder().WithScheme(scheme).Build()
	ctx := context.Background()

	err = waitSyncLinkerdIdentitySecrets(ctx, c, "linkerd", mat, 50*time.Millisecond, 10*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	got := err.Error()
	if !strings.Contains(got, "sync linkerd identity secrets") || !strings.Contains(got, linkerdIdentityIssuerSecret) {
		t.Fatalf("unexpected error: %v", err)
	}
}
