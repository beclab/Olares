package terminus

import (
	"context"
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
	done := make(chan struct {
		changed bool
		err     error
	}, 1)
	go func() {
		changed, err := waitSyncLinkerdIdentitySecrets(ctx, c, ns, mat, 2*time.Second, 20*time.Millisecond)
		done <- struct {
			changed bool
			err     error
		}{changed, err}
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
	case result := <-done:
		if result.err != nil {
			t.Fatalf("expected success after identity resources created, got %v", result.err)
		}
		if !result.changed {
			t.Fatal("expected pki sync to report changed material")
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

func TestWaitSyncLinkerdIdentitySecrets_noChangeWhenAlreadySynced(t *testing.T) {
	mat, err := generateInitialLinkerdPKIMaterial()
	if err != nil {
		t.Fatalf("generate pki material: %v", err)
	}

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	const ns = "linkerd"
	c := fake.NewClientBuilder().WithScheme(scheme).
		WithObjects(
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: linkerdIdentityIssuerSecret, Namespace: ns},
				Data: map[string][]byte{
					linkerdIdentityIssuerCrtKey: mat.IssuerCrt,
					linkerdIdentityIssuerKeyKey: mat.IssuerKey,
				},
			},
			&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: linkerdIdentityTrustRootsCM, Namespace: ns},
				Data: map[string]string{
					linkerdIdentityTrustRootsKey: string(mat.CACrt),
				},
			},
		).Build()
	ctx := context.Background()

	changed, err := waitSyncLinkerdIdentitySecrets(ctx, c, ns, mat, time.Second, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if changed {
		t.Fatal("expected no change when issuer and trust roots already match olares pki")
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

	_, err = waitSyncLinkerdIdentitySecrets(ctx, c, "linkerd", mat, 50*time.Millisecond, 10*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	got := err.Error()
	if !strings.Contains(got, "sync linkerd identity secrets") || !strings.Contains(got, linkerdIdentityIssuerSecret) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRestartLinkerdControlPlaneAfterPKISync_restartsAllDeployments(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)
	const ns = "linkerd"
	c := fake.NewClientBuilder().WithScheme(scheme).Build()
	ctx := context.Background()

	for _, name := range linkerdControlPlaneDeployments {
		if err := c.Create(ctx, &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		}); err != nil {
			t.Fatalf("create deployment %s: %v", name, err)
		}
	}

	if err := restartLinkerdControlPlaneAfterPKISync(ctx, c, ns); err != nil {
		t.Fatalf("restart control plane: %v", err)
	}

	var restartedAt string
	for _, name := range linkerdControlPlaneDeployments {
		var dep appsv1.Deployment
		if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: name}, &dep); err != nil {
			t.Fatalf("get deployment %s: %v", name, err)
		}
		got := dep.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"]
		if got == "" {
			t.Fatalf("deployment %s missing restartedAt annotation", name)
		}
		if restartedAt == "" {
			restartedAt = got
			continue
		}
		if got != restartedAt {
			t.Fatalf("deployment %s restartedAt %q != %q", name, got, restartedAt)
		}
	}
}
